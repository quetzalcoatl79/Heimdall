package wifi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
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
	"github.com/nxo/engine/internal/models"
	"github.com/nxo/engine/internal/plugins"
	"github.com/nxo/engine/internal/ui"
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
	deauthRunning  bool
	deauthCmd      *exec.Cmd
	deauthCancel   context.CancelFunc
	deauthTarget   string
	deauthBSSID    string
	deauthCount    int
	deauthSent     int
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

	// UI schema
	group.GET("/view", func(c buffalo.Context) error {
		view := ui.NewView("Pentest Wi-Fi").
			WithDescription("Scan, capture de handshakes et bruteforce WPA/WEP").
			WithIcon("wifi").
			WithRefresh(5)

		// Actions globales
		view.AddAction(ui.Action{ID: "refresh", Label: "Rafraîchir", Icon: "refresh", Variant: "secondary"})
		view.AddAction(ui.Action{ID: "scan", Label: "Scanner", Icon: "radar", Variant: "primary"})

		// Sélecteur d'interface
		view.AddComponent(ui.Card("Interface Wi-Fi",
			ui.Form("wifi-select-iface",
				[]ui.FormField{
					ui.SelectField("interface", "Interface", []ui.SelectOption{{Value: "", Label: "Choisir"}}).
						WithHelp("Interfaces monitor/injection détectées"),
				},
				ui.WithSubmitLabel("Utiliser cette interface"),
			),
		))

		// Tableau des réseaux détectés avec filtres et tri
		cols := []ui.TableColumn{
			{Key: "ssid", Label: "SSID", Sortable: true, Filterable: true},
			{Key: "bssid", Label: "BSSID", Filterable: true},
			{Key: "channel", Label: "Canal", Sortable: true, Filterable: true, FilterType: "select"},
			{Key: "signal", Label: "Signal", Sortable: true, Render: "signal"},
			{Key: "security", Label: "Sécurité", Sortable: true, Filterable: true, FilterType: "select", Render: "badge"},
			{Key: "vendor", Label: "Vendor", Filterable: true},
			{Key: "wps", Label: "WPS", Filterable: true, FilterType: "select"},
		}
		table := ui.TableWithOptions(cols, []map[string]any{},
			ui.TableFilterable(),
			ui.TableSearchable(),
			ui.TableSelectable(),
			ui.TablePaginated(15),
			ui.TableRowKey("bssid"),
			ui.TableEmptyMessage("Aucun réseau détecté. Cliquez sur 'Scanner' pour démarrer."),
		)
		table.ID = "wifi-networks"
		view.AddComponent(ui.Card("Réseaux Wi-Fi détectés", table))

		// Actions de capture
		view.AddComponent(ui.Card("Capture",
			ui.Grid(2,
				ui.Form("wifi-capture",
					[]ui.FormField{
						ui.TextField("targets", "BSSIDs ciblés").WithPlaceholder("bssid1,bssid2"),
						ui.TextField("channel", "Canal optionnel"),
						ui.SelectField("mode", "Mode", []ui.SelectOption{
							{Value: "wpa", Label: "WPA/WPA2 Handshake"},
							{Value: "wpa3", Label: "WPA3 (PMKID)"},
							{Value: "wep", Label: "WEP IVs"},
						}),
						ui.SelectField("duration", "Durée", []ui.SelectOption{
							{Value: "60", Label: "1 minute (test rapide)"},
							{Value: "300", Label: "5 minutes (WPA2 recommandé)"},
							{Value: "600", Label: "10 minutes"},
							{Value: "1800", Label: "30 minutes (WPA3/WEP)"},
							{Value: "3600", Label: "1 heure"},
							{Value: "0", Label: "Illimité (arrêt manuel)"},
						}).WithHelp("WPA2: 5min suffisent. WPA3/WEP: 30min+ recommandé"),
					},
					ui.WithSubmitLabel("Lancer la capture"),
				),
				ui.Form("wifi-bruteforce",
					[]ui.FormField{
						ui.TextField("capture_path", "Fichier capture").WithPlaceholder("/data/caps/handshake.cap"),
						ui.TextField("wordlist", "Wordlist").WithPlaceholder("/data/wordlists/rockyou.txt"),
						ui.TextField("bssid", "BSSID"),
						ui.TextField("ssid", "SSID"),
					},
					ui.WithSubmitLabel("Bruteforce"),
				),
			),
		))

		// Logs / statut
		view.AddComponent(ui.Card("Tâches en cours",
			ui.List(
				ui.ListItem("Plugin", p.Key()),
				ui.ListItem("Version", p.Version()),
				ui.ListItem("Démarré", p.startedAt.Format(time.RFC3339)),
			),
		))

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

		networks, err := scanWiFi(iface)
		if err != nil {
			return writeJSON(c, 500, map[string]any{"error": err.Error()})
		}

		p.mu.Lock()
		p.lastScan = networks
		p.lastAt = time.Now()
		p.mu.Unlock()

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
		results := append([]WiFiNetwork(nil), p.lastScan...)
		scannedAt := p.lastAt
		p.mu.Unlock()

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
		capture := models.WifiCapture{
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
				p.db.Model(&models.WifiCapture{}).Where("id = ?", captureDBID).Updates(map[string]any{
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
			BSSID     string `json:"bssid"`     // AP cible
			Client    string `json:"client"`    // Client MAC (optionnel, FF:FF:FF:FF:FF:FF = broadcast)
			Count     int    `json:"count"`     // Nombre de paquets (0 = continu)
			Channel   int    `json:"channel"`   // Canal du réseau
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
		var captures []models.WifiCapture
		
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
						p.db.Model(&models.WifiCapture{}).Where("id = ?", cap.ID).Update("has_handshake", true)
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
				p.db.Model(&models.WifiCapture{}).
					Where("capture_path LIKE ?", strings.TrimSuffix(req.CapturePath, ".cap")+"%").
					Updates(map[string]any{
						"cracked":          true,
						"cracked_password": result.Password,
						"cracked_at":       now,
					})
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

	// Wordlists disponibles
	group.GET("/wordlists", func(c buffalo.Context) error {
		wordlistDirs := []string{
			"/opt/heimdall/wordlists",
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
				path := filepath.Join(dir, f.Name())
				
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

				wordlists = append(wordlists, map[string]any{
					"path":  path,
					"name":  f.Name(),
					"size":  info.Size(),
					"lines": lineCount,
				})
			}
		}

		return writeJSON(c, 200, map[string]any{
			"wordlists": wordlists,
			"count":     len(wordlists),
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

		// Format: BSSID(1):SSID(1):MODE(1):CHAN(1):RATE(1):SIGNAL(1):SECURITY(1+)
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
