package wifi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/google/uuid"
	"github.com/nxo/engine/internal/database"
	"github.com/nxo/engine/internal/plugins"
	"gorm.io/gorm"
)

// WiFiPlugin expose les endpoints de pilotage Wi-Fi.
type WiFiPlugin struct {
	startedAt      time.Time
	mu             sync.Mutex
	db             *database.DB // Database reference
	lastScan       []WiFiNetwork
	lastAt         time.Time
	captureRunning bool
	captureFile    string
	captureTargets []string
	captureCmd     *exec.Cmd
	captureCancel  context.CancelFunc
	captureID      uuid.UUID // Current capture DB ID
	captureStart   time.Time // Capture start time
	// Bruteforce state
	bruteRunning   bool
	bruteCmd       *exec.Cmd
	bruteCancel    context.CancelFunc
	bruteCapture   string
	bruteStartedAt time.Time
	bruteResult    *BruteforceResult
	// Deauth state
	deauthRunning bool
	deauthCmd     *exec.Cmd
	deauthCancel  context.CancelFunc
	deauthTarget  string
	deauthBSSID   string
	deauthCount   int
	deauthSent    int
}

// BruteforceResult holds the result of a bruteforce attack
type BruteforceResult struct {
	Success   bool      `json:"success"`
	Password  string    `json:"password,omitempty"`
	SSID      string    `json:"ssid"`
	BSSID     string    `json:"bssid"`
	Capture   string    `json:"capture"`
	Wordlist  string    `json:"wordlist"`
	Duration  float64   `json:"duration_seconds"`
	TestedAt  time.Time `json:"tested_at"`
	KeysTotal int64     `json:"keys_total,omitempty"`
}

// WiFiNetwork describes a detected network during scan
type WiFiNetwork struct {
	SSID      string    `json:"ssid"`
	BSSID     string    `json:"bssid"`
	Channel   int       `json:"channel"`
	Signal    int       `json:"signal"`
	Security  string    `json:"security"`
	Vendor    string    `json:"vendor,omitempty"`
	WPS       string    `json:"wps,omitempty"`
	Interface string    `json:"interface"`
	LastSeen  time.Time `json:"last_seen"`
}

func (p *WiFiPlugin) Key() string         { return "wifi" }
func (p *WiFiPlugin) Version() string     { return "0.1.0" }
func (p *WiFiPlugin) Description() string { return "Pentest Wi-Fi (scan, capture, bruteforce)" }

// Models implements PluginWithModels interface.
// Returns all GORM models that need to be auto-migrated for this plugin.
func (p *WiFiPlugin) Models() []interface{} {
	return []interface{}{
		&WifiCapture{},
		&WifiNetwork{},
		&WifiBruteforceResult{},
		&WifiDeauthLog{},
		&WifiAudit{},
		&WifiAuditNetwork{},
	}
}

func (p *WiFiPlugin) Manifest() map[string]any {
	return map[string]any{
		"name":        "Wi-Fi",
		"version":     p.Version(),
		"description": p.Description(),
		"routes":      []map[string]any{},
		"permissions": []string{"wifi:scan", "wifi:capture", "wifi:crack"},
		"menu_items": []map[string]any{
			{
				"label":    "Wi-Fi Pentest",
				"path":     "/admin/plugins/wifi",
				"icon":     "wifi",
				"position": 120,
			},
		},
	}
}

// RegisterRoutes expose des handlers minimalistes; la logique CLI sera branchée ultérieurement.
func (p *WiFiPlugin) RegisterRoutes(group *buffalo.App, deps plugins.Deps) {
	// Store DB reference for use in handlers
	p.db = deps.DB

	// UI schema - construit dynamiquement via ViewBuilder
	group.GET("/view", func(c buffalo.Context) error {
		// Récupérer les données actuelles
		interfaces, _ := getWiFiInterfaces()

		p.mu.Lock()
		networks := append([]WiFiNetwork(nil), p.lastScan...)
		captureRunning := p.captureRunning
		bruteRunning := p.bruteRunning
		p.mu.Unlock()

		// Récupérer les captures depuis la BDD
		viewBuilder := NewViewBuilder(p.db)
		captures := viewBuilder.GetCapturesForView()

		// Récupérer les wordlists
		wordlists := getWordlistsForView()

		// Construire la vue complète
		view := viewBuilder.BuildMainView(
			interfaces,
			networks,
			captures,
			wordlists,
			captureRunning,
			bruteRunning,
		)

		return writeJSON(c, 200, view)
	})

	// Lister les vraies interfaces WiFi du système
	group.GET("/interfaces", func(c buffalo.Context) error {
		interfaces, err := getWiFiInterfaces()
		if err != nil {
			return writeJSON(c, 500, map[string]any{"error": err.Error()})
		}
		return writeJSON(c, 200, map[string]any{"interfaces": interfaces})
	})

	// Lancer un scan (synchrone pour l'instant)
	group.POST("/scan", func(c buffalo.Context) error {
		var req struct {
			Interface string `json:"interface"`
		}
		_ = json.NewDecoder(c.Request().Body).Decode(&req)

		iface := strings.TrimSpace(req.Interface)
		if iface == "" {
			// Choisir la première interface détectée
			ifaces, _ := getWiFiInterfaces()
			if len(ifaces) == 0 {
				return writeJSON(c, 400, map[string]any{"error": "Aucune interface Wi-Fi détectée"})
			}
			iface, _ = ifaces[0]["name"].(string)
		}

		       fmt.Printf("[WIFI-SCAN] Lancement du scan sur l'interface: %s\n", iface)
		       networks, err := scanWiFi(iface)
		       if err != nil {
			       fmt.Printf("[WIFI-SCAN] Erreur lors du scan: %v\n", err)
			       return writeJSON(c, 500, map[string]any{"error": err.Error()})
		       }

		       fmt.Printf("[WIFI-SCAN] %d réseaux détectés\n", len(networks))
		       for _, n := range networks {
			       fmt.Printf("[WIFI-SCAN] Réseau: SSID=%s BSSID=%s Channel=%d Signal=%d\n", n.SSID, n.BSSID, n.Channel, n.Signal)
		       }

		       // Sauvegarde en mémoire
		       p.mu.Lock()
		       p.lastScan = networks
		       p.lastAt = time.Now()
		       p.mu.Unlock()

		       // Sauvegarde en base (table wifi_networks)
		       if p.db != nil {
			       for _, n := range networks {
				       // Upsert par BSSID + Channel
				       var existing WifiNetwork
				       err := p.db.Where("bssid = ? AND channel = ?", n.BSSID, n.Channel).First(&existing).Error
				       if err == gorm.ErrRecordNotFound {
					       _ = p.db.Create(&n)
				       } else if err == nil {
					       _ = p.db.Model(&existing).Updates(n)
				       }
			       }
		       }

		       return writeJSON(c, 200, map[string]any{
			       "status":     "ok",
			       "count":      len(networks),
			       "scanned_at": p.lastAt,
			       "results":    networks,
		       })
	})

	// Récupérer les derniers résultats de scan
	group.GET("/scan/results", func(c buffalo.Context) error {
		p.mu.Lock()
		networks := append([]WiFiNetwork(nil), p.lastScan...)
		scannedAt := p.lastAt
		p.mu.Unlock()

		// Enrichir les résultats avec les actions (comme dans le ViewBuilder)
		results := make([]map[string]any, len(networks))
		for i, n := range networks {
			results[i] = map[string]any{
				"ssid":     n.SSID,
				"bssid":    n.BSSID,
				"channel":  n.Channel,
				"signal":   n.Signal,
				"security": n.Security,
				"vendor":   n.Vendor,
				"wps":      n.WPS,
				"actions": map[string]any{
					"type": "rowActions",
					"items": []map[string]any{
						{
							"id":       "deauth",
							"label":    "Deauth",
							"icon":     "zap",
							"variant":  "danger",
							"endpoint": "/plugins/wifi/deauth",
							"method":   "POST",
							"data": map[string]any{
								"bssid":   n.BSSID,
								"channel": n.Channel,
							},
							"confirm": fmt.Sprintf("Lancer une attaque deauth sur %s ?", n.SSID),
						},
					},
				},
			}
		}

		return writeJSON(c, 200, map[string]any{
			"results":    results,
			"count":      len(results),
			"scanned_at": scannedAt,
		})
	})

	// Capture de handshakes
	group.POST("/capture", func(c buffalo.Context) error {
		var req struct {
			Interface string   `json:"interface"`
			Targets   []string `json:"targets"` // BSSIDs
			Channel   int      `json:"channel"`
			Mode      string   `json:"mode"`     // wpa, wep
			Duration  int      `json:"duration"` // seconds, 0 = until stopped
		}
		_ = json.NewDecoder(c.Request().Body).Decode(&req)

		iface := strings.TrimSpace(req.Interface)
		if iface == "" {
			return writeJSON(c, 400, map[string]any{"error": "Interface manquante"})
		}
		if len(req.Targets) == 0 {
			return writeJSON(c, 400, map[string]any{"error": "Aucun BSSID cible"})
		}

		// Récupérer les infos SSID/Security depuis le dernier scan
		var targetSSID, targetSecurity string
		var targetChannel int
		p.mu.Lock()
		for _, net := range p.lastScan {
			for _, target := range req.Targets {
				if strings.EqualFold(net.BSSID, target) {
					if targetSSID == "" {
						targetSSID = net.SSID
						targetSecurity = net.Security
						targetChannel = net.Channel
					}
					break
				}
			}
		}
		p.mu.Unlock()

		// Activer le mode monitor sur l'interface
		// L'interface garde le même nom (iw ne crée pas de nouvelle interface)
		monitorIface := iface
		if err := enableMonitorMode(iface); err != nil {
			return writeJSON(c, 500, map[string]any{
				"error": fmt.Sprintf("Impossible d'activer le mode monitor: %v", err),
				"hint":  "Essayez: sudo airmon-ng start " + iface,
			})
		}

		// Vérifier si airmon-ng a créé une interface avec suffixe "mon"
		// (sur certains systèmes, airmon-ng crée wlan0mon au lieu de modifier wlan0)
		possibleMonIface := iface + "mon"
		if ifaceExists(possibleMonIface) {
			monitorIface = possibleMonIface
		}

		// Trouver le canal des cibles
		channel := req.Channel
		if channel == 0 {
			channel = targetChannel
		}

		// Créer le répertoire de capture
		captureDir := "/opt/heimdall/captures"
		if err := os.MkdirAll(captureDir, 0755); err != nil {
			return writeJSON(c, 500, map[string]any{"error": fmt.Sprintf("Impossible de créer le répertoire de capture: %v", err)})
		}
		timestamp := time.Now().Format("20060102_150405")
		captureName := fmt.Sprintf("capture_%s", timestamp)
		captureFile := fmt.Sprintf("%s/%s", captureDir, captureName)

		duration := req.Duration
		// Si duration < 0 (non spécifié), utiliser 5 minutes par défaut
		// Si duration == 0, c'est illimité (arrêt manuel uniquement)
		if duration < 0 {
			duration = 300 // 5 minutes par défaut (recommandé WPA2)
		}

		// Créer l'enregistrement en base de données
		startTime := time.Now()
		capture := WifiCapture{
			SSID:            targetSSID,
			BSSID:           strings.Join(req.Targets, ","),
			Channel:         channel,
			Security:        targetSecurity,
			CapturePath:     captureFile,
			CaptureName:     captureName,
			InterfaceUsed:   monitorIface,
			DurationSeconds: duration,
			StartedAt:       &startTime,
			Status:          "running",
		}
		if p.db != nil {
			if err := p.db.Create(&capture).Error; err != nil {
				fmt.Printf("[WIFI] Erreur création capture en BDD: %v\n", err)
			}
		}

		// Créer le contexte annulable
		var ctx context.Context
		var cancel context.CancelFunc
		if duration == 0 {
			// Durée illimitée - juste un contexte annulable
			ctx, cancel = context.WithCancel(context.Background())
		} else {
			// Durée limitée
			ctx, cancel = context.WithTimeout(context.Background(), time.Duration(duration)*time.Second)
		}

		// Stocker le cancel pour pouvoir arrêter la capture
		p.mu.Lock()
		p.captureCancel = cancel
		p.captureID = capture.ID
		p.captureStart = startTime
		p.mu.Unlock()

		// Lancer airodump-ng en background
		go func() {
			defer cancel()

			p.mu.Lock()
			p.captureRunning = true
			p.captureFile = captureFile
			p.captureTargets = req.Targets
			captureDBID := p.captureID
			p.mu.Unlock()

			args := []string{
				"-w", captureFile,
				"--bssid", strings.Join(req.Targets, ","),
				"--write-interval", "1",
			}
			if channel > 0 {
				args = append(args, "-c", strconv.Itoa(channel))
			}
			args = append(args, monitorIface)

			cmd := exec.CommandContext(ctx, "/usr/sbin/airodump-ng", args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			// Créer un process group pour pouvoir tuer tous les sous-processus
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

			// Stocker la commande pour pouvoir la tuer
			p.mu.Lock()
			p.captureCmd = cmd
			p.mu.Unlock()

			err := cmd.Run()
			if err != nil && ctx.Err() == nil {
				fmt.Printf("airodump-ng error: %v\n", err)
			}

			// Mettre à jour la capture en BDD
			endTime := time.Now()
			capFiles, _ := filepath.Glob(captureFile + "*.cap")
			var fileSize int64
			hasHandshake := false

			if len(capFiles) > 0 {
				if info, err := os.Stat(capFiles[0]); err == nil {
					fileSize = info.Size()
				}
				// Vérifier le handshake
				out, _ := exec.Command("/usr/bin/aircrack-ng", capFiles[0]).Output()
				hasHandshake = strings.Contains(string(out), "1 handshake")
			}

			if p.db != nil {
				p.db.Model(&WifiCapture{}).Where("id = ?", captureDBID).Updates(map[string]any{
					"ended_at":      endTime,
					"status":        "completed",
					"file_size":     fileSize,
					"has_handshake": hasHandshake,
				})
			}

			p.mu.Lock()
			p.captureRunning = false
			p.captureCmd = nil
			p.captureCancel = nil
			p.mu.Unlock()
		}()

		return writeJSON(c, 200, map[string]any{
			"status":     "started",
			"message":    fmt.Sprintf("Capture démarrée sur %d cible(s)", len(req.Targets)),
			"file":       captureFile,
			"targets":    req.Targets,
			"channel":    channel,
			"interface":  monitorIface,
			"ssid":       targetSSID,
			"capture_id": capture.ID,
		})
	})

	// Statut de la capture
	group.GET("/capture/status", func(c buffalo.Context) error {
		p.mu.Lock()
		running := p.captureRunning
		file := p.captureFile
		targets := p.captureTargets
		p.mu.Unlock()

		// Vérifier si un fichier .cap existe
		capFiles, _ := filepath.Glob(file + "*.cap")
		hasHandshake := false
		var capSize int64

		if len(capFiles) > 0 {
			if info, err := os.Stat(capFiles[0]); err == nil {
				capSize = info.Size()
				// Vérifier le handshake avec aircrack-ng
				out, _ := exec.Command("/usr/bin/aircrack-ng", capFiles[0]).Output()
				hasHandshake = strings.Contains(string(out), "1 handshake")
			}
		}

		return writeJSON(c, 200, map[string]any{
			"running":       running,
			"file":          file,
			"targets":       targets,
			"cap_files":     capFiles,
			"cap_size":      capSize,
			"has_handshake": hasHandshake,
		})
	})

	// Arrêter la capture
	group.POST("/capture/stop", func(c buffalo.Context) error {
		fmt.Println("========================================")
		fmt.Println("[WIFI-STOP] === REQUÊTE STOP REÇUE ===")
		fmt.Printf("[WIFI-STOP] Timestamp: %s\n", time.Now().Format(time.RFC3339))
		fmt.Printf("[WIFI-STOP] Remote IP: %s\n", c.Request().RemoteAddr)
		fmt.Println("========================================")

		p.mu.Lock()
		file := p.captureFile
		wasRunning := p.captureRunning
		fmt.Printf("[WIFI-STOP] Capture était en cours: %v\n", wasRunning)
		fmt.Printf("[WIFI-STOP] Fichier capture: %s\n", file)
		p.captureRunning = false
		p.captureCmd = nil
		p.captureCancel = nil
		p.mu.Unlock()

		// 1. Tuer airodump-ng avec pkill
		fmt.Println("[WIFI-STOP] Étape 1: pkill airodump-ng...")
		pkillOut, pkillErr := exec.Command("pkill", "-9", "-f", "airodump-ng").CombinedOutput()
		fmt.Printf("[WIFI-STOP] pkill result: output=%s, err=%v\n", string(pkillOut), pkillErr)

		// 2. Tuer aireplay-ng aussi
		fmt.Println("[WIFI-STOP] Étape 2: pkill aireplay-ng...")
		exec.Command("pkill", "-9", "-f", "aireplay-ng").Run()

		// 3. Attendre un peu
		fmt.Println("[WIFI-STOP] Étape 3: Attente 500ms...")
		time.Sleep(500 * time.Millisecond)

		// 4. Redémarrer NetworkManager EN BACKGROUND après la réponse
		// (sinon le restart coupe la connexion HTTP avant qu'on puisse répondre)
		fmt.Println("[WIFI-STOP] Étape 4: Planification redémarrage NetworkManager...")
		go func() {
			time.Sleep(100 * time.Millisecond) // Laisser le temps à la réponse HTTP de partir
			fmt.Println("[WIFI-STOP] Exécution: systemctl restart NetworkManager...")
			nmCmd := exec.Command("/usr/bin/systemctl", "restart", "NetworkManager")
			nmOut, nmErr := nmCmd.CombinedOutput()
			fmt.Printf("[WIFI-STOP] NetworkManager result: output=%s, err=%v\n", string(nmOut), nmErr)
			fmt.Println("[WIFI-STOP] === REDÉMARRAGE TERMINÉ ===")
		}()

		// 5. Log final (avant le restart pour que la réponse HTTP parte)
		fmt.Println("========================================")
		fmt.Println("[WIFI-STOP] === STOP TERMINÉ (NetworkManager restart planifié) ===")
		fmt.Println("========================================")

		return writeJSON(c, 200, map[string]any{
			"status":      "stopped",
			"message":     "Capture arrêtée, WiFi reconnecté",
			"file":        file,
			"was_running": wasRunning,
			"pkill_error": fmt.Sprintf("%v", pkillErr),
			"nm_error":    "scheduled",
		})
	})

	// Reconnecter le WiFi (désactiver mode monitor)
	group.POST("/reconnect", func(c buffalo.Context) error {
		fmt.Println("[WIFI-RECONNECT] === REQUÊTE RECONNECT REÇUE ===")

		var req struct {
			Interface string `json:"interface"`
		}
		_ = json.NewDecoder(c.Request().Body).Decode(&req)

		iface := strings.TrimSpace(req.Interface)
		fmt.Printf("[WIFI-RECONNECT] Interface demandée: %s\n", iface)

		// Arrêter tous les processus de capture (même sans interface spécifiée)
		fmt.Println("[WIFI-RECONNECT] Arrêt des processus airodump/aireplay...")
		_ = exec.Command("pkill", "-9", "-f", "airodump-ng").Run()
		_ = exec.Command("pkill", "-9", "-f", "aireplay-ng").Run()

		p.mu.Lock()
		p.captureRunning = false
		p.mu.Unlock()

		// Si une interface est spécifiée, la remettre en mode managed
		if iface != "" {
			fmt.Printf("[WIFI-RECONNECT] Remise en mode managed de %s...\n", iface)
			// Utiliser iw directement (plus fiable que disableMonitorMode)
			ipPath := findTool("ip")
			iwPath := findTool("iw")

			exec.Command(ipPath, "link", "set", iface, "down").Run()
			time.Sleep(100 * time.Millisecond)
			exec.Command(iwPath, "dev", iface, "set", "type", "managed").Run()
			exec.Command(ipPath, "link", "set", iface, "up").Run()
		}

		// Planifier le restart NetworkManager EN BACKGROUND
		// pour que la réponse HTTP parte AVANT le restart
		fmt.Println("[WIFI-RECONNECT] Planification restart NetworkManager en background...")
		go func() {
			time.Sleep(100 * time.Millisecond) // Laisser partir la réponse HTTP
			fmt.Println("[WIFI-RECONNECT] Exécution: systemctl restart NetworkManager...")
			cmd := exec.Command("/usr/bin/systemctl", "restart", "NetworkManager")
			out, err := cmd.CombinedOutput()
			fmt.Printf("[WIFI-RECONNECT] NetworkManager result: %s, err=%v\n", string(out), err)
			fmt.Println("[WIFI-RECONNECT] === RECONNEXION TERMINÉE ===")
		}()

		fmt.Println("[WIFI-RECONNECT] Envoi réponse HTTP (NM restart planifié)...")
		return writeJSON(c, 200, map[string]any{
			"status":    "reconnecting",
			"message":   "Reconnexion WiFi en cours...",
			"interface": iface,
		})
	})

	// ========== DEAUTH ATTACK ==========

	// Lancer une attaque de deauth pour forcer la reconnexion d'un client
	group.POST("/deauth", func(c buffalo.Context) error {
		var req struct {
			Interface string `json:"interface"`
			BSSID     string `json:"bssid"`   // AP cible
			Client    string `json:"client"`  // Client MAC (optionnel, FF:FF:FF:FF:FF:FF = broadcast)
			Count     int    `json:"count"`   // Nombre de paquets (0 = continu)
			Channel   int    `json:"channel"` // Canal du réseau
		}
		_ = json.NewDecoder(c.Request().Body).Decode(&req)

		iface := strings.TrimSpace(req.Interface)
		bssid := strings.TrimSpace(req.BSSID)

		if iface == "" {
			return writeJSON(c, 400, map[string]any{"error": "Interface manquante"})
		}
		if bssid == "" {
			return writeJSON(c, 400, map[string]any{"error": "BSSID cible manquant"})
		}

		// Client par défaut = broadcast (tous les clients)
		client := strings.TrimSpace(req.Client)
		if client == "" {
			client = "FF:FF:FF:FF:FF:FF"
		}

		// Nombre de paquets par défaut
		count := req.Count
		if count <= 0 {
			count = 10 // 10 paquets par défaut
		}

		// Vérifier si un deauth est déjà en cours
		p.mu.Lock()
		if p.deauthRunning {
			p.mu.Unlock()
			return writeJSON(c, 409, map[string]any{
				"error":  "Un deauth est déjà en cours",
				"target": p.deauthBSSID,
			})
		}
		p.deauthRunning = true
		p.deauthBSSID = bssid
		p.deauthTarget = client
		p.deauthCount = count
		p.deauthSent = 0
		p.mu.Unlock()

		// S'assurer que l'interface est en mode monitor
		monitorIface := iface
		if !strings.HasSuffix(iface, "mon") {
			// Vérifier si déjà en monitor
			out, _ := exec.Command("iwconfig", iface).CombinedOutput()
			if !strings.Contains(string(out), "Mode:Monitor") {
				// Activer le mode monitor
				if err := enableMonitorMode(iface); err != nil {
					p.mu.Lock()
					p.deauthRunning = false
					p.mu.Unlock()
					return writeJSON(c, 500, map[string]any{
						"error": fmt.Sprintf("Impossible d'activer le mode monitor: %v", err),
					})
				}
			}
			// Vérifier si airmon-ng a créé une interface "mon"
			if ifaceExists(iface + "mon") {
				monitorIface = iface + "mon"
			}
		}

		// Changer le canal si spécifié
		if req.Channel > 0 {
			iwPath := findTool("iw")
			exec.Command(iwPath, "dev", monitorIface, "set", "channel", strconv.Itoa(req.Channel)).Run()
			time.Sleep(100 * time.Millisecond)
		}

		// Créer le contexte annulable
		ctx, cancel := context.WithCancel(context.Background())
		p.mu.Lock()
		p.deauthCancel = cancel
		p.mu.Unlock()

		// Lancer aireplay-ng en background
		go func() {
			defer func() {
				p.mu.Lock()
				p.deauthRunning = false
				p.deauthCmd = nil
				p.deauthCancel = nil
				p.mu.Unlock()
			}()

			// aireplay-ng --deauth COUNT -a BSSID -c CLIENT INTERFACE
			args := []string{
				"--deauth", strconv.Itoa(count),
				"-a", bssid,
			}
			if client != "FF:FF:FF:FF:FF:FF" {
				args = append(args, "-c", client)
			}
			args = append(args, monitorIface)

			fmt.Printf("[WIFI-DEAUTH] Lancement: aireplay-ng %v\n", args)

			cmd := exec.CommandContext(ctx, "/usr/sbin/aireplay-ng", args...)
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

			p.mu.Lock()
			p.deauthCmd = cmd
			p.mu.Unlock()

			output, err := cmd.CombinedOutput()
			if err != nil && ctx.Err() == nil {
				fmt.Printf("[WIFI-DEAUTH] Erreur aireplay-ng: %v\n", err)
			}

			// Compter les paquets envoyés
			lines := strings.Split(string(output), "\n")
			sent := 0
			for _, line := range lines {
				if strings.Contains(line, "Sending DeAuth") || strings.Contains(line, "deauthentication") {
					sent++
				}
			}

			p.mu.Lock()
			p.deauthSent = sent
			p.mu.Unlock()

			fmt.Printf("[WIFI-DEAUTH] Terminé: %d paquets envoyés vers %s\n", sent, bssid)
		}()

		return writeJSON(c, 200, map[string]any{
			"status":    "started",
			"message":   fmt.Sprintf("Deauth lancé: %d paquets vers %s", count, bssid),
			"bssid":     bssid,
			"client":    client,
			"count":     count,
			"interface": monitorIface,
		})
	})

	// Statut du deauth
	group.GET("/deauth/status", func(c buffalo.Context) error {
		p.mu.Lock()
		running := p.deauthRunning
		bssid := p.deauthBSSID
		target := p.deauthTarget
		count := p.deauthCount
		sent := p.deauthSent
		p.mu.Unlock()

		return writeJSON(c, 200, map[string]any{
			"running": running,
			"bssid":   bssid,
			"client":  target,
			"count":   count,
			"sent":    sent,
		})
	})

	// Arrêter le deauth
	group.POST("/deauth/stop", func(c buffalo.Context) error {
		p.mu.Lock()
		wasRunning := p.deauthRunning
		if p.deauthCancel != nil {
			p.deauthCancel()
		}
		p.deauthRunning = false
		p.mu.Unlock()

		// Tuer aireplay-ng
		_ = exec.Command("pkill", "-9", "-f", "aireplay-ng").Run()

		return writeJSON(c, 200, map[string]any{
			"status":      "stopped",
			"was_running": wasRunning,
		})
	})

	// Lister les captures depuis la BDD
	group.GET("/captures", func(c buffalo.Context) error {
		var captures []WifiCapture

		if p.db != nil {
			// Récupérer les captures depuis la BDD (les plus récentes d'abord)
			p.db.Order("created_at DESC").Find(&captures)
		}

		// Enrichir avec les infos fichier actuelles
		var result []map[string]any
		for _, cap := range captures {
			capInfo := map[string]any{
				"id":            cap.ID,
				"ssid":          cap.SSID,
				"bssid":         cap.BSSID,
				"channel":       cap.Channel,
				"security":      cap.Security,
				"path":          cap.CapturePath,
				"name":          cap.CaptureName,
				"interface":     cap.InterfaceUsed,
				"duration":      cap.DurationSeconds,
				"started_at":    cap.StartedAt,
				"ended_at":      cap.EndedAt,
				"status":        cap.Status,
				"cracked":       cap.Cracked,
				"password":      cap.CrackedPassword,
				"has_handshake": cap.HasHandshake,
				"size":          cap.FileSize,
			}

			// Vérifier le fichier .cap actuel
			capFiles, _ := filepath.Glob(cap.CapturePath + "*.cap")
			if len(capFiles) > 0 {
				if info, err := os.Stat(capFiles[0]); err == nil {
					capInfo["size"] = info.Size()
					capInfo["modified"] = info.ModTime()
					capInfo["cap_file"] = capFiles[0]
				}
				// Revérifier handshake si pas encore marqué
				if !cap.HasHandshake {
					out, _ := exec.Command("/usr/bin/aircrack-ng", capFiles[0]).Output()
					if strings.Contains(string(out), "1 handshake") {
						capInfo["has_handshake"] = true
						// Mettre à jour en BDD
						p.db.Model(&WifiCapture{}).Where("id = ?", cap.ID).Update("has_handshake", true)
					}
				}
			}

			result = append(result, capInfo)
		}

		// Fallback: si pas de captures en BDD, scanner les fichiers
		if len(result) == 0 {
			captureDir := "/opt/heimdall/captures"
			files, _ := filepath.Glob(captureDir + "/*.cap")

			for _, f := range files {
				info, err := os.Stat(f)
				if err != nil {
					continue
				}

				out, _ := exec.Command("/usr/bin/aircrack-ng", f).Output()
				hasHandshake := strings.Contains(string(out), "handshake")

				result = append(result, map[string]any{
					"path":          f,
					"name":          filepath.Base(f),
					"size":          info.Size(),
					"modified":      info.ModTime(),
					"has_handshake": hasHandshake,
				})
			}
		}

		return writeJSON(c, 200, map[string]any{
			"captures": result,
			"count":    len(result),
		})
	})

	// Bruteforce - Lancer l'attaque
	group.POST("/bruteforce", func(c buffalo.Context) error {
		var req struct {
			CapturePath string `json:"capture_path"`
			Wordlist    string `json:"wordlist"`
			BSSID       string `json:"bssid"`
			SSID        string `json:"ssid"`
		}
		_ = json.NewDecoder(c.Request().Body).Decode(&req)

		if req.CapturePath == "" {
			return writeJSON(c, 400, map[string]any{"error": "Fichier capture manquant"})
		}
		if req.Wordlist == "" {
			req.Wordlist = "/opt/heimdall/wordlists/rockyou.txt"
		}

		// Vérifier que les fichiers existent
		if _, err := os.Stat(req.CapturePath); err != nil {
			return writeJSON(c, 400, map[string]any{"error": "Fichier capture introuvable"})
		}
		if _, err := os.Stat(req.Wordlist); err != nil {
			return writeJSON(c, 400, map[string]any{"error": "Wordlist introuvable", "path": req.Wordlist})
		}

		// Vérifier si un bruteforce est déjà en cours
		p.mu.Lock()
		if p.bruteRunning {
			p.mu.Unlock()
			return writeJSON(c, 409, map[string]any{"error": "Un bruteforce est déjà en cours", "capture": p.bruteCapture})
		}
		p.bruteRunning = true
		p.bruteCapture = req.CapturePath
		p.bruteStartedAt = time.Now()
		p.bruteResult = nil
		p.mu.Unlock()

		// Créer le contexte annulable
		ctx, cancel := context.WithCancel(context.Background())
		p.mu.Lock()
		p.bruteCancel = cancel
		p.mu.Unlock()

		// Lancer aircrack-ng en background
		go func() {
			defer func() {
				p.mu.Lock()
				p.bruteRunning = false
				p.bruteCmd = nil
				p.bruteCancel = nil
				p.mu.Unlock()
			}()

			args := []string{"-w", req.Wordlist, "-b", req.BSSID, "-l", req.CapturePath + ".key", req.CapturePath}
			cmd := exec.CommandContext(ctx, "/usr/bin/aircrack-ng", args...)

			p.mu.Lock()
			p.bruteCmd = cmd
			p.mu.Unlock()

			output, err := cmd.CombinedOutput()
			duration := time.Since(p.bruteStartedAt).Seconds()

			// Parser le résultat
			result := &BruteforceResult{
				SSID:     req.SSID,
				BSSID:    req.BSSID,
				Capture:  req.CapturePath,
				Wordlist: req.Wordlist,
				Duration: duration,
				TestedAt: time.Now(),
			}

			outputStr := string(output)

			// Chercher le mot de passe trouvé
			if strings.Contains(outputStr, "KEY FOUND!") {
				result.Success = true
				// Extraire le mot de passe entre [ ]
				if idx := strings.Index(outputStr, "KEY FOUND! [ "); idx != -1 {
					start := idx + len("KEY FOUND! [ ")
					end := strings.Index(outputStr[start:], " ]")
					if end != -1 {
						result.Password = outputStr[start : start+end]
					}
				}
				// Lire aussi le fichier .key créé par -l
				if keyData, err := os.ReadFile(req.CapturePath + ".key"); err == nil {
					result.Password = strings.TrimSpace(string(keyData))
				}
			}

			// Sauvegarder le résultat
			p.mu.Lock()
			p.bruteResult = result
			p.mu.Unlock()

			// Sauvegarder dans un fichier JSON
			resultJSON, _ := json.MarshalIndent(result, "", "  ")
			_ = os.WriteFile(req.CapturePath+".result.json", resultJSON, 0644)
			_ = os.WriteFile(req.CapturePath+".output.txt", output, 0644)

			// Mettre à jour la capture en BDD si mot de passe trouvé
			if result.Success && p.db != nil {
				now := time.Now()
				p.db.Model(&WifiCapture{}).
					Where("capture_path LIKE ?", strings.TrimSuffix(req.CapturePath, ".cap")+"%").
					Updates(map[string]any{
						"cracked":          true,
						"cracked_password": result.Password,
						"cracked_at":       now,
					})
			}

			// Sauvegarder dans la table des résultats de bruteforce (succès ou échec)
			if p.db != nil {
				now := time.Now()
				status := "completed"
				if !result.Success {
					status = "failed"
				}
				bruteResult := WifiBruteforceResult{
					SSID:         result.SSID,
					BSSID:        result.BSSID,
					CapturePath:  req.CapturePath,
					WordlistPath: req.Wordlist,
					WordlistName: filepath.Base(req.Wordlist),
					Success:      result.Success,
					Password:     result.Password,
					DurationSecs: duration,
					StartedAt:    p.bruteStartedAt,
					Status:       status,
				}
				bruteResult.CompletedAt = &now
				p.db.Create(&bruteResult)
			}

			if err != nil && ctx.Err() == nil {
				fmt.Printf("[WIFI-BRUTE] aircrack-ng error: %v\n", err)
			}
			if result.Success {
				fmt.Printf("[WIFI-BRUTE] ✅ MOT DE PASSE TROUVÉ: %s pour %s\n", result.Password, result.SSID)
			} else {
				fmt.Printf("[WIFI-BRUTE] ❌ Mot de passe non trouvé pour %s (durée: %.1fs)\n", result.SSID, duration)
			}
		}()

		return writeJSON(c, 200, map[string]any{
			"status":     "started",
			"message":    "Bruteforce démarré",
			"capture":    req.CapturePath,
			"wordlist":   req.Wordlist,
			"bssid":      req.BSSID,
			"started_at": p.bruteStartedAt,
		})
	})

	// Bruteforce Combiné - Mode intelligent avec wordlist et/ou pattern
	group.POST("/bruteforce/combined", func(c buffalo.Context) error {
		var req struct {
			CaptureID      string   `json:"capture_id"`
			CapturePath    string   `json:"capture_path"`
			BSSID          string   `json:"bssid"`
			SSID           string   `json:"ssid"`
			UseWordlist    bool     `json:"use_wordlist"`
			UsePattern     bool     `json:"use_pattern"`
			Wordlists      []string `json:"wordlists"` // Multi-wordlist
			Wordlist       string   `json:"wordlist"`  // Pour compatibilité
			CustomWordlist string   `json:"custom_wordlist"`
			ISP            string   `json:"isp"`
			CustomMask     string   `json:"custom_mask"`
			TimeoutHours   int      `json:"timeout_hours"`
		}
		_ = json.NewDecoder(c.Request().Body).Decode(&req)

		// Timeout par défaut: 2 heures
		timeoutDuration := 2 * time.Hour
		if req.TimeoutHours > 0 {
			timeoutDuration = time.Duration(req.TimeoutHours) * time.Hour
		}

		// Récupérer le chemin de capture
		capturePath := req.CapturePath
		if req.CaptureID != "" && p.db != nil {
			var capture WifiCapture
			if err := p.db.First(&capture, "id = ?", req.CaptureID).Error; err == nil {
				// Trouver le fichier .cap réel
				capFiles, _ := filepath.Glob(capture.CapturePath + "*.cap")
				if len(capFiles) > 0 {
					capturePath = capFiles[0]
				} else {
					capturePath = capture.CapturePath
				}
				if req.BSSID == "" {
					req.BSSID = capture.BSSID
				}
				if req.SSID == "" {
					req.SSID = capture.SSID
				}
			}
		}

		if capturePath == "" {
			return writeJSON(c, 400, map[string]any{"error": "Fichier capture manquant"})
		}

		// Vérifier que le fichier existe
		if _, err := os.Stat(capturePath); err != nil {
			return writeJSON(c, 400, map[string]any{"error": "Fichier capture introuvable", "path": capturePath})
		}

		// Au moins un mode doit être sélectionné
		if !req.UseWordlist && !req.UsePattern {
			return writeJSON(c, 400, map[string]any{"error": "Sélectionnez au moins un mode (Wordlist ou Pattern)"})
		}

		// Vérifier si un bruteforce est déjà en cours
		p.mu.Lock()
		if p.bruteRunning {
			p.mu.Unlock()
			return writeJSON(c, 409, map[string]any{"error": "Un bruteforce est déjà en cours", "capture": p.bruteCapture})
		}
		p.bruteRunning = true
		p.bruteCapture = capturePath
		p.bruteStartedAt = time.Now()
		p.bruteResult = nil
		p.mu.Unlock()

		// Créer le contexte avec timeout
		ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
		p.mu.Lock()
		p.bruteCancel = cancel
		p.mu.Unlock()
		fmt.Printf("[WIFI-BRUTE] ⏱️ Timeout configuré: %v\n", timeoutDuration)

		// Déterminer la wordlist
		wordlist := req.Wordlist
		if wordlist == "custom" && req.CustomWordlist != "" {
			wordlist = req.CustomWordlist
		}

		// Logique spéciale pour Freebox : utiliser la wordlist de mots latins
		isp := req.ISP
		if isp == "Free" {
			freeboxWordlist := GetFreeboxWordlistPath()
			if freeboxWordlist != "" {
				fmt.Printf("[WIFI-BRUTE] 🇫🇷 Freebox détectée! Utilisation de la wordlist latine: %s\n", freeboxWordlist)
				// Si pas de wordlist spécifiée, utiliser la wordlist Freebox
				if wordlist == "" || wordlist == "/opt/heimdall/wordlists/rockyou.txt" {
					wordlist = freeboxWordlist
				}
			}
		}

		if wordlist == "" {
			wordlist = "/opt/heimdall/wordlists/rockyou.txt"
		}

		// Déterminer le masque pour le pattern
		mask := req.CustomMask
		if mask == "" && isp != "" {
			masks := GetMasksForISP(isp)
			if len(masks) > 0 {
				mask = masks[0].Mask
				fmt.Printf("[WIFI-BRUTE] Masque trouvé pour %s: %s\n", isp, mask)
			} else {
				fmt.Printf("[WIFI-BRUTE] ⚠️ Aucun masque défini pour l'ISP '%s'\n", isp)
			}
		}

		// Modes à exécuter
		modes := []string{}
		if req.UseWordlist {
			modes = append(modes, "wordlist")
		}
		if req.UsePattern {
			if mask != "" {
				modes = append(modes, "pattern")
			} else {
				fmt.Printf("[WIFI-BRUTE] ⚠️ Pattern demandé mais aucun masque disponible (ISP: %s)\n", isp)
			}
		}

		// Lancer en background
		go func() {
			defer func() {
				p.mu.Lock()
				p.bruteRunning = false
				p.bruteCmd = nil
				p.bruteCancel = nil
				p.mu.Unlock()
			}()

			var finalResult *BruteforceResult
			startTime := time.Now()

			// Phase 1: Wordlist (aircrack-ng - plus rapide, multi)
			if req.UseWordlist {
				var wordlists []string
				if len(req.Wordlists) > 0 {
					wordlists = req.Wordlists
				} else if req.Wordlist != "" {
					wordlists = []string{req.Wordlist}
				} else {
					wordlists = []string{"/opt/heimdall/wordlists/rockyou.txt"}
				}
				for _, wl := range wordlists {
					fmt.Printf("[WIFI-BRUTE] 📖 Phase 1: Wordlist (%s)\n", wl)
					if _, err := os.Stat(wl); err == nil {
						args := []string{"-w", wl, "-b", req.BSSID, "-l", capturePath + ".key", capturePath}
						cmd := exec.CommandContext(ctx, "/usr/bin/aircrack-ng", args...)
						p.mu.Lock()
						p.bruteCmd = cmd
						p.mu.Unlock()
						output, err := cmd.CombinedOutput()
						outputStr := string(output)
						if strings.Contains(outputStr, "KEY FOUND!") {
							// Mot de passe trouvé!
							password := ""
							if idx := strings.Index(outputStr, "KEY FOUND! [ "); idx != -1 {
								start := idx + len("KEY FOUND! [ ")
								end := strings.Index(outputStr[start:], " ]")
								if end != -1 {
									password = outputStr[start : start+end]
								}
							}
							if keyData, err := os.ReadFile(capturePath + ".key"); err == nil {
								password = strings.TrimSpace(string(keyData))
							}
							finalResult = &BruteforceResult{
								Success:  true,
								Password: password,
								SSID:     req.SSID,
								BSSID:    req.BSSID,
								Capture:  capturePath,
								Wordlist: wl,
								Duration: time.Since(startTime).Seconds(),
								TestedAt: time.Now(),
							}
							fmt.Printf("[WIFI-BRUTE] ✅ MOT DE PASSE TROUVÉ (wordlist): %s\n", password)
							break
						} else if err != nil && ctx.Err() != nil {
							// Annulé
							return
						}
					}
				}
			}

			// Phase 2: Pattern (hashcat - si wordlist n'a pas trouvé)
			if finalResult == nil && req.UsePattern && mask != "" {
				// Avertissement pour Free/Freebox - le mask seul est inefficace
				if isp == "Free" {
					fmt.Printf("[WIFI-BRUTE] ⚠️ Freebox détectée: les mots de passe utilisent des mots latins.\n")
					fmt.Printf("[WIFI-BRUTE] ⚠️ Le mode Pattern seul est peu efficace. Recommandation: utiliser une wordlist de mots latins.\n")
				}
				fmt.Printf("[WIFI-BRUTE] 🎯 Phase 2: Pattern ISP (%s, mask: %s)\n", isp, mask)

				// Vérifier si hashcat est disponible
				if CheckHashcatInstalled()["hashcat"] {
					config := BruteforceConfig{
						CapturePath: capturePath,
						BSSID:       req.BSSID,
						SSID:        req.SSID,
						Mode:        ModeMask,
						ISP:         isp,
						MaskPattern: mask,
					}

					result, err := RunMaskAttack(ctx, config, nil)
					if err == nil && result != nil && result.Success {
						finalResult = result
						fmt.Printf("[WIFI-BRUTE] ✅ MOT DE PASSE TROUVÉ (pattern): %s\n", result.Password)
					} else if ctx.Err() != nil {
						if ctx.Err() == context.DeadlineExceeded {
							fmt.Printf("[WIFI-BRUTE] ⏱️ TIMEOUT atteint après %.1f heures\n", time.Since(startTime).Hours())
						}
						// Continue to save result even on timeout
					}
				} else {
					fmt.Println("[WIFI-BRUTE] ⚠️ hashcat non installé, pattern attack ignoré")
				}
			}

			// Phase 3: Attaque combinatoire Freebox (si c'est une Freebox et pas encore trouvé)
			if finalResult == nil && isp == "Free" && ctx.Err() == nil {
				fmt.Println("[WIFI-BRUTE] 🇫🇷 Phase 3: Attaque combinatoire Freebox (mots latins)")

				// Utiliser hashcat en mode combinatoire (-a 1) avec la wordlist de mots latins
				if CheckHashcatInstalled()["hashcat"] {
					freeboxWordlist := GetFreeboxWordlistPath()
					if freeboxWordlist != "" {
						// Convertir le fichier .cap si nécessaire
						hcPath := capturePath
						if strings.HasSuffix(capturePath, ".cap") {
							var err error
							hcPath, err = ConvertCapToHashcat(capturePath)
							if err != nil {
								fmt.Printf("[WIFI-BRUTE] ⚠️ Conversion échouée: %v\n", err)
							}
						}

						if hcPath != "" {
							// Mode combinatoire: combine 2 wordlists
							// On utilise la même wordlist 2 fois pour avoir word1-word2
							args := []string{
								"-m", "22000",
								"-a", "1", // Combinatory attack
								"--status",
								"--status-timer=30",
								"-j", "$-", // Add "-" suffix to left word
								"-o", hcPath + ".cracked",
								hcPath,
								freeboxWordlist,
								freeboxWordlist,
							}

							fmt.Printf("[WIFI-BRUTE] 🔧 Commande hashcat combinatoire: hashcat %v\n", args)

							cmd := exec.CommandContext(ctx, "hashcat", args...)
							output, err := cmd.CombinedOutput()

							if err == nil || strings.Contains(string(output), "Cracked") {
								// Check for cracked password
								crackedFile := hcPath + ".cracked"
								if data, readErr := os.ReadFile(crackedFile); readErr == nil && len(data) > 0 {
									parts := strings.Split(string(data), ":")
									if len(parts) >= 2 {
										password := strings.TrimSpace(parts[len(parts)-1])
										finalResult = &BruteforceResult{
											Success:  true,
											Password: password,
											SSID:     req.SSID,
											BSSID:    req.BSSID,
											Capture:  capturePath,
											Duration: time.Since(startTime).Seconds(),
											TestedAt: time.Now(),
										}
										fmt.Printf("[WIFI-BRUTE] ✅ MOT DE PASSE TROUVÉ (combinatoire): %s\n", password)
									}
								}
							}
						}
					}
				}
			}

			// Vérifier si timeout
			timedOut := ctx.Err() == context.DeadlineExceeded

			// Résultat final
			duration := time.Since(startTime).Seconds()
			if finalResult == nil {
				reason := "non trouvé"
				if timedOut {
					reason = fmt.Sprintf("timeout après %.1fh", duration/3600)
				}
				finalResult = &BruteforceResult{
					Success:  false,
					SSID:     req.SSID,
					BSSID:    req.BSSID,
					Capture:  capturePath,
					Wordlist: wordlist,
					Duration: duration,
					TestedAt: time.Now(),
				}
				fmt.Printf("[WIFI-BRUTE] ❌ Mot de passe %s pour %s (durée: %.1fs)\n", reason, req.SSID, duration)
			}

			p.mu.Lock()
			p.bruteResult = finalResult
			p.mu.Unlock()

			// Sauvegarder le résultat
			resultJSON, _ := json.MarshalIndent(finalResult, "", "  ")
			_ = os.WriteFile(capturePath+".result.json", resultJSON, 0644)

			// Mettre à jour la capture en BDD si mot de passe trouvé
			if finalResult.Success && p.db != nil {
				now := time.Now()
				p.db.Model(&WifiCapture{}).
					Where("capture_path LIKE ?", strings.TrimSuffix(capturePath, ".cap")+"%").
					Updates(map[string]any{
						"cracked":          true,
						"cracked_password": finalResult.Password,
						"cracked_at":       now,
					})
			}

			// Sauvegarder dans la table des résultats de bruteforce (succès ou échec)
			if p.db != nil {
				now := time.Now()
				status := "completed"
				if !finalResult.Success {
					status = "failed"
				}

				wordlistPath := wordlist
				wordlistName := ""
				if wordlist != "" {
					wordlistName = filepath.Base(wordlist)
				}
				if req.UseWordlist && len(req.Wordlists) > 0 {
					wordlistPath = ""
					wordlistName = strings.Join(req.Wordlists, ",")
				} else if !req.UseWordlist && req.UsePattern {
					wordlistPath = ""
					if mask != "" {
						wordlistName = "pattern:" + mask
					}
				}

				bruteResult := WifiBruteforceResult{
					SSID:         finalResult.SSID,
					BSSID:        finalResult.BSSID,
					CapturePath:  capturePath,
					WordlistPath: wordlistPath,
					WordlistName: wordlistName,
					Success:      finalResult.Success,
					Password:     finalResult.Password,
					DurationSecs: duration,
					StartedAt:    startTime,
					Status:       status,
				}
				bruteResult.CompletedAt = &now
				p.db.Create(&bruteResult)
			}
		}()

		return writeJSON(c, 200, map[string]any{
			"status":       "started",
			"message":      fmt.Sprintf("Bruteforce démarré (modes: %v)", modes),
			"capture":      capturePath,
			"modes":        modes,
			"use_wordlist": req.UseWordlist,
			"use_pattern":  req.UsePattern,
			"wordlist":     wordlist,
			"isp":          isp,
			"mask":         mask,
			"bssid":        req.BSSID,
			"ssid":         req.SSID,
			"started_at":   p.bruteStartedAt,
		})
	})

	// Bruteforce - Statut
	group.GET("/bruteforce/status", func(c buffalo.Context) error {
		p.mu.Lock()
		running := p.bruteRunning
		capture := p.bruteCapture
		startedAt := p.bruteStartedAt
		result := p.bruteResult
		p.mu.Unlock()

		response := map[string]any{
			"running": running,
			"capture": capture,
		}

		if running {
			response["started_at"] = startedAt
			response["duration"] = time.Since(startedAt).Seconds()
		}

		if result != nil {
			response["result"] = result
		}

		return writeJSON(c, 200, response)
	})

	// Bruteforce - Arrêter
	group.POST("/bruteforce/stop", func(c buffalo.Context) error {
		p.mu.Lock()
		wasRunning := p.bruteRunning
		if p.bruteCancel != nil {
			p.bruteCancel()
		}
		p.bruteRunning = false
		p.mu.Unlock()

		// Tuer le processus aircrack-ng si nécessaire
		_ = exec.Command("pkill", "-9", "-f", "aircrack-ng").Run()

		return writeJSON(c, 200, map[string]any{
			"status":      "stopped",
			"was_running": wasRunning,
		})
	})

	// Bruteforce - Résultats sauvegardés
	group.GET("/bruteforce/results", func(c buffalo.Context) error {
		if p.db != nil {
			var dbResults []WifiBruteforceResult
			p.db.Order("created_at DESC").Limit(50).Find(&dbResults)

			results := make([]map[string]any, 0, len(dbResults))
			for _, r := range dbResults {
				testedAt := r.CompletedAt
				if testedAt == nil {
					t := r.CreatedAt
					testedAt = &t
				}
				results = append(results, map[string]any{
					"success":          r.Success,
					"password":         r.Password,
					"ssid":             r.SSID,
					"bssid":            r.BSSID,
					"capture":          r.CapturePath,
					"wordlist":         r.WordlistPath,
					"wordlist_name":    r.WordlistName,
					"duration_seconds": r.DurationSecs,
					"tested_at":        testedAt,
					"status":           r.Status,
				})
			}

			return writeJSON(c, 200, map[string]any{
				"results": results,
				"count":   len(results),
			})
		}

		captureDir := "/opt/heimdall/captures"
		files, _ := filepath.Glob(captureDir + "/*.result.json")

		var results []map[string]any
		for _, f := range files {
			data, err := os.ReadFile(f)
			if err != nil {
				continue
			}
			var result map[string]any
			if err := json.Unmarshal(data, &result); err == nil {
				result["result_file"] = f
				results = append(results, result)
			}
		}

		return writeJSON(c, 200, map[string]any{
			"results": results,
			"count":   len(results),
		})
	})

	// Wordlists disponibles (version améliorée avec téléchargement)
	group.GET("/wordlists", func(c buffalo.Context) error {
		// Lister les wordlists connues avec leur statut
		knownWordlists := ListWordlists()
		
		// Lister aussi les fichiers locaux non-répertoriés
		wordlistDirs := []string{
			GetWordlistsDir(),
			"/usr/share/wordlists",
			"/usr/share/seclists/Passwords",
		}

		var localWordlists []map[string]any
		seenPaths := make(map[string]bool)
		
		// Marquer les wordlists connues comme vues
		for _, wl := range knownWordlists {
			if wl.LocalPath != "" {
				seenPaths[wl.LocalPath] = true
			}
		}

		for _, dir := range wordlistDirs {
			files, err := os.ReadDir(dir)
			if err != nil {
				continue
			}
			for _, f := range files {
				if f.IsDir() {
					continue
				}
				info, _ := f.Info()
				path := filepath.Join(dir, f.Name())
				
				// Skip si déjà dans les wordlists connues
				if seenPaths[path] {
					continue
				}

				// Compter le nombre de lignes
				var lineCount int64
				if info != nil {
					if info.Size() < 200*1024*1024 { // < 200MB: count exactly
						if data, err := exec.Command("wc", "-l", path).Output(); err == nil {
							fmt.Sscanf(string(data), "%d", &lineCount)
						}
					} else {
						// For very large files, estimate (~10 bytes per line average)
						lineCount = info.Size() / 10
					}
				}

				localWordlists = append(localWordlists, map[string]any{
					"path":      path,
					"name":      f.Name(),
					"size":      info.Size(),
					"lines":     lineCount,
					"installed": true,
					"category":  "local",
				})
			}
		}

		return writeJSON(c, 200, map[string]any{
			"known":  knownWordlists,
			"local":  localWordlists,
			"count":  len(knownWordlists) + len(localWordlists),
		})
	})
	
	// Télécharger une wordlist
	group.POST("/wordlists/download", func(c buffalo.Context) error {
		var req struct {
			Name string `json:"name"`
		}
		_ = json.NewDecoder(c.Request().Body).Decode(&req)
		
		if req.Name == "" {
			return writeJSON(c, 400, map[string]any{"error": "Nom de wordlist manquant"})
		}
		
		path, err := DownloadWordlist(req.Name)
		if err != nil {
			return writeJSON(c, 500, map[string]any{"error": err.Error()})
		}
		
		return writeJSON(c, 200, map[string]any{
			"success": true,
			"path":    path,
			"message": fmt.Sprintf("Wordlist %s téléchargée avec succès", req.Name),
		})
	})
	
	// Obtenir les meilleures wordlists pour un ISP donné
	group.GET("/wordlists/recommend/{isp}", func(c buffalo.Context) error {
		isp := c.Param("isp")
		if isp == "" {
			isp = "unknown"
		}
		
		recommended := GetBestWordlistsForISP(isp)
		
		return writeJSON(c, 200, map[string]any{
			"isp":         isp,
			"recommended": recommended,
			"count":       len(recommended),
		})
	})

	// ========== VENDOR / ISP DETECTION ==========

	// Analyser un réseau (BSSID + SSID) pour identifier le vendeur et le pattern de mot de passe
	group.POST("/analyze", func(c buffalo.Context) error {
		var req struct {
			BSSID string `json:"bssid"`
			SSID  string `json:"ssid"`
		}
		_ = json.NewDecoder(c.Request().Body).Decode(&req)

		if req.BSSID == "" {
			return writeJSON(c, 400, map[string]any{"error": "BSSID manquant"})
		}

		analysis := AnalyzeNetwork(req.BSSID, req.SSID)
		return writeJSON(c, 200, analysis)
	})

	// Lookup vendeur par BSSID
	group.GET("/vendor/{bssid}", func(c buffalo.Context) error {
		bssid := c.Param("bssid")
		if bssid == "" {
			return writeJSON(c, 400, map[string]any{"error": "BSSID manquant"})
		}

		vendor := LookupVendor(bssid)
		isp := vendor.ISP
		if isp == "" {
			// Essayer de détecter via un SSID passé en query param
			ssid := c.Request().URL.Query().Get("ssid")
			if ssid != "" {
				isp = DetectISPFromSSID(ssid)
			}
		}

		result := map[string]any{
			"bssid":        bssid,
			"manufacturer": vendor.Manufacturer,
			"isp":          isp,
			"router_model": vendor.RouterModel,
			"country":      vendor.Country,
		}

		// Ajouter le pattern de mot de passe si ISP connu
		if isp != "" {
			if pattern, ok := GetPasswordPattern(isp); ok {
				result["password_pattern"] = pattern
				result["suggestion"] = GenerateWordlistSuggestion(isp)
			}
		}

		return writeJSON(c, 200, result)
	})

	// Obtenir tous les patterns de mots de passe connus
	group.GET("/patterns", func(c buffalo.Context) error {
		patterns := make([]map[string]any, 0)
		for isp, pattern := range passwordPatterns {
			patterns = append(patterns, map[string]any{
				"isp":            isp,
				"description":    pattern.Description,
				"length":         pattern.Length,
				"character_set":  pattern.CharacterSet,
				"format":         pattern.Format,
				"example":        pattern.Example,
				"generator_type": pattern.GeneratorType,
				"mask_pattern":   pattern.MaskPattern,
				"notes":          pattern.Notes,
			})
		}
		return writeJSON(c, 200, map[string]any{
			"patterns": patterns,
			"count":    len(patterns),
		})
	})

	// Générer une suggestion de wordlist pour un ISP
	group.GET("/suggest/{isp}", func(c buffalo.Context) error {
		isp := c.Param("isp")
		if isp == "" {
			return writeJSON(c, 400, map[string]any{"error": "ISP manquant"})
		}

		suggestion := GenerateWordlistSuggestion(isp)
		return writeJSON(c, 200, suggestion)
	})

	// ========== MASK/PATTERN BRUTEFORCE (hashcat) ==========

	// Vérifier les outils disponibles
	group.GET("/bruteforce/tools", func(c buffalo.Context) error {
		tools := CheckHashcatInstalled()
		return writeJSON(c, 200, map[string]any{
			"tools": tools,
			"ready": tools["hashcat"] && tools["hcxpcapngtool"],
		})
	})

	// Obtenir les masques disponibles pour un ISP
	group.GET("/bruteforce/masks/{isp}", func(c buffalo.Context) error {
		isp := c.Param("isp")
		masks := GetMasksForISP(isp)
		if len(masks) == 0 {
			return writeJSON(c, 404, map[string]any{"error": "Aucun masque pour cet ISP"})
		}
		return writeJSON(c, 200, map[string]any{
			"isp":   isp,
			"masks": masks,
		})
	})

	// Lancer un bruteforce par pattern (mask attack)
	group.POST("/bruteforce/mask", func(c buffalo.Context) error {
		var req struct {
			CapturePath   string `json:"capture_path"`
			BSSID         string `json:"bssid"`
			SSID          string `json:"ssid"`
			ISP           string `json:"isp,omitempty"`
			MaskPattern   string `json:"mask_pattern,omitempty"`
			CustomCharset string `json:"custom_charset,omitempty"`
		}
		_ = json.NewDecoder(c.Request().Body).Decode(&req)

		if req.CapturePath == "" {
			return writeJSON(c, 400, map[string]any{"error": "Fichier capture manquant"})
		}

		// Vérifier que hashcat est installé
		tools := CheckHashcatInstalled()
		if !tools["hashcat"] {
			return writeJSON(c, 500, map[string]any{
				"error": "hashcat non installé",
				"hint":  "sudo apt install hashcat hcxtools",
			})
		}

		config := BruteforceConfig{
			CapturePath:   req.CapturePath,
			BSSID:         req.BSSID,
			SSID:          req.SSID,
			Mode:          ModeMask,
			ISP:           req.ISP,
			MaskPattern:   req.MaskPattern,
			CustomCharset: req.CustomCharset,
		}

		// Lancer en background
		ctx, cancel := context.WithCancel(context.Background())
		progressChan := make(chan BruteforceState, 10)

		p.mu.Lock()
		if p.bruteRunning {
			p.mu.Unlock()
			cancel()
			return writeJSON(c, 409, map[string]any{"error": "Un bruteforce est déjà en cours"})
		}
		p.bruteRunning = true
		p.bruteCapture = req.CapturePath
		p.bruteStartedAt = time.Now()
		p.bruteCancel = cancel
		p.mu.Unlock()

		go func() {
			defer func() {
				p.mu.Lock()
				p.bruteRunning = false
				p.bruteCancel = nil
				p.mu.Unlock()
				close(progressChan)
			}()

			result, err := RunMaskAttack(ctx, config, progressChan)
			if err != nil {
				fmt.Printf("[WIFI-MASK] Erreur: %v\n", err)
				return
			}

			p.mu.Lock()
			p.bruteResult = result
			p.mu.Unlock()

			// Sauvegarder en BDD
			if p.db != nil && result != nil {
				bruteResult := WifiBruteforceResult{
					SSID:         result.SSID,
					BSSID:        result.BSSID,
					CapturePath:  result.Capture,
					WordlistName: "mask:" + config.MaskPattern,
					Success:      result.Success,
					Password:     result.Password,
					DurationSecs: result.Duration,
					StartedAt:    p.bruteStartedAt,
					Status:       "completed",
				}
				now := time.Now()
				bruteResult.CompletedAt = &now
				p.db.Create(&bruteResult)

				if result.Success {
					fmt.Printf("[WIFI-MASK] ✅ MOT DE PASSE TROUVÉ: %s\n", result.Password)
				}
			}
		}()

		return writeJSON(c, 200, map[string]any{
			"status":     "started",
			"mode":       "mask",
			"capture":    req.CapturePath,
			"mask":       config.MaskPattern,
			"isp":        req.ISP,
			"started_at": p.bruteStartedAt,
		})
	})

	// Lancer un bruteforce incrémental
	group.POST("/bruteforce/increment", func(c buffalo.Context) error {
		var req struct {
			CapturePath string `json:"capture_path"`
			BSSID       string `json:"bssid"`
			SSID        string `json:"ssid"`
			MinLength   int    `json:"min_length"`
			MaxLength   int    `json:"max_length"`
			CharsetType string `json:"charset_type"` // digits, lower, upper, alnum, all
		}
		_ = json.NewDecoder(c.Request().Body).Decode(&req)

		if req.CapturePath == "" {
			return writeJSON(c, 400, map[string]any{"error": "Fichier capture manquant"})
		}

		config := BruteforceConfig{
			CapturePath: req.CapturePath,
			BSSID:       req.BSSID,
			SSID:        req.SSID,
			Mode:        ModeIncrement,
			MinLength:   req.MinLength,
			MaxLength:   req.MaxLength,
			CharsetType: req.CharsetType,
		}

		// Lancer en background (similaire à mask)
		ctx, cancel := context.WithCancel(context.Background())

		p.mu.Lock()
		if p.bruteRunning {
			p.mu.Unlock()
			cancel()
			return writeJSON(c, 409, map[string]any{"error": "Un bruteforce est déjà en cours"})
		}
		p.bruteRunning = true
		p.bruteCapture = req.CapturePath
		p.bruteStartedAt = time.Now()
		p.bruteCancel = cancel
		p.mu.Unlock()

		go func() {
			defer func() {
				p.mu.Lock()
				p.bruteRunning = false
				p.bruteCancel = nil
				p.mu.Unlock()
			}()

			result, err := RunIncrementalAttack(ctx, config, nil)
			if err != nil {
				fmt.Printf("[WIFI-INCREMENT] Erreur: %v\n", err)
				return
			}

			p.mu.Lock()
			p.bruteResult = result
			p.mu.Unlock()

			if p.db != nil && result != nil {
				bruteResult := WifiBruteforceResult{
					SSID:         result.SSID,
					BSSID:        result.BSSID,
					CapturePath:  result.Capture,
					WordlistName: fmt.Sprintf("increment:%d-%d:%s", config.MinLength, config.MaxLength, config.CharsetType),
					Success:      result.Success,
					Password:     result.Password,
					DurationSecs: result.Duration,
					StartedAt:    p.bruteStartedAt,
					Status:       "completed",
				}
				now := time.Now()
				bruteResult.CompletedAt = &now
				p.db.Create(&bruteResult)
			}
		}()

		return writeJSON(c, 200, map[string]any{
			"status":     "started",
			"mode":       "increment",
			"capture":    req.CapturePath,
			"min_length": config.MinLength,
			"max_length": config.MaxLength,
			"charset":    config.CharsetType,
			"started_at": p.bruteStartedAt,
		})
	})

	// ========== AUDITS & RAPPORTS ==========

	// Créer un nouvel audit
	group.POST("/audits", func(c buffalo.Context) error {
		var req struct {
			ClientName    string `json:"client_name"`
			ClientContact string `json:"client_contact,omitempty"`
			Location      string `json:"location,omitempty"`
			TesterName    string `json:"tester_name,omitempty"`
			AuditType     string `json:"audit_type,omitempty"`
			Notes         string `json:"notes,omitempty"`
		}
		_ = json.NewDecoder(c.Request().Body).Decode(&req)

		if req.ClientName == "" {
			return writeJSON(c, 400, map[string]any{"error": "Nom du client requis"})
		}

		audit := WifiAudit{
			ClientName:    req.ClientName,
			ClientContact: req.ClientContact,
			Location:      req.Location,
			TesterName:    req.TesterName,
			AuditType:     req.AuditType,
			StartDate:     time.Now(),
			Status:        "in_progress",
			Notes:         req.Notes,
		}

		if p.db != nil {
			if err := p.db.Create(&audit).Error; err != nil {
				return writeJSON(c, 500, map[string]any{"error": "Erreur création audit"})
			}
		}

		return writeJSON(c, 201, map[string]any{
			"status": "created",
			"audit":  audit,
		})
	})

	// Lister les audits
	group.GET("/audits", func(c buffalo.Context) error {
		var audits []WifiAudit
		if p.db != nil {
			p.db.Order("created_at DESC").Find(&audits)
		}
		return writeJSON(c, 200, map[string]any{
			"audits": audits,
			"count":  len(audits),
		})
	})

	// Obtenir un audit spécifique avec ses réseaux
	group.GET("/audits/{id}", func(c buffalo.Context) error {
		id := c.Param("id")
		var audit WifiAudit
		if p.db != nil {
			if err := p.db.First(&audit, "id = ?", id).Error; err != nil {
				return writeJSON(c, 404, map[string]any{"error": "Audit non trouvé"})
			}
		}

		// Récupérer les réseaux liés
		var networks []WifiAuditNetwork
		if p.db != nil {
			p.db.Where("audit_id = ?", id).Find(&networks)
		}

		return writeJSON(c, 200, map[string]any{
			"audit":    audit,
			"networks": networks,
		})
	})

	// Ajouter un réseau à un audit
	group.POST("/audits/{id}/networks", func(c buffalo.Context) error {
		auditID := c.Param("id")

		var req struct {
			SSID              string  `json:"ssid"`
			BSSID             string  `json:"bssid"`
			Channel           int     `json:"channel"`
			Security          string  `json:"security"`
			Vendor            string  `json:"vendor,omitempty"`
			ISP               string  `json:"isp,omitempty"`
			HandshakeCaptured bool    `json:"handshake_captured"`
			PasswordCracked   bool    `json:"password_cracked"`
			Password          string  `json:"password,omitempty"`
			CrackMethod       string  `json:"crack_method,omitempty"`
			CrackDuration     float64 `json:"crack_duration,omitempty"`
		}
		_ = json.NewDecoder(c.Request().Body).Decode(&req)

		auditUUID, err := uuid.Parse(auditID)
		if err != nil {
			return writeJSON(c, 400, map[string]any{"error": "ID audit invalide"})
		}

		network := WifiAuditNetwork{
			AuditID:           auditUUID,
			SSID:              req.SSID,
			BSSID:             req.BSSID,
			Channel:           req.Channel,
			Security:          req.Security,
			Vendor:            req.Vendor,
			ISP:               req.ISP,
			HandshakeCaptured: req.HandshakeCaptured,
			PasswordCracked:   req.PasswordCracked,
			Password:          req.Password,
			CrackMethod:       req.CrackMethod,
			CrackDuration:     req.CrackDuration,
		}

		// Déterminer le niveau de vulnérabilité
		netEntry := NetworkReportEntry{
			Security:          req.Security,
			HandshakeCaptured: req.HandshakeCaptured,
			PasswordCracked:   req.PasswordCracked,
		}
		network.VulnerabilityLevel = DetermineVulnerabilityLevel(netEntry)

		if p.db != nil {
			if err := p.db.Create(&network).Error; err != nil {
				return writeJSON(c, 500, map[string]any{"error": "Erreur ajout réseau"})
			}

			// Mettre à jour les compteurs de l'audit
			p.db.Model(&WifiAudit{}).Where("id = ?", auditID).Updates(map[string]any{
				"networks_tested":     gorm.Expr("networks_tested + 1"),
				"handshakes_captured": gorm.Expr("handshakes_captured + ?", boolToInt(req.HandshakeCaptured)),
				"passwords_cracked":   gorm.Expr("passwords_cracked + ?", boolToInt(req.PasswordCracked)),
			})
		}

		return writeJSON(c, 201, map[string]any{
			"status":  "added",
			"network": network,
		})
	})

	// Terminer un audit
	group.POST("/audits/{id}/complete", func(c buffalo.Context) error {
		id := c.Param("id")
		now := time.Now()

		if p.db != nil {
			p.db.Model(&WifiAudit{}).Where("id = ?", id).Updates(map[string]any{
				"status":   "completed",
				"end_date": now,
			})
		}

		return writeJSON(c, 200, map[string]any{
			"status":   "completed",
			"end_date": now,
		})
	})

	// Générer le rapport PDF
	group.POST("/audits/{id}/report", func(c buffalo.Context) error {
		id := c.Param("id")

		var req struct {
			TesterName  string `json:"tester_name,omitempty"`
			CompanyLogo string `json:"company_logo,omitempty"`
		}
		_ = json.NewDecoder(c.Request().Body).Decode(&req)

		// Récupérer l'audit
		var audit WifiAudit
		if p.db != nil {
			if err := p.db.First(&audit, "id = ?", id).Error; err != nil {
				return writeJSON(c, 404, map[string]any{"error": "Audit non trouvé"})
			}
		}

		// Récupérer les réseaux
		var auditNetworks []WifiAuditNetwork
		if p.db != nil {
			p.db.Where("audit_id = ?", id).Find(&auditNetworks)
		}

		// Construire les données du rapport
		networks := make([]NetworkReportEntry, len(auditNetworks))
		for i, n := range auditNetworks {
			networks[i] = NetworkReportEntry{
				SSID:               n.SSID,
				BSSID:              n.BSSID,
				Channel:            n.Channel,
				Security:           n.Security,
				Vendor:             n.Vendor,
				ISP:                n.ISP,
				HandshakeCaptured:  n.HandshakeCaptured,
				PasswordCracked:    n.PasswordCracked,
				Password:           n.Password,
				CrackMethod:        n.CrackMethod,
				CrackDuration:      n.CrackDuration,
				VulnerabilityLevel: n.VulnerabilityLevel,
				TestedAt:           n.CreatedAt,
			}
		}

		endDate := time.Now()
		if audit.EndDate != nil {
			endDate = *audit.EndDate
		}

		tester := req.TesterName
		if tester == "" {
			tester = audit.TesterName
		}

		reportData := ReportData{
			ClientName:         audit.ClientName,
			ClientContact:      audit.ClientContact,
			Location:           audit.Location,
			AuditID:            audit.ID.String(),
			StartDate:          audit.StartDate,
			EndDate:            endDate,
			TesterName:         tester,
			NetworksScanned:    audit.NetworksScanned,
			NetworksTested:     audit.NetworksTested,
			HandshakesCaptured: audit.HandshakesCaptured,
			PasswordsCracked:   audit.PasswordsCracked,
			Networks:           networks,
		}

		config := ReportConfig{
			AuditID:     audit.ID,
			ClientName:  audit.ClientName,
			TesterName:  tester,
			CompanyLogo: req.CompanyLogo,
		}

		// Générer le PDF
		reportPath, err := GenerateReport(reportData, config)
		if err != nil {
			return writeJSON(c, 500, map[string]any{"error": fmt.Sprintf("Erreur génération PDF: %v", err)})
		}

		// Mettre à jour l'audit
		now := time.Now()
		if p.db != nil {
			p.db.Model(&WifiAudit{}).Where("id = ?", id).Updates(map[string]any{
				"report_generated":    true,
				"report_path":         reportPath,
				"report_generated_at": now,
			})
		}

		return writeJSON(c, 200, map[string]any{
			"status":       "generated",
			"report_path":  reportPath,
			"generated_at": now,
		})
	})

	// Télécharger un rapport
	group.GET("/audits/{id}/report/download", func(c buffalo.Context) error {
		id := c.Param("id")

		var audit WifiAudit
		if p.db != nil {
			if err := p.db.First(&audit, "id = ?", id).Error; err != nil {
				return writeJSON(c, 404, map[string]any{"error": "Audit non trouvé"})
			}
		}

		if !audit.ReportGenerated || audit.ReportPath == "" {
			return writeJSON(c, 404, map[string]any{"error": "Rapport non généré"})
		}

		// Servir le fichier PDF
		c.Response().Header().Set("Content-Type", "application/pdf")
		c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"",
			filepath.Base(audit.ReportPath)))

		http.ServeFile(c.Response(), c.Request(), audit.ReportPath)
		return nil
	})

	// Lister les rapports générés
	group.GET("/reports", func(c buffalo.Context) error {
		var audits []WifiAudit
		if p.db != nil {
			p.db.Where("report_generated = ?", true).Order("report_generated_at DESC").Find(&audits)
		}

		reports := make([]map[string]any, len(audits))
		for i, a := range audits {
			reports[i] = map[string]any{
				"audit_id":     a.ID,
				"client_name":  a.ClientName,
				"report_path":  a.ReportPath,
				"generated_at": a.ReportGeneratedAt,
				"start_date":   a.StartDate,
				"end_date":     a.EndDate,
			}
		}

		return writeJSON(c, 200, map[string]any{
			"reports": reports,
			"count":   len(reports),
		})
	})
}

func init() {
	p := &WiFiPlugin{startedAt: time.Now()}
	plugins.Register(p)
}

func writeJSON(c buffalo.Context, status int, payload any) error {
	c.Response().Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Response().WriteHeader(status)
	_ = json.NewEncoder(c.Response()).Encode(payload)
	return nil
}

// getWiFiInterfaces retourne la liste des interfaces réseau sans-fil
func getWiFiInterfaces() ([]map[string]any, error) {
	var result []map[string]any

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		name := iface.Name

		// Vérifier si c'est une interface WiFi via /sys/class/net/{iface}/wireless
		// C'est la méthode fiable sur Linux (fonctionne avec wlan*, wlp*, wlx*, etc.)
		wirelessPath := fmt.Sprintf("/sys/class/net/%s/wireless", name)
		_, wirelessErr := os.Stat(wirelessPath)
		isWiFi := wirelessErr == nil

		// Fallback: heuristique sur le nom si /sys n'est pas disponible
		if !isWiFi {
			nameLower := strings.ToLower(name)
			isWiFi = strings.HasPrefix(nameLower, "wlan") ||
				strings.HasPrefix(nameLower, "wlp") ||
				strings.HasPrefix(nameLower, "wlx") || // USB WiFi adapters
				strings.HasPrefix(nameLower, "wifi") ||
				strings.HasPrefix(nameLower, "ath") ||
				strings.HasPrefix(nameLower, "ra") ||
				strings.Contains(nameLower, "wireless") ||
				strings.Contains(nameLower, "wi-fi")
		}

		if !isWiFi {
			continue
		}

		// Vérifier si en mode monitor (Linux: nom finit par "mon")
		isMonitor := strings.HasSuffix(strings.ToLower(name), "mon")

		info := map[string]any{
			"name":    iface.Name,
			"mac":     iface.HardwareAddr.String(),
			"monitor": isMonitor,
			"up":      iface.Flags&net.FlagUp != 0,
			"index":   iface.Index,
		}

		// Essayer d'obtenir le driver (Linux only)
		driver := getInterfaceDriver(iface.Name)
		if driver != "" {
			info["driver"] = driver
		}

		result = append(result, info)
	}

	return result, nil
}

// getInterfaceDriver tente de récupérer le driver de l'interface (Linux)
func getInterfaceDriver(ifaceName string) string {
	// Lire le lien symbolique /sys/class/net/<iface>/device/driver
	out, err := exec.Command("readlink", "-f", "/sys/class/net/"+ifaceName+"/device/driver").Output()
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.TrimSpace(string(out)), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// ifaceExists vérifie si une interface réseau existe
func ifaceExists(name string) bool {
	_, err := net.InterfaceByName(name)
	return err == nil
}

// scanWiFi exécute "nmcli device wifi list" et parse les réseaux (ne nécessite pas sudo)
func scanWiFi(iface string) ([]WiFiNetwork, error) {
	if strings.TrimSpace(iface) == "" {
		return nil, errors.New("interface manquante")
	}

	// Utiliser nmcli qui ne nécessite pas sudo
	cmd := exec.Command("nmcli", "-t", "-f", "BSSID,SSID,MODE,CHAN,RATE,SIGNAL,SECURITY", "device", "wifi", "list", "ifname", iface)
	out, err := cmd.Output()
	if err != nil {
		// Fallback: essayer sans ifname spécifique
		cmd = exec.Command("nmcli", "-t", "-f", "BSSID,SSID,MODE,CHAN,RATE,SIGNAL,SECURITY", "device", "wifi", "list")
		out, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("scan échoué: %w", err)
		}
	}

	return parseNmcliScan(string(out), iface), nil
}

// parseNmcliScan transforme la sortie nmcli -t en slice de WiFiNetwork
func parseNmcliScan(out string, iface string) []WiFiNetwork {
	var results []WiFiNetwork

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// nmcli échappe les ":" dans les valeurs avec "\:"
		// On remplace temporairement "\:" par un placeholder
		const placeholder = "##COLON##"
		escaped := strings.ReplaceAll(line, `\:`, placeholder)

		// Format nmcli -t: BSSID:SSID:MODE:CHAN:RATE:SIGNAL:SECURITY
		parts := strings.Split(escaped, ":")
		if len(parts) < 7 {
			continue
		}

		// Restaurer les ":" dans chaque partie
		for i := range parts {
			parts[i] = strings.ReplaceAll(parts[i], placeholder, ":")
		}

		// Format: BSSID:SSID:MODE:CHAN:RATE:SIGNAL:SECURITY
		// Après unescape, BSSID est déjà complet (ex: "22:66:CF:57:C4:20")
		bssid := parts[0]

		// Compter depuis la fin pour les champs fixes
		n := len(parts)
		security := parts[n-1]
		signalStr := parts[n-2]
		_ = parts[n-3] // rateStr - non utilisé
		chanStr := parts[n-4]
		_ = parts[n-5] // mode - non utilisé

		// SSID = tout ce qui reste entre BSSID et MODE (index 1 à n-6)
		ssid := ""
		if n > 6 {
			ssid = strings.Join(parts[1:n-5], ":")
		}

		channel, _ := strconv.Atoi(chanStr)
		signal, _ := strconv.Atoi(signalStr)

		// Convertir le signal en dBm approximatif (nmcli donne un %)
		// 100% ≈ -30dBm, 0% ≈ -90dBm
		signalDbm := -90 + (signal * 60 / 100)

		securityType := "open"
		if strings.Contains(security, "WPA3") {
			securityType = "WPA3"
		} else if strings.Contains(security, "WPA2") || strings.Contains(security, "WPA1") {
			securityType = "WPA/WPA2"
		} else if strings.Contains(security, "WEP") {
			securityType = "WEP"
		}

		network := WiFiNetwork{
			BSSID:     strings.ToLower(bssid),
			SSID:      ssid,
			Channel:   channel,
			Signal:    signalDbm,
			Security:  securityType,
			Interface: iface,
			LastSeen:  time.Now(),
		}

		// Enrichir avec les infos vendeur
		vendor := LookupVendor(bssid)
		if vendor.ISP != "" {
			network.Vendor = vendor.ISP
		} else if vendor.Manufacturer != "" {
			network.Vendor = vendor.Manufacturer
		}

		results = append(results, network)
	}

	return results
}

// freqToChannel convertit la fréquence MHz en numéro de canal approximatif
func freqToChannel(freq int) int {
	if freq >= 2412 && freq <= 2472 {
		return (freq - 2407) / 5
	}
	if freq == 2484 {
		return 14
	}
	if freq >= 5000 && freq <= 5895 {
		return (freq - 5000) / 5
	}
	return 0
}

// checkToolInstalled vérifie si un outil est disponible
func checkToolInstalled(tool string) bool {
	// Essayer avec which
	if err := exec.Command("which", tool).Run(); err == nil {
		return true
	}
	// Vérifier les chemins courants
	paths := []string{
		"/usr/sbin/" + tool,
		"/usr/bin/" + tool,
		"/sbin/" + tool,
		"/bin/" + tool,
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

// findTool retourne le chemin complet d'un outil
func findTool(tool string) string {
	paths := []string{
		"/usr/sbin/" + tool,
		"/usr/bin/" + tool,
		"/sbin/" + tool,
		"/bin/" + tool,
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return tool // fallback au nom simple
}

// enableMonitorMode tente de passer l'interface en mode monitor via airmon-ng ou iw
func enableMonitorMode(iface string) error {
	// Vérifier si airmon-ng est installé
	airmonPath := findTool("airmon-ng")
	iwPath := findTool("iw")
	ipPath := findTool("ip")

	hasAirmon := checkToolInstalled("airmon-ng")
	hasIw := checkToolInstalled("iw")

	if !hasAirmon && !hasIw {
		return fmt.Errorf("aircrack-ng et iw non installés. Installez avec: sudo apt install aircrack-ng iw")
	}

	// NOTE: On ne tue PAS NetworkManager pour garder la connexion internet active
	// sur les autres interfaces WiFi. On utilise iw directement sur l'interface cible.

	// Méthode simple et directe avec iw (ne coupe pas les autres interfaces)
	if hasIw {
		// D'abord désactiver l'interface
		exec.Command(ipPath, "link", "set", iface, "down").Run()
		time.Sleep(200 * time.Millisecond)

		// Passer en mode monitor
		cmd := exec.Command(iwPath, "dev", iface, "set", "type", "monitor")
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Réactiver l'interface même en cas d'erreur
			exec.Command(ipPath, "link", "set", iface, "up").Run()

			// Message d'erreur explicite
			outputStr := string(output)
			if strings.Contains(outputStr, "Operation not supported") || strings.Contains(outputStr, "-95") {
				return fmt.Errorf("Cette carte WiFi (%s) ne supporte pas le mode monitor", iface)
			}
			if strings.Contains(outputStr, "Device or resource busy") {
				// L'interface est utilisée par NetworkManager, on doit la libérer
				// Utiliser nmcli pour déconnecter cette interface spécifique
				exec.Command("nmcli", "device", "disconnect", iface).Run()
				time.Sleep(300 * time.Millisecond)

				// Réessayer
				exec.Command(ipPath, "link", "set", iface, "down").Run()
				time.Sleep(100 * time.Millisecond)
				cmd2 := exec.Command(iwPath, "dev", iface, "set", "type", "monitor")
				if err2 := cmd2.Run(); err2 != nil {
					exec.Command(ipPath, "link", "set", iface, "up").Run()
					return fmt.Errorf("iw set monitor failed après disconnect: %v", err2)
				}
			} else {
				return fmt.Errorf("iw set monitor failed: %v (output: %s)", err, outputStr)
			}
		}

		// Réactiver l'interface en mode monitor
		exec.Command(ipPath, "link", "set", iface, "up").Run()
		return nil
	}

	// Fallback: airmon-ng (mais ça risque de couper NetworkManager)
	if hasAirmon {
		cmd := exec.Command(airmonPath, "start", iface)
		output, err := cmd.CombinedOutput()
		if err == nil {
			return nil
		}
		// Log pour debug
		fmt.Printf("airmon-ng start failed: %v, output: %s\n", err, string(output))
	}

	return fmt.Errorf("impossible d'activer le mode monitor - installez aircrack-ng: sudo apt install aircrack-ng")
}

// disableMonitorMode désactive le mode monitor et restaure le mode managed
func disableMonitorMode(iface string) error {
	airmonPath := findTool("airmon-ng")
	iwPath := findTool("iw")
	ipPath := findTool("ip")

	hasAirmon := checkToolInstalled("airmon-ng")

	// Si l'interface finit par "mon", c'est une interface monitor créée par airmon-ng
	if strings.HasSuffix(iface, "mon") {
		if hasAirmon {
			cmd := exec.Command(airmonPath, "stop", iface)
			if err := cmd.Run(); err == nil {
				return nil
			}
		}
		// Fallback: supprimer l'interface virtuelle
		exec.Command(iwPath, "dev", iface, "del").Run()
		return nil
	}

	// Sinon, remettre l'interface en mode managed
	// D'abord, la désactiver
	exec.Command(ipPath, "link", "set", iface, "down").Run()

	// Remettre en mode managed
	cmd := exec.Command(iwPath, "dev", iface, "set", "type", "managed")
	if err := cmd.Run(); err != nil {
		// Essayer via iwconfig (legacy)
		exec.Command("iwconfig", iface, "mode", "managed").Run()
	}

	// Réactiver l'interface
	exec.Command(ipPath, "link", "set", iface, "up").Run()

	return nil
}

// boolToInt converts a boolean to int (0 or 1) for SQL expressions
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// getWordlistsForView returns wordlists formatted for the view builder
func getWordlistsForView() []map[string]any {
	wordlistDirs := []string{
		GetWordlistsDir(),
		"/usr/share/wordlists",
		"/usr/share/seclists/Passwords",
	}

	var wordlists []map[string]any
	for _, dir := range wordlistDirs {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			info, _ := f.Info()
			if info == nil {
				continue
			}
			path := filepath.Join(dir, f.Name())

			// Compter le nombre de lignes (estimation rapide pour gros fichiers)
			var lineCount int64
			if info.Size() < 200*1024*1024 {
				if data, err := exec.Command("wc", "-l", path).Output(); err == nil {
					fmt.Sscanf(string(data), "%d", &lineCount)
				}
			} else {
				lineCount = info.Size() / 10
			}

			wordlists = append(wordlists, map[string]any{
				"path":  path,
				"name":  f.Name(),
				"size":  info.Size(),
				"lines": lineCount,
			})
		}
	}
	return wordlists
}
