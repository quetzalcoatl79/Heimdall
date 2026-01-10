package handlers

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/google/uuid"
	"github.com/nxo/engine/internal/services"
)

// PluginHandler handles plugin endpoints
type PluginHandler struct {
	service *services.PluginService
	render  *render.Engine
}

// NewPluginHandler creates a new plugin handler
func NewPluginHandler(service *services.PluginService, r *render.Engine) *PluginHandler {
	return &PluginHandler{service: service, render: r}
}

// List returns all plugins
func (h *PluginHandler) List(c buffalo.Context) error {
	plugins, err := h.service.List()
	if err != nil {
		return c.Render(500, h.render.JSON(map[string]string{"error": "Failed to list plugins"}))
	}

	return c.Render(200, h.render.JSON(map[string]interface{}{
		"plugins": plugins,
	}))
}

// Get returns a plugin by ID
func (h *PluginHandler) Get(c buffalo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Render(400, h.render.JSON(map[string]string{"error": "Invalid plugin ID"}))
	}

	plugin, err := h.service.GetByID(id)
	if err != nil {
		return c.Render(404, h.render.JSON(map[string]string{"error": "Plugin not found"}))
	}

	return c.Render(200, h.render.JSON(plugin))
}

// Enable enables a plugin
func (h *PluginHandler) Enable(c buffalo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Render(400, h.render.JSON(map[string]string{"error": "Invalid plugin ID"}))
	}

	plugin, err := h.service.Enable(id)
	if err != nil {
		return c.Render(500, h.render.JSON(map[string]string{"error": "Failed to enable plugin"}))
	}

	return c.Render(200, h.render.JSON(plugin))
}

// Disable disables a plugin
func (h *PluginHandler) Disable(c buffalo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Render(400, h.render.JSON(map[string]string{"error": "Invalid plugin ID"}))
	}

	plugin, err := h.service.Disable(id)
	if err != nil {
		return c.Render(500, h.render.JSON(map[string]string{"error": "Failed to disable plugin"}))
	}

	return c.Render(200, h.render.JSON(plugin))
}

// Manifest returns the plugin manifest for the frontend
func (h *PluginHandler) Manifest(c buffalo.Context) error {
	manifest, err := h.service.GetManifest()
	if err != nil {
		return c.Render(500, h.render.JSON(map[string]string{"error": "Failed to get plugin manifest"}))
	}

	return c.Render(200, h.render.JSON(manifest))
}
