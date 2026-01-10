package services

import (
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/nxo/engine/internal/database"
	"github.com/nxo/engine/internal/models"
)

var (
	ErrPluginNotFound = errors.New("plugin not found")
)

// PluginService handles plugin operations
type PluginService struct {
	db *database.DB
}

// NewPluginService creates a new plugin service
func NewPluginService(db *database.DB) *PluginService {
	return &PluginService{db: db}
}

// PluginManifest represents the frontend plugin manifest
type PluginManifest struct {
	Plugins []PluginManifestEntry `json:"plugins"`
}

// PluginManifestEntry represents a single plugin in the manifest
type PluginManifestEntry struct {
	ID          string     `json:"id"`
	Key         string     `json:"key"`
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	Description string     `json:"description,omitempty"`
	Routes      []Route    `json:"routes"`
	Permissions []string   `json:"permissions"`
	MenuItems   []MenuItem `json:"menu_items"`
}

// Route represents a plugin route
type Route struct {
	Path      string `json:"path"`
	Component string `json:"component"`
}

// MenuItem represents a plugin menu item
type MenuItem struct {
	Label    string `json:"label"`
	Path     string `json:"path"`
	Icon     string `json:"icon"`
	Position int    `json:"position"`
}

// List returns all plugins
func (s *PluginService) List() ([]models.Plugin, error) {
	var plugins []models.Plugin
	if err := s.db.Order("name ASC").Find(&plugins).Error; err != nil {
		return nil, err
	}
	return plugins, nil
}

// GetByID returns a plugin by ID
func (s *PluginService) GetByID(id uuid.UUID) (*models.Plugin, error) {
	var plugin models.Plugin
	if err := s.db.First(&plugin, id).Error; err != nil {
		return nil, ErrPluginNotFound
	}
	return &plugin, nil
}

// Enable enables a plugin
func (s *PluginService) Enable(id uuid.UUID) (*models.Plugin, error) {
	plugin, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if err := s.db.Model(plugin).Update("enabled", true).Error; err != nil {
		return nil, err
	}

	// Note: Plugin lifecycle hooks (onEnable) can be added when needed
	// Plugins are compiled-in, so enable/disable is mainly a DB flag

	return s.GetByID(id)
}

// Disable disables a plugin
func (s *PluginService) Disable(id uuid.UUID) (*models.Plugin, error) {
	plugin, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if err := s.db.Model(plugin).Update("enabled", false).Error; err != nil {
		return nil, err
	}

	// Note: Plugin lifecycle hooks (onDisable) can be added when needed
	// Plugins are compiled-in, so enable/disable is mainly a DB flag

	return s.GetByID(id)
}

// GetManifest returns the plugin manifest for the frontend
func (s *PluginService) GetManifest() (*PluginManifest, error) {
	var plugins []models.Plugin
	if err := s.db.Where("enabled = ?", true).Find(&plugins).Error; err != nil {
		return nil, err
	}

	manifest := &PluginManifest{
		Plugins: make([]PluginManifestEntry, 0, len(plugins)),
	}

	for _, p := range plugins {
		entry := PluginManifestEntry{
			ID:          p.ID.String(),
			Key:         p.Name, // name is used as key
			Name:        p.Name,
			Version:     p.Version,
			Description: p.Description,
		}

		// Parse manifest from plugin config
		if p.Manifest != nil {
			data, _ := json.Marshal(p.Manifest)
			json.Unmarshal(data, &entry)
			// Ensure key is always set from DB name (lower case)
			entry.Key = p.Name
		}

		manifest.Plugins = append(manifest.Plugins, entry)
	}

	return manifest, nil
}
