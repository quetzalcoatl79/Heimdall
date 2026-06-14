package plugins

import (
	"fmt"
	"log"
	"sort"
	"sync"

	"github.com/gobuffalo/buffalo"
	"github.com/nxo/engine/internal/cache"
	"github.com/nxo/engine/internal/config"
	"github.com/nxo/engine/internal/database"
	"gorm.io/gorm"
)

// Plugin is a compiled-in extension point.
//
// Important: Go cannot safely load arbitrary Go code at runtime in a portable way.
// Plugins must be compiled into the backend binary (like apps registered in nano-core).
type Plugin interface {
	Key() string
	Version() string
	Description() string

	// RegisterRoutes should attach plugin routes to the provided group.
	// Routes are mounted under /api/v1/plugins/{Key()} and are JWT-protected.
	RegisterRoutes(group *buffalo.App, deps Deps)

	// Manifest returns a JSON-compatible description used by the admin manifest endpoint.
	Manifest() map[string]any
}

// PluginWithModels is an optional interface for plugins that have database models.
// Plugins implementing this interface will have their models auto-migrated on startup.
type PluginWithModels interface {
	Plugin
	// Models returns all GORM models that need to be auto-migrated for this plugin
	Models() []interface{}
}

// PluginWithMigrations is an optional interface for plugins that have custom migrations.
type PluginWithMigrations interface {
	Plugin
	// Migrate runs custom migrations for the plugin
	Migrate(db *gorm.DB) error
}

type Deps struct {
	DB     *database.DB
	Cache  *cache.Redis
	Config *config.Config
}

var (
	mu       sync.RWMutex
	registry = map[string]Plugin{}
)

func Register(p Plugin) {
	mu.Lock()
	defer mu.Unlock()

	key := p.Key()
	if key == "" {
		panic("plugins: plugin key must not be empty")
	}
	if _, exists := registry[key]; exists {
		panic(fmt.Sprintf("plugins: duplicate plugin key: %s", key))
	}
	registry[key] = p
}

func All() []Plugin {
	mu.RLock()
	defer mu.RUnlock()

	out := make([]Plugin, 0, len(registry))
	for _, p := range registry {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key() < out[j].Key() })
	return out
}

// AutoMigrateAll runs auto-migration for all plugins that implement PluginWithModels.
// This should be called after database initialization.
func AutoMigrateAll(db *gorm.DB) error {
	mu.RLock()
	defer mu.RUnlock()

	for key, p := range registry {
		// Check if plugin has models to migrate
		if pm, ok := p.(PluginWithModels); ok {
			models := pm.Models()
			if len(models) > 0 {
				log.Printf("plugins: auto-migrating %d models for plugin %q", len(models), key)
				if err := db.AutoMigrate(models...); err != nil {
					return fmt.Errorf("plugins: failed to migrate models for %s: %w", key, err)
				}
			}
		}

		// Check if plugin has custom migrations
		if pm, ok := p.(PluginWithMigrations); ok {
			log.Printf("plugins: running custom migrations for plugin %q", key)
			if err := pm.Migrate(db); err != nil {
				return fmt.Errorf("plugins: failed custom migration for %s: %w", key, err)
			}
		}
	}

	return nil
}
