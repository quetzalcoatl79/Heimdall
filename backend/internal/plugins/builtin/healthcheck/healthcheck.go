package healthcheck

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/nxo/engine/internal/plugins"
	"github.com/nxo/engine/internal/ui"
	"github.com/nxo/engine/internal/workers"
)

type HealthcheckPlugin struct {
	startedAt time.Time
}

func (p *HealthcheckPlugin) Key() string         { return "healthcheck" }
func (p *HealthcheckPlugin) Version() string     { return "0.1.0" }
func (p *HealthcheckPlugin) Description() string { return "Healthcheck endpoints for app status" }

func (p *HealthcheckPlugin) Manifest() map[string]any {
	return map[string]any{
		"name":        "Healthcheck",
		"version":     p.Version(),
		"description": p.Description(),
		"routes":      []map[string]any{},
		"permissions": []string{},
		"menu_items": []map[string]any{
			{
				"label":    "Santé",
				"path":     "/admin/plugins/healthcheck",
				"icon":     "activity",
				"position": 100,
			},
		},
	}
}

func (p *HealthcheckPlugin) RegisterRoutes(group *buffalo.App, deps plugins.Deps) {
	// GET /view - Returns the UI schema for the frontend to render
	group.GET("/view", func(c buffalo.Context) error {
		ctx := context.Background()
		
		// Gather health data
		uptime := time.Since(p.startedAt).Seconds()
		dbStatus := "healthy"
		redisStatus := "healthy"
		workerStatus := "unknown"
		overallHealthy := true

		if deps.DB != nil {
			sqlDB, err := deps.DB.DB.DB()
			if err != nil || sqlDB.Ping() != nil {
				dbStatus = "unhealthy"
				overallHealthy = false
			}
		} else {
			dbStatus = "missing"
			overallHealthy = false
		}

		if deps.Cache != nil {
			if err := deps.Cache.Ping(ctx); err != nil {
				redisStatus = "unhealthy"
				overallHealthy = false
			}
		} else {
			redisStatus = "missing"
			overallHealthy = false
		}

		// Get worker stats
		var workerStats *workers.WorkerStats
		activeWorkers := 0
		queueLength := int64(0)
		jobsPending := int64(0)
		jobsRunning := int64(0)
		jobsCompleted := int64(0)
		jobsFailed := int64(0)
		
		if deps.Cache != nil && deps.DB != nil {
			stats, err := workers.GetWorkerStats(deps.Cache.Client(), deps.DB, "jobs:default")
			if err == nil {
				workerStats = stats
				activeWorkers = stats.ActiveWorkers
				queueLength = stats.QueueLength
				jobsPending = stats.JobsPending
				jobsRunning = stats.JobsRunning
				jobsCompleted = stats.JobsCompleted
				jobsFailed = stats.JobsFailed
				
				if stats.ActiveWorkers > 0 {
					workerStatus = "healthy"
				} else {
					workerStatus = "no workers"
					// Don't mark as unhealthy if no workers, just a warning
				}
			} else {
				workerStatus = "error"
			}
		}

		overallStatus := "ready"
		statusVariant := "success"
		if !overallHealthy {
			overallStatus = "not_ready"
			statusVariant = "error"
		}

		// Build the view schema
		view := ui.NewView("État de santé").
			WithDescription("Surveillance en temps réel de l'application").
			WithIcon("activity").
			WithRefresh(5)

		// Overall status alert
		statusMessage := "Tous les systèmes sont opérationnels"
		if !overallHealthy {
			statusMessage = "Certains services ne sont pas disponibles"
		}
		view.AddComponent(ui.Alert(statusVariant, statusMessage))

		// Main stats grid (4 columns now)
		view.AddComponent(ui.Grid(4,
			ui.Card("Uptime",
				ui.Stat("Temps de fonctionnement", formatUptime(uptime), ui.WithIcon("clock"), ui.WithColor("blue")),
				ui.Text("Démarré le "+p.startedAt.Format("02/01/2006 15:04:05")),
			),
			ui.Card("Base de données",
				ui.Stat("PostgreSQL", dbStatus,
					ui.WithIcon("database"),
					ui.WithColor(statusColor(dbStatus)),
				),
			),
			ui.Card("Cache",
				ui.Stat("Redis", redisStatus,
					ui.WithIcon("server"),
					ui.WithColor(statusColor(redisStatus)),
				),
			),
			ui.Card("Workers",
				ui.Stat("Background Jobs", workerStatus,
					ui.WithIcon("zap"),
					ui.WithColor(workerStatusColor(workerStatus)),
				),
				ui.Text(fmt.Sprintf("%d worker(s) actif(s)", activeWorkers)),
			),
		))

		// Workers section
		view.AddComponent(ui.Card("État des Workers",
			ui.Grid(4,
				ui.Stat("Workers actifs", fmt.Sprintf("%d", activeWorkers), ui.WithIcon("zap"), ui.WithColor("green")),
				ui.Stat("Jobs en attente", fmt.Sprintf("%d", jobsPending+queueLength), ui.WithIcon("clock"), ui.WithColor("yellow")),
				ui.Stat("Jobs en cours", fmt.Sprintf("%d", jobsRunning), ui.WithIcon("activity"), ui.WithColor("blue")),
				ui.Stat("Jobs terminés", fmt.Sprintf("%d", jobsCompleted), ui.WithIcon("check"), ui.WithColor("green")),
			),
		))

		// Worker details table if we have workers
		if workerStats != nil && len(workerStats.Workers) > 0 {
			workerData := make([]map[string]any, 0)
			for _, w := range workerStats.Workers {
				workerData = append(workerData, map[string]any{
					"id":           w.ID,
					"status":       w.Status,
					"started_at":   w.StartedAt.Format("02/01/2006 15:04:05"),
					"last_seen":    w.LastSeen.Format("15:04:05"),
					"jobs_handled": w.JobsHandled,
				})
			}
			view.AddComponent(ui.Card("Détail des Workers",
				ui.Table(
					[]ui.TableColumn{
						{Key: "id", Label: "Worker ID"},
						{Key: "status", Label: "Statut", Render: "badge"},
						{Key: "started_at", Label: "Démarré le"},
						{Key: "last_seen", Label: "Dernière activité"},
						{Key: "jobs_handled", Label: "Jobs traités"},
					},
					workerData,
				),
			))
		}

		// Job failure stats if any
		if jobsFailed > 0 {
			view.AddComponent(ui.Alert("warning", fmt.Sprintf("%d job(s) en échec", jobsFailed)))
		}

		// Details section
		view.AddComponent(ui.Card("Informations détaillées",
			ui.Grid(2,
				ui.Col(1,
					ui.ListItem("Plugin", p.Key()),
					ui.ListItem("Version", p.Version()),
					ui.ListItem("Statut global", overallStatus),
				),
				ui.Col(1,
					ui.ListItem("Uptime (secondes)", formatFloat(uptime)),
					ui.ListItem("Dernière vérification", time.Now().UTC().Format(time.RFC3339)),
				),
			),
		))

		return writeJSON(c, 200, view)
	})

	// Legacy endpoints for backward compatibility
	group.GET("/health", func(c buffalo.Context) error {
		uptime := time.Since(p.startedAt).Seconds()
		return writeJSON(c, 200, map[string]any{
			"plugin":         p.Key(),
			"status":         "ok",
			"timestamp":      time.Now().UTC(),
			"started_at":     p.startedAt,
			"uptime_seconds": uptime,
		})
	})

	group.GET("/ready", func(c buffalo.Context) error {
		ctx := context.Background()
		checks := map[string]string{}
		healthy := true

		if deps.DB == nil {
			checks["database"] = "missing"
			healthy = false
		} else {
			sqlDB, err := deps.DB.DB.DB()
			if err != nil || sqlDB.Ping() != nil {
				checks["database"] = "unhealthy"
				healthy = false
			} else {
				checks["database"] = "healthy"
			}
		}

		if deps.Cache == nil {
			checks["redis"] = "missing"
			healthy = false
		} else if err := deps.Cache.Ping(ctx); err != nil {
			checks["redis"] = "unhealthy"
			healthy = false
		} else {
			checks["redis"] = "healthy"
		}

		statusCode := 200
		overallStatus := "ready"
		if !healthy {
			statusCode = 503
			overallStatus = "not_ready"
		}

		return writeJSON(c, statusCode, map[string]any{
			"plugin":    p.Key(),
			"status":    overallStatus,
			"checks":    checks,
			"timestamp": time.Now().UTC(),
		})
	})
}

func formatUptime(seconds float64) string {
	d := int(seconds) / 86400
	h := (int(seconds) % 86400) / 3600
	m := (int(seconds) % 3600) / 60
	s := int(seconds) % 60
	
	if d > 0 {
		return formatInt(d) + "j " + formatInt(h) + "h " + formatInt(m) + "m"
	}
	if h > 0 {
		return formatInt(h) + "h " + formatInt(m) + "m " + formatInt(s) + "s"
	}
	if m > 0 {
		return formatInt(m) + "m " + formatInt(s) + "s"
	}
	return formatInt(s) + "s"
}

func formatInt(i int) string {
	return string(rune('0'+i/10)) + string(rune('0'+i%10))
}

func formatFloat(f float64) string {
	return time.Duration(f * float64(time.Second)).String()
}

func statusColor(status string) string {
	if status == "healthy" {
		return "green"
	}
	return "red"
}

func workerStatusColor(status string) string {
	switch status {
	case "healthy":
		return "green"
	case "no workers":
		return "yellow"
	case "error", "unhealthy":
		return "red"
	default:
		return "gray"
	}
}

func writeJSON(c buffalo.Context, status int, payload any) error {
	c.Response().Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Response().WriteHeader(status)
	_ = json.NewEncoder(c.Response()).Encode(payload)
	return nil
}

func init() {
	plugins.Register(&HealthcheckPlugin{startedAt: time.Now().UTC()})
}
