package wifi

import (
	"encoding/json"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/nxo/engine/internal/plugins"
	"github.com/nxo/engine/internal/ui"
)

// WiFiPlugin expose les endpoints de pilotage Wi-Fi.
type WiFiPlugin struct {
	startedAt time.Time
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

	// Stub: scan
	group.POST("/scan", func(c buffalo.Context) error {
		return writeJSON(c, 200, map[string]any{"status": "scheduled", "message": "Scan en attente (stub)"})
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
