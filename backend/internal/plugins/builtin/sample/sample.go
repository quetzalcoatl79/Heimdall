package sample

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/nxo/engine/internal/plugins"
	"github.com/nxo/engine/internal/ui"
)

type SamplePlugin struct {
	startedAt time.Time
}

func (p *SamplePlugin) Key() string         { return "sample" }
func (p *SamplePlugin) Version() string     { return "0.1.0" }
func (p *SamplePlugin) Description() string { return "Plugin de démonstration avec exemples de composants UI" }

func (p *SamplePlugin) Manifest() map[string]any {
	return map[string]any{
		"name":        "Sample",
		"version":     p.Version(),
		"description": p.Description(),
		"routes":      []map[string]any{},
		"permissions": []string{},
		"menu_items": []map[string]any{
			{
				"label":    "Démo",
				"path":     "/admin/plugins/sample",
				"icon":     "plug",
				"position": 999,
			},
		},
	}
}

func (p *SamplePlugin) RegisterRoutes(group *buffalo.App, deps plugins.Deps) {
	// GET /view - Returns the UI schema for the frontend to render
	group.GET("/view", func(c buffalo.Context) error {
		view := ui.NewView("Plugin Sample").
			WithDescription("Démonstration des composants UI disponibles").
			WithIcon("plug")

		// Add action buttons
		view.AddAction(ui.Action{
			ID:      "refresh",
			Label:   "Rafraîchir",
			Icon:    "refresh",
			Variant: "secondary",
		})
		view.AddAction(ui.Action{
			ID:      "export",
			Label:   "Exporter",
			Icon:    "download",
			Variant: "primary",
		})

		// Stats section
		view.AddComponent(ui.Stats(
			ui.Stat("Utilisateurs", 1234, ui.WithIcon("users"), ui.WithColor("blue"), ui.WithTrend(12.5, "ce mois")),
			ui.Stat("Plugins actifs", 3, ui.WithIcon("plug"), ui.WithColor("green")),
			ui.Stat("Requêtes/min", 847, ui.WithIcon("activity"), ui.WithColor("purple"), ui.WithTrend(-2.3, "vs hier")),
		))

		// Charts section - Line Chart
		view.AddComponent(ui.Card("Graphique linéaire - Trafic mensuel",
			ui.LineChart(
				[]string{"Jan", "Fév", "Mar", "Avr", "Mai", "Jun"},
				[]ui.ChartSeries{
					{Name: "Visiteurs", Data: []float64{4500, 5200, 4800, 6100, 5800, 7200}, Color: "#3b82f6"},
					{Name: "Pages vues", Data: []float64{12000, 15000, 13500, 18000, 16500, 21000}, Color: "#10b981"},
				},
				ui.ChartWithHeight(280),
			),
		))

		// Bar Chart
		view.AddComponent(ui.Card("Graphique en barres - Ventes par catégorie",
			ui.BarChart(
				[]string{"Électronique", "Vêtements", "Maison", "Sport", "Livres"},
				[]ui.ChartSeries{
					{Name: "2025", Data: []float64{45000, 32000, 28000, 18000, 12000}, Color: "#8b5cf6"},
					{Name: "2026", Data: []float64{52000, 38000, 31000, 24000, 15000}, Color: "#f59e0b"},
				},
			),
		))

		// Pie and Donut Charts
		view.AddComponent(ui.Grid(2,
			ui.Card("Répartition des utilisateurs",
				ui.PieChart(
					[]string{"Desktop", "Mobile", "Tablet"},
					[]float64{55, 35, 10},
					[]string{"#3b82f6", "#10b981", "#f59e0b"},
					ui.ChartWithHeight(250),
				),
			),
			ui.Card("Sources de trafic",
				ui.DonutChart(
					[]string{"Organique", "Direct", "Référents", "Social"},
					[]float64{42, 28, 18, 12},
					[]string{"#10b981", "#3b82f6", "#f59e0b", "#ec4899"},
					ui.ChartWithHeight(250),
				),
			),
		))

		// Area Chart
		view.AddComponent(ui.Card("Graphique en aire - Revenus",
			ui.AreaChart(
				[]string{"Sem 1", "Sem 2", "Sem 3", "Sem 4"},
				[]ui.ChartSeries{
					{Name: "Revenus", Data: []float64{15000, 18500, 22000, 28000}, Color: "#8b5cf6"},
					{Name: "Dépenses", Data: []float64{12000, 13500, 15000, 17000}, Color: "#ef4444"},
				},
				ui.ChartWithStacked(),
			),
		))

		// Alert examples
		view.AddComponent(ui.Grid(2,
			ui.Alert("info", "Ceci est un message d'information"),
			ui.Alert("success", "Opération réussie !"),
		))

		// Sample table with data
		columns := []ui.TableColumn{
			{Key: "id", Label: "ID", Width: "60px"},
			{Key: "name", Label: "Nom", Sortable: true},
			{Key: "email", Label: "Email"},
			{Key: "status", Label: "Statut", Render: "badge"},
			{Key: "created", Label: "Créé le", Render: "date"},
		}

		sampleData := []map[string]any{
			{"id": 1, "name": "Jean Dupont", "email": "jean@example.com", "status": "active", "created": "2026-01-01T10:00:00Z"},
			{"id": 2, "name": "Marie Martin", "email": "marie@example.com", "status": "active", "created": "2026-01-05T14:30:00Z"},
			{"id": 3, "name": "Pierre Bernard", "email": "pierre@example.com", "status": "inactive", "created": "2026-01-08T09:15:00Z"},
		}

		view.AddComponent(ui.Card("Liste des utilisateurs (exemple)",
			ui.Table(columns, sampleData),
		))

		// Progress and badges demo
		view.AddComponent(ui.Card("Progression et badges",
			ui.Grid(2,
				ui.Col(1,
					ui.Heading(4, "Progression du projet"),
					ui.Progress(75, 100, ui.ProgressWithColor("green")),
					ui.Text("75% complété"),
				),
				ui.Col(1,
					ui.Heading(4, "Statuts disponibles"),
					ui.Row(
						ui.Badge("Actif", "success"),
						ui.Badge("En attente", "warning"),
						ui.Badge("Inactif", "danger"),
						ui.Badge("Nouveau", "info"),
					),
				),
			),
		))

		// Form example
		view.AddComponent(ui.Card("Formulaire de contact (exemple)",
			ui.Form("contact-form",
				[]ui.FormField{
					ui.TextField("name", "Nom").WithPlaceholder("Votre nom").WithRequired(),
					ui.EmailField("email", "Email").WithPlaceholder("votre@email.com").WithRequired(),
					ui.SelectField("subject", "Sujet", []ui.SelectOption{
						{Value: "general", Label: "Question générale"},
						{Value: "support", Label: "Support technique"},
						{Value: "sales", Label: "Commercial"},
					}),
					ui.TextareaField("message", "Message").WithPlaceholder("Votre message...").WithRequired(),
					ui.CheckboxField("newsletter", "S'inscrire à la newsletter"),
				},
				ui.WithSubmitLabel("Envoyer"),
			),
		))

		// List example
		view.AddComponent(ui.Card("Informations du plugin",
			ui.List(
				ui.ListItem("Nom", p.Key()),
				ui.ListItem("Version", p.Version()),
				ui.ListItem("Description", p.Description()),
				ui.ListItem("Démarré", p.startedAt.Format("02/01/2006 15:04:05")),
			),
		))

		// JSON example
		view.AddComponent(ui.Card("Données JSON brutes",
			ui.JSON(map[string]any{
				"plugin":      p.Key(),
				"version":     p.Version(),
				"environment": "development",
				"features":    []string{"tables", "forms", "stats", "charts"},
			}),
		))

		return writeJSON(c, 200, view)
	})

	// Legacy endpoint
	group.GET("/hello", func(c buffalo.Context) error {
		return writeJSON(c, 200, map[string]any{
			"plugin":  p.Key(),
			"message": "hello from sample plugin",
		})
	})

	// Data endpoint for table
	group.GET("/users", func(c buffalo.Context) error {
		return writeJSON(c, 200, map[string]any{
			"data": []map[string]any{
				{"id": 1, "name": "Jean Dupont", "email": "jean@example.com", "status": "active"},
				{"id": 2, "name": "Marie Martin", "email": "marie@example.com", "status": "active"},
				{"id": 3, "name": "Pierre Bernard", "email": "pierre@example.com", "status": "inactive"},
			},
		})
	})
}

func writeJSON(c buffalo.Context, status int, payload any) error {
	c.Response().Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Response().WriteHeader(status)
	_ = json.NewEncoder(c.Response()).Encode(payload)
	return nil
}

func init() {
	plugins.Register(&SamplePlugin{startedAt: time.Now().UTC()})
}
