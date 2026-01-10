package handlers

import (
	"encoding/json"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/nxo/engine/internal/cache"
	"github.com/nxo/engine/internal/database"
	"github.com/nxo/engine/internal/models"
	"github.com/nxo/engine/internal/workers"
)

// WorkerHandler handles worker-related requests
type WorkerHandler struct {
	db    *database.DB
	cache *cache.Redis
	r     *render.Engine
}

// NewWorkerHandler creates a new worker handler
func NewWorkerHandler(db *database.DB, cache *cache.Redis) *WorkerHandler {
	return &WorkerHandler{
		db:    db,
		cache: cache,
		r:     render.New(render.Options{}),
	}
}

// GetStats returns worker statistics
func (h *WorkerHandler) GetStats(c buffalo.Context) error {
	stats, err := workers.GetWorkerStats(h.cache.Client(), h.db, "jobs:default")
	if err != nil {
		return c.Render(500, h.r.JSON(map[string]string{"error": err.Error()}))
	}
	return c.Render(200, h.r.JSON(stats))
}

// ListJobs returns a list of jobs
func (h *WorkerHandler) ListJobs(c buffalo.Context) error {
	limit := 50
	if l := c.Param("limit"); l != "" {
		if _, err := json.Marshal(l); err == nil {
			// Parse limit if valid
		}
	}

	var jobs []models.Job
	if err := h.db.Order("created_at DESC").Limit(limit).Find(&jobs).Error; err != nil {
		return c.Render(500, h.r.JSON(map[string]string{"error": err.Error()}))
	}

	return c.Render(200, h.r.JSON(map[string]any{
		"jobs":  jobs,
		"count": len(jobs),
	}))
}

// GetJob returns a single job by ID
func (h *WorkerHandler) GetJob(c buffalo.Context) error {
	id := c.Param("id")
	
	var job models.Job
	if err := h.db.First(&job, "id = ?", id).Error; err != nil {
		return c.Render(404, h.r.JSON(map[string]string{"error": "job not found"}))
	}

	return c.Render(200, h.r.JSON(job))
}
