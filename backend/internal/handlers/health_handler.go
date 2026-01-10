package handlers

import (
	"context"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/nxo/engine/internal/cache"
	"github.com/nxo/engine/internal/database"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db     *database.DB
	cache  *cache.Redis
	render *render.Engine
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *database.DB, cache *cache.Redis, r *render.Engine) *HealthHandler {
	return &HealthHandler{db: db, cache: cache, render: r}
}

// Check returns basic health status
func (h *HealthHandler) Check(c buffalo.Context) error {
	return c.Render(200, h.render.JSON(map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
	}))
}

// Ready returns readiness status including dependencies
func (h *HealthHandler) Ready(c buffalo.Context) error {
	ctx := context.Background()
	status := map[string]string{}
	healthy := true

	// Check database
	sqlDB, err := h.db.DB.DB()
	if err != nil || sqlDB.Ping() != nil {
		status["database"] = "unhealthy"
		healthy = false
	} else {
		status["database"] = "healthy"
	}

	// Check Redis
	if err := h.cache.Ping(ctx); err != nil {
		status["redis"] = "unhealthy"
		healthy = false
	} else {
		status["redis"] = "healthy"
	}

	statusCode := 200
	overallStatus := "ready"
	if !healthy {
		statusCode = 503
		overallStatus = "not_ready"
	}

	return c.Render(statusCode, h.render.JSON(map[string]interface{}{
		"status":     overallStatus,
		"checks":     status,
		"timestamp":  time.Now().UTC(),
	}))
}
