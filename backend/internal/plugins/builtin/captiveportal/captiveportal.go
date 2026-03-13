package captiveportal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/nxo/engine/internal/plugins"
	"github.com/nxo/engine/internal/ui"
)

// writeJSON helper for HTTP responses
func writeJSON(c buffalo.Context, status int, payload any) error {
	c.Response().Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Response().WriteHeader(status)
	_ = json.NewEncoder(c.Response()).Encode(payload)
	return nil
}

// CaptivePortalPlugin implements the Plugin interface
var _ plugins.Plugin = (*CaptivePortalPlugin)(nil)

type CaptivePortalPlugin struct{}

type Attempt struct {
	Timestamp time.Time `json:"timestamp"`
	TargetURL string    `json:"target_url"`
	Code      string    `json:"code"`
	Login     string    `json:"login"`
	Password  string    `json:"password"`
	Result    string    `json:"result"`
	Status    int       `json:"status,omitempty"`
}

var (
	attemptsMu       sync.Mutex
	attempts         []Attempt
	lastAnalysis     *FormAnalysis
	currentReport    *CaptivePortalReport
	bruteforceActive bool
)

func init() {
	plugins.Register(&CaptivePortalPlugin{})
}

func (p *CaptivePortalPlugin) Key() string     { return "captiveportal" }
func (p *CaptivePortalPlugin) Version() string { return "2.0.0" }
func (p *CaptivePortalPlugin) Description() string {
	return "Pentest des portails captifs avec analyse, bruteforce avancé et rapports"
}

func (p *CaptivePortalPlugin) Manifest() map[string]any {
	return map[string]any{
		"menu_items": []map[string]any{
			{"label": "Captive Portal", "icon": "wifi", "route": "/admin/plugins/captiveportal"},
		},
	}
}

func (p *CaptivePortalPlugin) RegisterRoutes(app *buffalo.App, deps plugins.Deps) {
	// app is already the group /api/v1/plugins/captiveportal from mount.go
	app.GET("/view", p.View)
	app.POST("/analyze", p.Analyze)
	app.POST("/test", p.Test)
	app.POST("/bruteforce/start", p.StartBruteforce)
	app.POST("/bruteforce/stop", p.StopBruteforce)
	app.GET("/bruteforce/status", p.BruteforceStatus)
	app.POST("/bruteforce/estimate", p.EstimateBruteforce)
	app.POST("/report/generate", p.GenerateReport)
	app.GET("/report/download", p.DownloadReport)
	app.DELETE("/history/clear", p.ClearHistory)
}

// View: UI schema for the plugin
func (p *CaptivePortalPlugin) View(c buffalo.Context) error {
	view := ui.NewView("Captive Portal Pentest").WithIcon("wifi").WithRefresh(5)

	// Générer dynamiquement la liste des wordlists intégrées
	wordlistsDir := getWordlistsDir()
	wordlistOptions := []ui.SelectOption{{Label: "Aucune (générer)", Value: ""}}
	if files, err := os.ReadDir(wordlistsDir); err == nil {
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".txt") {
				name := strings.TrimSuffix(f.Name(), ".txt")
				wordlistOptions = append(wordlistOptions, ui.SelectOption{Label: name, Value: name})
			}
		}
	}

	// === SECTION 1: Analyse du portail ===
	analyzeFields := []ui.FormField{
		{Name: "target_url", Label: "URL du portail captif", Type: "text", Required: true, Placeholder: "http://192.168.1.1/login"},
	}
	view.AddComponent(ui.Card("1. Analyser le Portail",
		ui.Form("analyze-form", analyzeFields,
			ui.WithSubmitURL("/api/v1/plugins/captiveportal/analyze", "POST"),
			ui.WithSubmitLabel("Analyser la page"),
		),
	))

	// Afficher l'analyse si disponible
	if lastAnalysis != nil && lastAnalysis.Error == "" {
		analysisData := []map[string]any{}
		for _, field := range lastAnalysis.Fields {
			if field.Type != "hidden" {
				analysisData = append(analysisData, map[string]any{
					"name":      field.Name,
					"type":      field.Type,
					"required":  field.Required,
					"minLength": field.MinLength,
					"maxLength": field.MaxLength,
					"pattern":   field.Pattern,
				})
			}
		}
		columns := []ui.TableColumn{
			{Key: "name", Label: "Champ"},
			{Key: "type", Label: "Type"},
			{Key: "required", Label: "Requis", Render: "boolean"},
			{Key: "minLength", Label: "Min"},
			{Key: "maxLength", Label: "Max"},
			{Key: "pattern", Label: "Pattern"},
		}

		infoItems := []ui.Component{
			ui.Text(fmt.Sprintf("URL: %s", lastAnalysis.URL)),
			ui.Text(fmt.Sprintf("Titre: %s", lastAnalysis.Title)),
			ui.Text(fmt.Sprintf("Méthode: %s → %s", lastAnalysis.FormMethod, lastAnalysis.FormAction)),
			ui.Text(fmt.Sprintf("CSRF: %v | Captcha: %v", lastAnalysis.HasCSRF, lastAnalysis.HasCaptcha)),
		}

		view.AddComponent(ui.Card("Résultat de l'analyse",
			append(infoItems, ui.Table(columns, analysisData))...,
		))
	}

	// === SECTION 2: Test manuel ===
	testFields := []ui.FormField{
		{Name: "target_url", Label: "URL du portail", Type: "text", Required: true},
		{Name: "method", Label: "Méthode HTTP", Type: "select", Default: "POST", Options: []ui.SelectOption{
			{Label: "POST (JSON)", Value: "POST"},
			{Label: "GET", Value: "GET"},
			{Label: "POST (form-data)", Value: "FORM"},
		}},
		{Name: "field_name", Label: "Champ à tester", Type: "select", Default: "code", Options: []ui.SelectOption{
			{Label: "Code d'accès", Value: "code"},
			{Label: "Login", Value: "login"},
			{Label: "Mot de passe", Value: "password"},
		}},
		{Name: "code", Label: "Valeur à tester", Type: "text"},
		{Name: "other_login", Label: "Login fixe (si bruteforce sur password)", Type: "text"},
		{Name: "other_password", Label: "Password fixe (si bruteforce sur login)", Type: "password"},
	}
	view.AddComponent(ui.Card("2. Test Manuel",
		ui.Form("test-form", testFields,
			ui.WithSubmitURL("/api/v1/plugins/captiveportal/test", "POST"),
			ui.WithSubmitLabel("Tester"),
		),
	))

	// === SECTION 3: Configuration Bruteforce ===
	bruteFields := []ui.FormField{
		{Name: "target_url", Label: "URL cible", Type: "text", Required: true},
		{Name: "field_name", Label: "Champ à bruteforcer", Type: "select", Default: "code", Options: []ui.SelectOption{
			{Label: "Code d'accès", Value: "code"},
			{Label: "Login", Value: "login"},
			{Label: "Mot de passe", Value: "password"},
		}},
		{Name: "min_length", Label: "Longueur min", Type: "number", Default: 4},
		{Name: "max_length", Label: "Longueur max", Type: "number", Default: 6},
		{Name: "use_digits", Label: "Chiffres (0-9)", Type: "checkbox", Default: true},
		{Name: "use_lowercase", Label: "Minuscules (a-z)", Type: "checkbox"},
		{Name: "use_uppercase", Label: "Majuscules (A-Z)", Type: "checkbox"},
		{Name: "use_special", Label: "Spéciaux (!@#$...)", Type: "checkbox"},
		{Name: "custom_charset", Label: "Charset personnalisé", Type: "text", Placeholder: "Ex: ABC123"},
		{Name: "rate_limit", Label: "Requêtes/seconde", Type: "number", Default: 2},
		{Name: "max_attempts", Label: "Max tentatives (0=illimité)", Type: "number", Default: 0},
		{Name: "wordlist_builtin", Label: "Ou utiliser une wordlist", Type: "select", Options: wordlistOptions},
		{Name: "success_pattern", Label: "Pattern succès (regex)", Type: "text", Placeholder: "success|bienvenue|welcome"},
		{Name: "failure_pattern", Label: "Pattern échec (regex)", Type: "text", Placeholder: "error|invalide|incorrect"},
	}
	view.AddComponent(ui.Card("3. Configuration Bruteforce",
		ui.Form("bruteforce-form", bruteFields,
			ui.WithSubmitURL("/api/v1/plugins/captiveportal/bruteforce/start", "POST"),
			ui.WithSubmitLabel("Lancer le Bruteforce"),
		),
	))

	// Progress du bruteforce
	progress := GetBruteforceProgress()
	if progress.Running || progress.Attempted > 0 {
		progressData := []map[string]any{
			{
				"status":    map[bool]string{true: "En cours", false: "Terminé"}[progress.Running],
				"attempted": progress.Attempted,
				"total":     progress.TotalPossible,
				"found":     len(progress.FoundCodes),
				"speed":     fmt.Sprintf("%.1f/s", progress.Speed),
				"eta":       progress.ETA,
				"current":   progress.CurrentValue,
			},
		}
		progressColumns := []ui.TableColumn{
			{Key: "status", Label: "Statut", Render: "badge"},
			{Key: "attempted", Label: "Tentatives"},
			{Key: "total", Label: "Total"},
			{Key: "found", Label: "Trouvés"},
			{Key: "speed", Label: "Vitesse"},
			{Key: "eta", Label: "ETA"},
			{Key: "current", Label: "Valeur actuelle"},
		}
		view.AddComponent(ui.Card("Progression Bruteforce",
			ui.Table(progressColumns, progressData),
		))

		if len(progress.FoundCodes) > 0 {
			codesData := make([]map[string]any, len(progress.FoundCodes))
			for i, code := range progress.FoundCodes {
				codesData[i] = map[string]any{"code": code}
			}
			view.AddComponent(ui.Card("Codes/Mots de passe trouvés",
				ui.Table([]ui.TableColumn{{Key: "code", Label: "Valeur"}}, codesData),
			))
		}
	}

	// === SECTION 4: Historique ===
	attemptsMu.Lock()
	history := make([]map[string]any, len(attempts))
	for i, a := range attempts {
		history[i] = map[string]any{
			"timestamp":  a.Timestamp,
			"target_url": a.TargetURL,
			"code":       a.Code,
			"login":      a.Login,
			"password":   a.Password,
			"result":     a.Result,
		}
	}
	attemptsMu.Unlock()

	historyColumns := []ui.TableColumn{
		{Key: "timestamp", Label: "Date", Render: "datetime"},
		{Key: "target_url", Label: "URL"},
		{Key: "code", Label: "Code"},
		{Key: "login", Label: "Login"},
		{Key: "password", Label: "MdP"},
		{Key: "result", Label: "Résultat", Render: "badge"},
	}

	view.AddComponent(ui.Card("4. Historique des tentatives", ui.Table(historyColumns, history)))

	// === SECTION 5: Génération de rapport ===
	reportFields := []ui.FormField{
		{Name: "client_name", Label: "Nom du client", Type: "text", Required: true},
		{Name: "tester_name", Label: "Nom du testeur", Type: "text"},
		{Name: "location", Label: "Lieu (hôtel, café...)", Type: "text"},
		{Name: "portal_type", Label: "Type de portail", Type: "select", Options: []ui.SelectOption{
			{Label: "Hôtel", Value: "hotel"},
			{Label: "Café/Restaurant", Value: "cafe"},
			{Label: "Aéroport", Value: "airport"},
			{Label: "Transport", Value: "transport"},
			{Label: "Entreprise", Value: "enterprise"},
			{Label: "Autre", Value: "other"},
		}},
	}
	view.AddComponent(ui.Card("5. Générer un Rapport",
		ui.Form("report-form", reportFields,
			ui.WithSubmitURL("/api/v1/plugins/captiveportal/report/generate", "POST"),
			ui.WithSubmitLabel("Générer le rapport PDF"),
		),
	))

	return writeJSON(c, 200, view)
}

// Analyze: analyser la page du portail
func (p *CaptivePortalPlugin) Analyze(c buffalo.Context) error {
	var req struct {
		TargetURL string `json:"target_url"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		req.TargetURL = c.Request().FormValue("target_url")
	}

	if req.TargetURL == "" {
		return writeJSON(c, 400, map[string]any{"error": "URL requise"})
	}

	analysis := AnalyzePortalPage(req.TargetURL)
	lastAnalysis = &analysis

	if analysis.Error != "" {
		return writeJSON(c, 400, map[string]any{"error": analysis.Error})
	}

	return writeJSON(c, 200, analysis)
}

// Test: test manuel d'une valeur
func (p *CaptivePortalPlugin) Test(c buffalo.Context) error {
	targetURL := c.Request().FormValue("target_url")
	method := c.Request().FormValue("method")
	fieldName := c.Request().FormValue("field_name")
	code := c.Request().FormValue("code")
	otherLogin := c.Request().FormValue("other_login")
	otherPassword := c.Request().FormValue("other_password")

	if targetURL == "" {
		return writeJSON(c, 400, map[string]any{"error": "URL requise"})
	}

	var testCode, testLogin, testPassword string
	switch fieldName {
	case "code":
		testCode = code
	case "login":
		testLogin = code
		testPassword = otherPassword
	case "password":
		testLogin = otherLogin
		testPassword = code
	default:
		testCode = code
	}

	att, respBody := doPortalRequest(targetURL, method, nil, testCode, testLogin, testPassword)
	attemptsMu.Lock()
	attempts = append(attempts, att)
	attemptsMu.Unlock()

	return writeJSON(c, 200, map[string]any{
		"status":   att.Result,
		"response": respBody,
		"attempt":  att,
	})
}

// StartBruteforce: démarrer une attaque bruteforce
func (p *CaptivePortalPlugin) StartBruteforce(c buffalo.Context) error {
	if bruteforceActive {
		return writeJSON(c, 409, map[string]any{"error": "Un bruteforce est déjà en cours"})
	}

	targetURL := c.Request().FormValue("target_url")
	fieldName := c.Request().FormValue("field_name")
	minLen, _ := strconv.Atoi(c.Request().FormValue("min_length"))
	maxLen, _ := strconv.Atoi(c.Request().FormValue("max_length"))
	rateLimit, _ := strconv.Atoi(c.Request().FormValue("rate_limit"))
	maxAttempts, _ := strconv.Atoi(c.Request().FormValue("max_attempts"))
	useDigits := c.Request().FormValue("use_digits") == "on" || c.Request().FormValue("use_digits") == "true"
	useLower := c.Request().FormValue("use_lowercase") == "on" || c.Request().FormValue("use_lowercase") == "true"
	useUpper := c.Request().FormValue("use_uppercase") == "on" || c.Request().FormValue("use_uppercase") == "true"
	useSpecial := c.Request().FormValue("use_special") == "on" || c.Request().FormValue("use_special") == "true"
	customCharset := c.Request().FormValue("custom_charset")
	wordlistBuiltin := c.Request().FormValue("wordlist_builtin")
	successPattern := c.Request().FormValue("success_pattern")
	failurePattern := c.Request().FormValue("failure_pattern")

	if targetURL == "" {
		return writeJSON(c, 400, map[string]any{"error": "URL requise"})
	}

	if minLen <= 0 {
		minLen = 4
	}
	if maxLen <= 0 || maxLen < minLen {
		maxLen = minLen + 2
	}
	if rateLimit <= 0 {
		rateLimit = 2
	}

	config := BruteforceConfig{
		TargetURL:      targetURL,
		FieldName:      fieldName,
		MinLength:      minLen,
		MaxLength:      maxLen,
		UseDigits:      useDigits,
		UseLowercase:   useLower,
		UseUppercase:   useUpper,
		UseSpecial:     useSpecial,
		CustomCharset:  customCharset,
		RateLimit:      rateLimit,
		MaxAttempts:    maxAttempts,
		SuccessPattern: successPattern,
		FailurePattern: failurePattern,
	}

	var wordlist []string
	if wordlistBuiltin != "" {
		path := getWordlistsDir() + wordlistBuiltin + ".txt"
		if data, err := ioutil.ReadFile(path); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					wordlist = append(wordlist, line)
				}
			}
		}
	}

	bruteCancel = make(chan struct{})
	bruteCancelOnce = sync.Once{}
	bruteforceActive = true

	go runBruteforce(config, wordlist)

	return writeJSON(c, 200, map[string]any{
		"message": "Bruteforce démarré",
		"config":  config,
		"total":   CalculateTotalCombinations(config),
	})
}

func runBruteforce(config BruteforceConfig, wordlist []string) {
	defer func() {
		bruteforceActive = false
		bruteProgress.mu.Lock()
		bruteProgress.Running = false
		bruteProgress.mu.Unlock()
	}()

	bruteProgress.mu.Lock()
	bruteProgress.Running = true
	bruteProgress.StartedAt = time.Now()
	bruteProgress.Attempted = 0
	bruteProgress.Successful = 0
	bruteProgress.Failed = 0
	bruteProgress.FoundCodes = nil
	bruteProgress.TotalPossible = CalculateTotalCombinations(config)
	if len(wordlist) > 0 {
		bruteProgress.TotalPossible = int64(len(wordlist))
	}
	bruteProgress.mu.Unlock()

	var successRe, failureRe *regexp.Regexp
	if config.SuccessPattern != "" {
		successRe, _ = regexp.Compile("(?i)" + config.SuccessPattern)
	}
	if config.FailurePattern != "" {
		failureRe, _ = regexp.Compile("(?i)" + config.FailurePattern)
	}

	testValue := func(value string) bool {
		select {
		case <-bruteCancel:
			return false
		default:
		}

		bruteProgress.mu.Lock()
		bruteProgress.CurrentValue = value
		bruteProgress.Attempted++
		elapsed := time.Since(bruteProgress.StartedAt).Seconds()
		if elapsed > 0 {
			bruteProgress.Speed = float64(bruteProgress.Attempted) / elapsed
		}
		bruteProgress.mu.Unlock()

		var code, login, password string
		switch config.FieldName {
		case "code":
			code = value
		case "login":
			login = value
		case "password":
			password = value
		default:
			code = value
		}

		att, respBody := doPortalRequest(config.TargetURL, "POST", config.Headers, code, login, password)

		isSuccess := false
		if successRe != nil && successRe.MatchString(respBody) {
			isSuccess = true
		} else if failureRe != nil && !failureRe.MatchString(respBody) {
			isSuccess = true
		} else if att.Status >= 200 && att.Status < 300 && att.Result == "Succès" {
			isSuccess = true
		}

		if isSuccess {
			bruteProgress.mu.Lock()
			bruteProgress.Successful++
			bruteProgress.FoundCodes = append(bruteProgress.FoundCodes, value)
			bruteProgress.mu.Unlock()
			att.Result = "Succès"
		} else {
			bruteProgress.mu.Lock()
			bruteProgress.Failed++
			bruteProgress.mu.Unlock()
		}

		attemptsMu.Lock()
		attempts = append(attempts, att)
		attemptsMu.Unlock()

		time.Sleep(time.Duration(1000/config.RateLimit) * time.Millisecond)

		if config.MaxAttempts > 0 {
			bruteProgress.mu.Lock()
			attempted := bruteProgress.Attempted
			bruteProgress.mu.Unlock()
			if int(attempted) >= config.MaxAttempts {
				return false
			}
		}

		return true
	}

	if len(wordlist) > 0 {
		for _, val := range wordlist {
			if !testValue(val) {
				break
			}
		}
	} else {
		GenerateCombinations(config, testValue)
	}
}

func (p *CaptivePortalPlugin) StopBruteforce(c buffalo.Context) error {
	StopBruteforce()
	bruteforceActive = false
	return writeJSON(c, 200, map[string]any{"message": "Bruteforce arrêté"})
}

func (p *CaptivePortalPlugin) BruteforceStatus(c buffalo.Context) error {
	return writeJSON(c, 200, GetBruteforceProgress())
}

func (p *CaptivePortalPlugin) EstimateBruteforce(c buffalo.Context) error {
	minLen, _ := strconv.Atoi(c.Request().FormValue("min_length"))
	maxLen, _ := strconv.Atoi(c.Request().FormValue("max_length"))
	rateLimit, _ := strconv.Atoi(c.Request().FormValue("rate_limit"))
	useDigits := c.Request().FormValue("use_digits") == "on" || c.Request().FormValue("use_digits") == "true"
	useLower := c.Request().FormValue("use_lowercase") == "on" || c.Request().FormValue("use_lowercase") == "true"
	useUpper := c.Request().FormValue("use_uppercase") == "on" || c.Request().FormValue("use_uppercase") == "true"
	useSpecial := c.Request().FormValue("use_special") == "on" || c.Request().FormValue("use_special") == "true"
	customCharset := c.Request().FormValue("custom_charset")

	config := BruteforceConfig{
		MinLength:     minLen,
		MaxLength:     maxLen,
		UseDigits:     useDigits,
		UseLowercase:  useLower,
		UseUppercase:  useUpper,
		UseSpecial:    useSpecial,
		CustomCharset: customCharset,
		RateLimit:     rateLimit,
	}

	return writeJSON(c, 200, EstimateBruteforceDuration(config))
}

func (p *CaptivePortalPlugin) GenerateReport(c buffalo.Context) error {
	clientName := c.Request().FormValue("client_name")
	testerName := c.Request().FormValue("tester_name")
	location := c.Request().FormValue("location")
	portalType := c.Request().FormValue("portal_type")

	if clientName == "" {
		return writeJSON(c, 400, map[string]any{"error": "Nom du client requis"})
	}

	attemptsMu.Lock()
	totalAttempts := len(attempts)
	var successfulCodes []string
	var startTime, endTime time.Time
	for i, att := range attempts {
		if i == 0 {
			startTime = att.Timestamp
		}
		endTime = att.Timestamp
		if att.Result == "Succès" {
			if att.Code != "" {
				successfulCodes = append(successfulCodes, att.Code)
			} else if att.Login != "" {
				successfulCodes = append(successfulCodes, att.Login)
			} else if att.Password != "" {
				successfulCodes = append(successfulCodes, att.Password)
			}
		}
	}
	attemptsMu.Unlock()

	vulnLevel := "Low"
	if len(successfulCodes) > 0 {
		vulnLevel = "Critical"
	} else if lastAnalysis != nil && !lastAnalysis.HasCSRF && !lastAnalysis.HasCaptcha {
		vulnLevel = "High"
	} else if lastAnalysis != nil && !lastAnalysis.HasCaptcha {
		vulnLevel = "Medium"
	}

	findings := []Finding{}
	if len(successfulCodes) > 0 {
		findings = append(findings, Finding{
			Title:       "Codes/mots de passe faibles découverts",
			Severity:    "Critical",
			Description: fmt.Sprintf("%d code(s)/mot(s) de passe valide(s) ont été découvert(s) par bruteforce.", len(successfulCodes)),
			Impact:      "Un attaquant peut accéder au réseau sans autorisation.",
			Remediation: "Implémenter des codes plus complexes, ajouter un rate limiting strict, utiliser des captchas.",
		})
	}
	if lastAnalysis != nil && !lastAnalysis.HasCSRF {
		findings = append(findings, Finding{
			Title:       "Absence de protection CSRF",
			Severity:    "Medium",
			Description: "Le formulaire ne dispose pas de token CSRF.",
			Impact:      "Vulnérabilité aux attaques Cross-Site Request Forgery.",
			Remediation: "Implémenter des tokens CSRF sur tous les formulaires.",
		})
	}
	if lastAnalysis != nil && !lastAnalysis.HasCaptcha {
		findings = append(findings, Finding{
			Title:       "Absence de Captcha",
			Severity:    "High",
			Description: "Aucun mécanisme de captcha n'a été détecté.",
			Impact:      "Facilite les attaques automatisées par bruteforce.",
			Remediation: "Implémenter un captcha (reCAPTCHA, hCaptcha) après plusieurs tentatives échouées.",
		})
	}

	recommendations := []string{
		"Implémenter un rate limiting strict (max 3-5 tentatives par minute)",
		"Ajouter un captcha après 3 tentatives échouées",
		"Utiliser des codes d'accès plus longs (minimum 8 caractères)",
		"Implémenter une protection CSRF sur tous les formulaires",
		"Logger et surveiller les tentatives d'authentification suspectes",
		"Envisager l'authentification à deux facteurs pour les accès sensibles",
	}

	targetURL := ""
	targetName := location
	var analysis FormAnalysis
	if lastAnalysis != nil {
		targetURL = lastAnalysis.URL
		analysis = *lastAnalysis
		if lastAnalysis.Title != "" {
			targetName = lastAnalysis.Title
		}
	}

	report := CaptivePortalReport{
		Config: CaptivePortalReportConfig{
			ClientName: clientName,
			TesterName: testerName,
			Location:   location,
		},
		GeneratedAt:        time.Now(),
		TargetURL:          targetURL,
		TargetName:         targetName,
		PortalType:         portalType,
		Analysis:           analysis,
		TotalAttempts:      totalAttempts,
		SuccessfulCodes:    successfulCodes,
		StartTime:          startTime,
		EndTime:            endTime,
		Duration:           endTime.Sub(startTime).String(),
		VulnerabilityLevel: vulnLevel,
		Findings:           findings,
		Recommendations:    recommendations,
	}

	path, err := GenerateCaptivePortalReport(report)
	if err != nil {
		return writeJSON(c, 500, map[string]any{"error": err.Error()})
	}

	currentReport = &report

	return writeJSON(c, 200, map[string]any{
		"message": "Rapport généré",
		"path":    path,
		"report":  report,
	})
}

func (p *CaptivePortalPlugin) DownloadReport(c buffalo.Context) error {
	reportsDir := "/opt/heimdall/reports/captiveportal"
	files, err := ioutil.ReadDir(reportsDir)
	if err != nil || len(files) == 0 {
		return writeJSON(c, 404, map[string]any{"error": "Aucun rapport disponible"})
	}

	var latest os.FileInfo
	for _, f := range files {
		if latest == nil || f.ModTime().After(latest.ModTime()) {
			latest = f
		}
	}

	if latest == nil {
		return writeJSON(c, 404, map[string]any{"error": "Aucun rapport trouvé"})
	}

	path := reportsDir + "/" + latest.Name()
	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", latest.Name()))
	http.ServeFile(c.Response(), c.Request(), path)
	return nil
}

func (p *CaptivePortalPlugin) ClearHistory(c buffalo.Context) error {
	attemptsMu.Lock()
	attempts = nil
	attemptsMu.Unlock()
	lastAnalysis = nil
	return writeJSON(c, 200, map[string]any{"message": "Historique effacé"})
}

func doPortalRequest(targetURL, method string, headers map[string]string, code, login, password string) (Attempt, string) {
	data := map[string]string{}
	if code != "" {
		data["code"] = code
	}
	if login != "" {
		data["login"] = login
	}
	if password != "" {
		data["password"] = password
	}

	var req *http.Request
	var err error
	if method == "GET" {
		q := targetURL + "?"
		for k, v := range data {
			q += k + "=" + v + "&"
		}
		if len(q) > len(targetURL)+1 {
			q = q[:len(q)-1]
		}
		req, err = http.NewRequest("GET", q, nil)
	} else if method == "FORM" {
		form := ""
		for k, v := range data {
			form += k + "=" + v + "&"
		}
		if len(form) > 0 {
			form = form[:len(form)-1]
		}
		req, err = http.NewRequest("POST", targetURL, bytes.NewBufferString(form))
		if req != nil {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	} else {
		body, _ := json.Marshal(data)
		req, err = http.NewRequest("POST", targetURL, bytes.NewBuffer(body))
		if req != nil {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	if req != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	result := "Erreur"
	status := 0
	respBody := ""
	if err == nil && req != nil {
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err2 := client.Do(req)
		if err2 != nil {
			respBody = err2.Error()
		} else {
			status = resp.StatusCode
			b, _ := ioutil.ReadAll(resp.Body)
			respBody = string(b)
			resp.Body.Close()
			if status >= 200 && status < 300 {
				result = "Succès"
			} else {
				result = "Échec"
			}
		}
	} else if err != nil {
		respBody = err.Error()
	}

	att := Attempt{
		Timestamp: time.Now(),
		TargetURL: targetURL,
		Code:      code,
		Login:     login,
		Password:  password,
		Result:    result,
		Status:    status,
	}
	return att, respBody
}

func getWordlistsDir() string {
	return "/home/quetzalcoalt/Documents/DEV/Heimdall/backend/internal/plugins/builtin/captiveportal/wordlists/"
}
