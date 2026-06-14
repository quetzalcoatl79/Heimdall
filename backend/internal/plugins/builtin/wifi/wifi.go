package wifi

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/nxo/engine/internal/plugins"
	"github.com/nxo/engine/internal/ui"
)

// WiFiState représente l'état actuel des interfaces WiFi
type WiFiState struct {
	Interface     string    `json:"interface"`
	OriginalMode  string    `json:"original_mode"` // "managed" ou "monitor"
	CurrentMode   string    `json:"current_mode"`
	IsDisabled    bool      `json:"is_disabled"`
	DisabledAt    time.Time `json:"disabled_at,omitempty"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	SessionID     string    `json:"session_id"`
}

// WiFiPlugin expose les endpoints de pilotage Wi-Fi.
type WiFiPlugin struct {
	startedAt time.Time

	// Gestion de l'état des interfaces
	mu              sync.RWMutex
	interfaceStates map[string]*WiFiState
	activeSession   string
	heartbeatTicker *time.Ticker
	stopHeartbeat   chan struct{}
}

func (p *WiFiPlugin) Key() string         { return "wifi" }
func (p *WiFiPlugin) Version() string     { return "0.2.0" }
func (p *WiFiPlugin) Description() string { return "Pentest Wi-Fi (scan, capture, bruteforce)" }

// restoreInterface remet une interface en mode managed
func (p *WiFiPlugin) restoreInterface(ifaceName string) error {
	p.mu.Lock()
	state, exists := p.interfaceStates[ifaceName]
	if !exists {
		p.mu.Unlock()
		return fmt.Errorf("interface %s not tracked", ifaceName)
	}
	p.mu.Unlock()

	// Si l'interface était en mode monitor, la remettre en managed
	if state.CurrentMode == "monitor" {
		// Arrêter le mode monitor avec airmon-ng
		if err := exec.Command("airmon-ng", "stop", ifaceName).Run(); err != nil {
			// Fallback: essayer avec iw/ifconfig
			exec.Command("ip", "link", "set", ifaceName, "down").Run()
			exec.Command("iw", ifaceName, "set", "type", "managed").Run()
			exec.Command("ip", "link", "set", ifaceName, "up").Run()
		}
	}

	// Réactiver l'interface si elle était désactivée
	if state.IsDisabled {
		exec.Command("ip", "link", "set", ifaceName, "up").Run()
		exec.Command("rfkill", "unblock", "wifi").Run()

		// Redémarrer NetworkManager si disponible
		exec.Command("systemctl", "restart", "NetworkManager").Run()
	}

	p.mu.Lock()
	state.CurrentMode = "managed"
	state.IsDisabled = false
	p.mu.Unlock()

	return nil
}

// RestoreAllInterfaces restaure toutes les interfaces trackées
func (p *WiFiPlugin) RestoreAllInterfaces() error {
	p.mu.RLock()
	interfaces := make([]string, 0, len(p.interfaceStates))
	for iface := range p.interfaceStates {
		interfaces = append(interfaces, iface)
	}
	p.mu.RUnlock()

	var lastErr error
	for _, iface := range interfaces {
		if err := p.restoreInterface(iface); err != nil {
			lastErr = err
		}
	}

	// Nettoyer l'état
	p.mu.Lock()
	p.interfaceStates = make(map[string]*WiFiState)
	p.activeSession = ""
	p.mu.Unlock()

	return lastErr
}

// startHeartbeatMonitor démarre la surveillance des heartbeats
func (p *WiFiPlugin) startHeartbeatMonitor() {
	p.heartbeatTicker = time.NewTicker(10 * time.Second)
	p.stopHeartbeat = make(chan struct{})

	go func() {
		for {
			select {
			case <-p.heartbeatTicker.C:
				p.checkHeartbeats()
			case <-p.stopHeartbeat:
				p.heartbeatTicker.Stop()
				return
			}
		}
	}()
}

// checkHeartbeats vérifie si le client est toujours connecté
func (p *WiFiPlugin) checkHeartbeats() {
	p.mu.RLock()
	var staleSessions []string
	timeout := 30 * time.Second // Si pas de heartbeat depuis 30s, restaurer

	for iface, state := range p.interfaceStates {
		if state.IsDisabled && time.Since(state.LastHeartbeat) > timeout {
			staleSessions = append(staleSessions, iface)
		}
	}
	p.mu.RUnlock()

	// Restaurer les interfaces avec des sessions expirées
	for _, iface := range staleSessions {
		fmt.Printf("[WiFi Plugin] Session timeout for %s, restoring interface...\n", iface)
		p.restoreInterface(iface)
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
	// Démarrer le monitoring des heartbeats
	p.startHeartbeatMonitor()

	// UI schema
	group.GET("/view", func(c buffalo.Context) error {
		// Vérifier si des interfaces sont en mode pentest
		p.mu.RLock()
		hasDisabledInterfaces := len(p.interfaceStates) > 0
		disabledList := make([]map[string]any, 0)
		for _, state := range p.interfaceStates {
			if state.IsDisabled || state.CurrentMode == "monitor" {
				disabledList = append(disabledList, map[string]any{
					"interface":    state.Interface,
					"current_mode": state.CurrentMode,
					"disabled_at":  state.DisabledAt.Format(time.RFC3339),
				})
			}
		}
		p.mu.RUnlock()

		view := ui.NewView("Pentest Wi-Fi").
			WithDescription("Scan, capture de handshakes et bruteforce WPA/WEP").
			WithIcon("wifi").
			WithRefresh(5)

		// Actions globales
		view.AddAction(ui.Action{ID: "refresh", Label: "Rafraîchir", Icon: "refresh", Variant: "secondary"})
		view.AddAction(ui.Action{ID: "scan", Label: "Scanner", Icon: "radar", Variant: "primary"})

		// Bouton de restauration WiFi (visible uniquement si interfaces désactivées)
		if hasDisabledInterfaces {
			view.AddAction(ui.Action{
				ID:      "restore-wifi",
				Label:   "🔄 Restaurer WiFi",
				Icon:    "wifi",
				Variant: "danger",
			})
		}

		// Alerte si WiFi désactivé
		if hasDisabledInterfaces {
			view.AddComponent(ui.Alert("warning",
				"⚠️ Mode Pentest actif - Le WiFi standard est désactivé. "+
					"Cliquez sur 'Restaurer WiFi' pour le réactiver.",
			))
		}

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

		// Tableau des réseaux détectés
		cols := []ui.TableColumn{
			{Key: "selected", Label: ""},
			{Key: "ssid", Label: "SSID", Sortable: true},
			{Key: "bssid", Label: "BSSID"},
			{Key: "channel", Label: "Canal"},
			{Key: "signal", Label: "Signal"},
			{Key: "security", Label: "Séc."},
			{Key: "vendor", Label: "Vendor"},
			{Key: "wps", Label: "WPS"},
		}
		view.AddComponent(ui.Card("Réseaux Wi-Fi détectés",
			ui.Table(cols, []map[string]any{}),
		))

		// Actions de capture
		view.AddComponent(ui.Card("Capture",
			ui.Grid(2,
				ui.Form("wifi-capture",
					[]ui.FormField{
						ui.TextField("targets", "BSSIDs ciblés").WithPlaceholder("bssid1,bssid2"),
						ui.TextField("channel", "Canal optionnel"),
						ui.SelectField("mode", "Mode", []ui.SelectOption{
							{Value: "wpa", Label: "WPA/WPA2 Handshake"},
							{Value: "wep", Label: "WEP IVs"},
						}),
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

	// État des interfaces WiFi
	group.GET("/state", func(c buffalo.Context) error {
		p.mu.RLock()
		defer p.mu.RUnlock()

		states := make([]map[string]any, 0, len(p.interfaceStates))
		for _, state := range p.interfaceStates {
			states = append(states, map[string]any{
				"interface":     state.Interface,
				"original_mode": state.OriginalMode,
				"current_mode":  state.CurrentMode,
				"is_disabled":   state.IsDisabled,
				"disabled_at":   state.DisabledAt,
				"session_id":    state.SessionID,
			})
		}

		return writeJSON(c, 200, map[string]any{
			"states":         states,
			"active_session": p.activeSession,
			"has_disabled":   len(states) > 0,
		})
	})

	// Heartbeat - Le frontend doit appeler cette route régulièrement
	group.POST("/heartbeat", func(c buffalo.Context) error {
		var req struct {
			SessionID string `json:"session_id"`
		}
		if err := c.Bind(&req); err != nil {
			return writeJSON(c, 400, map[string]any{"error": "invalid request"})
		}

		p.mu.Lock()
		defer p.mu.Unlock()

		// Mettre à jour le heartbeat pour toutes les interfaces de cette session
		for _, state := range p.interfaceStates {
			if state.SessionID == req.SessionID {
				state.LastHeartbeat = time.Now()
			}
		}

		return writeJSON(c, 200, map[string]any{
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Restaurer une interface spécifique
	group.POST("/restore", func(c buffalo.Context) error {
		var req struct {
			Interface string `json:"interface"`
		}
		if err := c.Bind(&req); err != nil {
			return writeJSON(c, 400, map[string]any{"error": "invalid request"})
		}

		if req.Interface == "" {
			// Restaurer toutes les interfaces
			if err := p.RestoreAllInterfaces(); err != nil {
				return writeJSON(c, 500, map[string]any{
					"error":   "failed to restore interfaces",
					"details": err.Error(),
				})
			}
			return writeJSON(c, 200, map[string]any{
				"status":  "restored",
				"message": "Toutes les interfaces WiFi ont été restaurées",
			})
		}

		// Restaurer une interface spécifique
		if err := p.restoreInterface(req.Interface); err != nil {
			return writeJSON(c, 500, map[string]any{
				"error":   "failed to restore interface",
				"details": err.Error(),
			})
		}

		return writeJSON(c, 200, map[string]any{
			"status":    "restored",
			"interface": req.Interface,
			"message":   fmt.Sprintf("Interface %s restaurée", req.Interface),
		})
	})

	// Restaurer toutes les interfaces (endpoint dédié)
	group.POST("/restore-all", func(c buffalo.Context) error {
		if err := p.RestoreAllInterfaces(); err != nil {
			return writeJSON(c, 500, map[string]any{
				"error":   "failed to restore interfaces",
				"details": err.Error(),
			})
		}
		return writeJSON(c, 200, map[string]any{
			"status":  "restored",
			"message": "Toutes les interfaces WiFi ont été restaurées",
		})
	})

	// Endpoint appelé lors de la déconnexion du client (beforeunload)
	group.POST("/disconnect", func(c buffalo.Context) error {
		var req struct {
			SessionID string `json:"session_id"`
		}
		c.Bind(&req) // Ignorer les erreurs car le beacon peut être incomplet

		fmt.Printf("[WiFi Plugin] Client disconnect signal received (session: %s)\n", req.SessionID)

		// Restaurer toutes les interfaces
		if err := p.RestoreAllInterfaces(); err != nil {
			fmt.Printf("[WiFi Plugin] Error restoring interfaces on disconnect: %v\n", err)
		}

		return writeJSON(c, 200, map[string]any{"status": "disconnected"})
	})

	// Stub: scan (avec tracking de l'état)
	group.POST("/scan", func(c buffalo.Context) error {
		var req struct {
			Interface string `json:"interface"`
			SessionID string `json:"session_id"`
		}
		if err := c.Bind(&req); err != nil {
			return writeJSON(c, 400, map[string]any{"error": "invalid request"})
		}

		// Tracker l'interface
		p.mu.Lock()
		p.interfaceStates[req.Interface] = &WiFiState{
			Interface:     req.Interface,
			OriginalMode:  "managed",
			CurrentMode:   "monitor",
			IsDisabled:    true,
			DisabledAt:    time.Now(),
			LastHeartbeat: time.Now(),
			SessionID:     req.SessionID,
		}
		p.activeSession = req.SessionID
		p.mu.Unlock()

		return writeJSON(c, 200, map[string]any{
			"status":     "scheduled",
			"message":    "Scan WiFi démarré",
			"session_id": req.SessionID,
		})
	})

	// Stub: capture
	group.POST("/capture", func(c buffalo.Context) error {
		return writeJSON(c, 200, map[string]any{"status": "scheduled", "message": "Capture en attente (stub)"})
	})

	// Stub: bruteforce
	group.POST("/bruteforce", func(c buffalo.Context) error {
		return writeJSON(c, 200, map[string]any{"status": "scheduled", "message": "Bruteforce en attente (stub)"})
	})
}

func init() {
	p := &WiFiPlugin{
		startedAt:       time.Now(),
		interfaceStates: make(map[string]*WiFiState),
	}
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
		// Filtrer les interfaces WiFi (heuristique basée sur le nom)
		name := strings.ToLower(iface.Name)
		isWiFi := strings.HasPrefix(name, "wlan") ||
			strings.HasPrefix(name, "wlp") ||
			strings.HasPrefix(name, "wifi") ||
			strings.HasPrefix(name, "ath") ||
			strings.HasPrefix(name, "ra") ||
			strings.Contains(name, "wireless") ||
			strings.Contains(name, "wi-fi")

		if !isWiFi {
			continue
		}

		// Vérifier si en mode monitor (Linux: nom finit par "mon")
		isMonitor := strings.HasSuffix(name, "mon")

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
