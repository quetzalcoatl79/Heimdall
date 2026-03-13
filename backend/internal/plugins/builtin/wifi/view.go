package wifi

import (
	"fmt"

	"github.com/nxo/engine/internal/database"
	"github.com/nxo/engine/internal/ui"
)

// ViewBuilder construit les vues du plugin WiFi
type ViewBuilder struct {
	db *database.DB
}

// NewViewBuilder crée un nouveau builder de vues
func NewViewBuilder(db *database.DB) *ViewBuilder {
	return &ViewBuilder{db: db}
}

// BuildMainView construit la vue principale avec tous les onglets
func (vb *ViewBuilder) BuildMainView(
	interfaces []map[string]any,
	networks []WiFiNetwork,
	captures []map[string]any,
	wordlists []map[string]any,
	captureRunning bool,
	bruteRunning bool,
) *ui.ViewSchema {

	view := ui.NewView("Pentest Wi-Fi").
		WithDescription("Scan, capture de handshakes et cracking WPA/WPA2").
		WithIcon("wifi").
		WithRefresh(10)

	// Actions globales
	view.AddAction(ui.Action{ID: "refresh", Label: "Rafraîchir", Icon: "refresh", Variant: "secondary"})

	// Onglets via le composant tabs
	tabs := vb.buildTabs(interfaces, networks, captures, wordlists, captureRunning, bruteRunning)
	view.AddComponent(tabs)

	return view
}

// buildTabs construit le composant onglets
func (vb *ViewBuilder) buildTabs(
	interfaces []map[string]any,
	networks []WiFiNetwork,
	captures []map[string]any,
	wordlists []map[string]any,
	captureRunning bool,
	bruteRunning bool,
) ui.Component {
	return ui.Component{
		Type: "tabs",
		ID:   "wifi-main-tabs",
		Props: map[string]any{
			"defaultTab": "scan",
		},
		Children: []ui.Component{
			vb.buildScanTab(interfaces, networks, captureRunning),
			vb.buildCapturesTab(captures),
			vb.buildCrackTab(captures, wordlists, bruteRunning),
			vb.buildReportsTab(),
		},
	}
}

// buildScanTab construit l'onglet Scan & Capture
func (vb *ViewBuilder) buildScanTab(interfaces []map[string]any, networks []WiFiNetwork, captureRunning bool) ui.Component {
	// Construire les options d'interface
	interfaceOptions := []ui.SelectOption{{Value: "", Label: "Choisir une interface", Disabled: true}}
	for _, iface := range interfaces {
		name := iface["name"].(string)
		label := name
		if monitor, ok := iface["monitor"].(bool); ok && monitor {
			label += " [monitor]"
		}
		if driver, ok := iface["driver"].(string); ok && driver != "" {
			label += fmt.Sprintf(" (%s)", driver)
		}
		interfaceOptions = append(interfaceOptions, ui.SelectOption{Value: name, Label: label})
	}

	// Construire les données du tableau de réseaux
	networkData := make([]map[string]any, len(networks))
	for i, n := range networks {
		networkData[i] = map[string]any{
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

	// Colonnes du tableau
	columns := []ui.TableColumn{
		{Key: "select", Label: "", Width: "40px"},
		{Key: "ssid", Label: "SSID", Sortable: true, Filterable: true},
		{Key: "bssid", Label: "BSSID", Filterable: true},
		{Key: "channel", Label: "Canal", Sortable: true, Filterable: true, FilterType: "select", Align: "center"},
		{Key: "signal", Label: "Signal", Sortable: true, Render: "signal"},
		{Key: "security", Label: "Sécurité", Sortable: true, Filterable: true, FilterType: "select", Render: "badge"},
		{Key: "vendor", Label: "Vendeur", Filterable: true},
		{Key: "actions", Label: "", Width: "100px"},
	}

	// Status de capture
	var captureAlerts []ui.Component
	if captureRunning {
		captureAlerts = append(captureAlerts, ui.Alert("warning", "📡 Capture en cours... Cliquez sur 'Arrêter' pour terminer."))
	}

	return ui.Component{
		Type: "tab",
		ID:   "scan",
		Props: map[string]any{
			"label": "Scan & Capture",
			"icon":  "radio",
		},
		Children: []ui.Component{
			// Sélecteur d'interface + bouton scan
			ui.Card("Interface Wi-Fi",
				ui.Row(
					ui.Col(8,
						ui.Form("wifi-interface-form",
							[]ui.FormField{
								ui.SelectField("interface", "Interface", interfaceOptions).
									WithHelp("Sélectionnez l'interface WiFi à utiliser"),
							},
						),
					),
					ui.Col(4,
						ui.Component{
							Type: "buttonGroup",
							Children: []ui.Component{
								{
									Type: "button",
									ID:   "btn-scan",
									Props: map[string]any{
										"label":    "🔍 Scanner",
										"variant":  "primary",
										"action":   "scan",
										"endpoint": "/plugins/wifi/scan",
										"method":   "POST",
										"formId":   "wifi-interface-form",
									},
								},
								{
									Type: "button",
									ID:   "btn-refresh-interfaces",
									Props: map[string]any{
										"label":    "↻",
										"variant":  "secondary",
										"action":   "refresh",
										"endpoint": "/plugins/wifi/interfaces",
									},
								},
							},
						},
					),
				),
			),

			// Alertes de capture
			ui.Component{
				Type:     "container",
				Children: captureAlerts,
			},

			// Tableau des réseaux
			ui.CardWithProps(map[string]any{
				"title":    fmt.Sprintf("Réseaux détectés (%d)", len(networks)),
				"subtitle": "Sélectionnez les réseaux cibles pour la capture",
			},
				ui.TableWithOptions(columns, networkData,
					ui.TableID("wifi-networks"),
					ui.TableFilterable(),
					ui.TableSearchable(),
					ui.TableSelectable(),
					ui.TablePaginated(15),
					ui.TableRowKey("bssid"),
					ui.TableEmptyMessage("Aucun réseau détecté. Cliquez sur 'Scanner' pour démarrer."),
				),
			),

			// Boutons d'action capture
			ui.Component{
				Type: "actionBar",
				ID:   "capture-actions",
				Props: map[string]any{
					"position": "bottom",
				},
				Children: []ui.Component{
					{
						Type: "button",
						ID:   "btn-capture",
						Props: map[string]any{
							"label":        "▶️ Capturer les réseaux sélectionnés",
							"variant":      "primary",
							"action":       "capture",
							"endpoint":     "/plugins/wifi/capture",
							"method":       "POST",
							"requiresRows": true,
							"disabled":     captureRunning,
						},
					},
					{
						Type:    "button",
						ID:      "btn-stop-capture",
						Visible: boolPtr(captureRunning),
						Props: map[string]any{
							"label":    "⏹️ Arrêter la capture",
							"variant":  "danger",
							"action":   "stopCapture",
							"endpoint": "/plugins/wifi/capture/stop",
							"method":   "POST",
						},
					},
				},
			},
		},
	}
}

// buildCapturesTab construit l'onglet Captures
func (vb *ViewBuilder) buildCapturesTab(captures []map[string]any) ui.Component {
	// Colonnes du tableau
	columns := []ui.TableColumn{
		{Key: "ssid", Label: "SSID", Sortable: true},
		{Key: "bssid", Label: "BSSID"},
		{Key: "channel", Label: "Canal", Align: "center"},
		{Key: "security", Label: "Sécurité", Render: "badge"},
		{Key: "has_handshake", Label: "Handshake", Render: "boolean", Align: "center"},
		{Key: "size", Label: "Taille", Sortable: true},
		{Key: "started_at", Label: "Date", Sortable: true, Render: "datetime"},
		{Key: "status", Label: "Statut", Render: "badge"},
		{Key: "cracked", Label: "Cracké", Render: "boolean", Align: "center"},
		{Key: "password", Label: "Mot de passe"},
		{Key: "actions", Label: "", Width: "120px"},
	}

	// Enrichir les données avec les actions
	for i := range captures {
		captures[i]["actions"] = map[string]any{
			"type": "rowActions",
			"items": []map[string]any{
				{
					"id":       "crack",
					"label":    "Cracker",
					"icon":     "zap",
					"variant":  "primary",
					"disabled": captures[i]["has_handshake"] != true,
					"action":   "goToCrack",
				},
				{
					"id":       "delete",
					"label":    "Supprimer",
					"icon":     "trash",
					"variant":  "danger",
					"confirm":  "Supprimer cette capture ?",
					"endpoint": "/plugins/wifi/captures/{id}",
					"method":   "DELETE",
				},
			},
		}
	}

	return ui.Component{
		Type: "tab",
		ID:   "captures",
		Props: map[string]any{
			"label": fmt.Sprintf("Captures (%d)", len(captures)),
			"icon":  "file-text",
		},
		Children: []ui.Component{
			ui.Card("Fichiers de capture",
				ui.TableWithOptions(columns, captures,
					ui.TableSearchable(),
					ui.TablePaginated(10),
					ui.TableRowKey("id"),
					ui.TableEmptyMessage("Aucune capture. Lancez une capture depuis l'onglet 'Scan & Capture'."),
				),
			),
		},
	}
}

// buildCrackTab construit l'onglet Cracker
func (vb *ViewBuilder) buildCrackTab(captures []map[string]any, wordlists []map[string]any, bruteRunning bool) ui.Component {
	// Options de capture (seulement celles avec handshake)
	captureOptions := []ui.SelectOption{{Value: "", Label: "Sélectionner une capture", Disabled: true}}
	for _, cap := range captures {
		if hasHandshake, ok := cap["has_handshake"].(bool); ok && hasHandshake {
			ssid := cap["ssid"].(string)
			bssid := cap["bssid"].(string)
			path := ""
			if capFile, ok := cap["cap_file"].(string); ok {
				path = capFile
			} else if p, ok := cap["path"].(string); ok {
				path = p
			}
			label := ssid
			if ssid == "" {
				label = "(Caché)"
			}
			label += " - " + bssid
			captureOptions = append(captureOptions, ui.SelectOption{
				Value: path,
				Label: label,
			})
		}
	}

	// Options de wordlist
	wordlistOptions := []ui.SelectOption{{Value: "", Label: "Sélectionner une wordlist", Disabled: true}}
	for _, wl := range wordlists {
		name := wl["name"].(string)
		lines := wl["lines"].(int64)
		label := name
		if lines > 0 {
			label += fmt.Sprintf(" (%.1fM mots)", float64(lines)/1000000)
		}
		wordlistOptions = append(wordlistOptions, ui.SelectOption{
			Value: wl["path"].(string),
			Label: label,
		})
	}

	// ISP options
	ispOptions := []ui.SelectOption{
		{Value: "", Label: "Auto-détection"},
		{Value: "Orange", Label: "🟠 Orange (Livebox) - 26 caractères hex"},
		{Value: "Free", Label: "🔴 Free (Freebox) - Mots latins"},
		{Value: "SFR", Label: "🔵 SFR (Box SFR) - 12 alphanum"},
		{Value: "Bouygues", Label: "🟢 Bouygues (Bbox) - 8-26 alphanum"},
		{Value: "TP-Link", Label: "📶 TP-Link - 8 chiffres"},
		{Value: "Netgear", Label: "📡 Netgear - Adjectif+Nom+3 chiffres"},
	}

	// Alerte si bruteforce en cours
	var bruteAlert ui.Component
	if bruteRunning {
		bruteAlert = ui.Alert("warning", "⚡ Bruteforce en cours...")
	}

	return ui.Component{
		Type: "tab",
		ID:   "crack",
		Props: map[string]any{
			"label": "Cracker",
			"icon":  "key",
			"badge": func() string {
				if bruteRunning {
					return "⏳"
				}
				return ""
			}(),
		},
		Children: []ui.Component{
			bruteAlert,

			ui.Card("Configuration du bruteforce",
				ui.Form("bruteforce-form",
					[]ui.FormField{
						// Mode de bruteforce
						ui.CheckboxField("use_wordlist", "📖 Dictionnaire (Wordlist)").
							WithDefault(true).
							WithHelp("Utilise une liste de mots de passe (rockyou, etc.)"),
						ui.CheckboxField("use_pattern", "🎯 Pattern ISP (Génération)").
							WithDefault(false).
							WithHelp("Génère les mots de passe selon le pattern de l'opérateur"),

						// Capture
						ui.SelectField("capture_path", "Fichier capture", captureOptions).
							WithHelp("Sélectionnez une capture avec handshake").
							WithOnChange("autofillFromCapture"),

						// BSSID et SSID (auto-remplis)
						ui.TextField("bssid", "BSSID").
							WithPlaceholder("00:11:22:33:44:55").
							WithHelp("MAC du point d'accès"),
						ui.TextField("ssid", "SSID").
							WithPlaceholder("MonReseau").
							WithHelp("Nom du réseau"),

						// Wordlist (si use_wordlist)
						ui.SelectField("wordlist", "Wordlist", wordlistOptions).
							WithHelp("Liste de mots de passe à tester").
							WithConditional("use_wordlist", true),

						// ISP (si use_pattern)
						ui.SelectField("isp", "Opérateur (ISP)", ispOptions).
							WithHelp("Le pattern sera utilisé pour générer les mots de passe").
							WithConditional("use_pattern", true),

						// Masque custom
						ui.TextField("custom_mask", "Masque personnalisé").
							WithPlaceholder("?d?d?d?d?d?d?d?d").
							WithHelp("Masque hashcat: ?d=chiffre, ?l=minuscule, ?u=majuscule").
							WithConditional("use_pattern", true),
					},
					ui.WithSubmitLabel("🚀 Lancer le bruteforce"),
					ui.WithSubmitEndpoint("/plugins/wifi/bruteforce/combined"),
					ui.WithSubmitMethod("POST"),
				),
			),

			// Boutons d'action
			ui.Component{
				Type: "actionBar",
				ID:   "brute-actions",
				Props: map[string]any{
					"position": "bottom",
				},
				Children: []ui.Component{
					{
						Type:    "button",
						ID:      "btn-stop-brute",
						Visible: boolPtr(bruteRunning),
						Props: map[string]any{
							"label":    "⏹️ Arrêter le bruteforce",
							"variant":  "danger",
							"endpoint": "/plugins/wifi/bruteforce/stop",
							"method":   "POST",
						},
					},
				},
			},

			// Historique des résultats
			ui.CardWithProps(map[string]any{
				"title": "Historique des attaques",
			},
				ui.TableWithOptions(
					[]ui.TableColumn{
						{Key: "ssid", Label: "SSID"},
						{Key: "bssid", Label: "BSSID"},
						{Key: "success", Label: "Résultat", Render: "boolean"},
						{Key: "password", Label: "Mot de passe"},
						{Key: "wordlist", Label: "Wordlist"},
						{Key: "duration_seconds", Label: "Durée (s)"},
						{Key: "tested_at", Label: "Date", Render: "datetime"},
					},
					[]map[string]any{}, // Données chargées dynamiquement
					ui.TableEmptyMessage("Aucun bruteforce effectué pour le moment."),
				),
			),
		},
	}
}

// buildReportsTab construit l'onglet Rapports
func (vb *ViewBuilder) buildReportsTab() ui.Component {
	return ui.Component{
		Type: "tab",
		ID:   "reports",
		Props: map[string]any{
			"label": "Rapports",
			"icon":  "clipboard-list",
		},
		Children: []ui.Component{
			// Header avec bouton nouveau
			ui.Row(
				ui.Col(8,
					ui.Heading(3, "Audits & Rapports"),
					ui.Text("Gérez vos audits WiFi et générez des rapports PDF professionnels"),
				),
				ui.Col(4,
					ui.Component{
						Type: "button",
						ID:   "btn-new-audit",
						Props: map[string]any{
							"label":   "+ Nouvel audit",
							"variant": "primary",
							"action":  "openModal",
							"modal":   "new-audit-modal",
						},
					},
				),
			),

			// Liste des audits
			ui.CardWithProps(map[string]any{
				"title": "Audits récents",
			},
				ui.TableWithOptions(
					[]ui.TableColumn{
						{Key: "client_name", Label: "Client", Sortable: true},
						{Key: "location", Label: "Lieu"},
						{Key: "status", Label: "Statut", Render: "badge"},
						{Key: "networks_tested", Label: "Réseaux", Align: "center"},
						{Key: "passwords_cracked", Label: "Crackés", Align: "center"},
						{Key: "start_date", Label: "Date", Render: "datetime"},
						{Key: "report_generated", Label: "Rapport", Render: "boolean"},
					},
					[]map[string]any{}, // Données chargées dynamiquement
					ui.TableEmptyMessage("Aucun audit. Créez-en un avec le bouton ci-dessus."),
				),
			),

			// Modal nouveau audit
			ui.Component{
				Type: "modal",
				ID:   "new-audit-modal",
				Props: map[string]any{
					"title": "Nouvel audit WiFi",
					"icon":  "clipboard-list",
				},
				Children: []ui.Component{
					ui.Form("new-audit-form",
						[]ui.FormField{
							ui.TextField("client_name", "Nom du client / entreprise").
								WithRequired().
								WithPlaceholder("Ex: Acme Corp"),
							ui.TextField("client_contact", "Contact").
								WithPlaceholder("Email ou téléphone"),
							ui.TextField("location", "Lieu du test").
								WithPlaceholder("Ex: Siège social Paris"),
							ui.TextField("tester_name", "Nom du testeur").
								WithPlaceholder("Votre nom"),
							ui.TextareaField("notes", "Notes").
								WithPlaceholder("Contexte, objectifs, périmètre..."),
						},
						ui.WithSubmitLabel("Créer l'audit"),
						ui.WithSubmitEndpoint("/plugins/wifi/audits"),
						ui.WithSubmitMethod("POST"),
						ui.WithCloseOnSubmit(true),
						ui.WithRefreshOn("audit-created"),
					),
				},
			},
		},
	}
}

// Helper function
func boolPtr(b bool) *bool {
	return &b
}

// GetCapturesForView récupère les captures formatées pour la vue
func (vb *ViewBuilder) GetCapturesForView() []map[string]any {
	if vb.db == nil {
		return []map[string]any{}
	}

	var captures []WifiCapture
	vb.db.Order("created_at DESC").Find(&captures)

	result := make([]map[string]any, len(captures))
	for i, cap := range captures {
		result[i] = map[string]any{
			"id":            cap.ID.String(),
			"ssid":          cap.SSID,
			"bssid":         cap.BSSID,
			"channel":       cap.Channel,
			"security":      cap.Security,
			"path":          cap.CapturePath,
			"name":          cap.CaptureName,
			"has_handshake": cap.HasHandshake,
			"size":          cap.FileSize,
			"started_at":    cap.StartedAt,
			"status":        cap.Status,
			"cracked":       cap.Cracked,
			"password":      cap.CrackedPassword,
		}
	}
	return result
}

// GetAuditsForView récupère les audits formatés pour la vue
func (vb *ViewBuilder) GetAuditsForView() []map[string]any {
	if vb.db == nil {
		return []map[string]any{}
	}

	var audits []WifiAudit
	vb.db.Order("created_at DESC").Find(&audits)

	result := make([]map[string]any, len(audits))
	for i, audit := range audits {
		result[i] = map[string]any{
			"id":                  audit.ID.String(),
			"client_name":         audit.ClientName,
			"client_contact":      audit.ClientContact,
			"location":            audit.Location,
			"tester_name":         audit.TesterName,
			"start_date":          audit.StartDate,
			"end_date":            audit.EndDate,
			"status":              audit.Status,
			"networks_tested":     audit.NetworksTested,
			"handshakes_captured": audit.HandshakesCaptured,
			"passwords_cracked":   audit.PasswordsCracked,
			"report_generated":    audit.ReportGenerated,
			"report_path":         audit.ReportPath,
		}
	}
	return result
}
