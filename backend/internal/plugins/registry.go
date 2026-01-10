package plugins

import (
	"fmt"
	"sort"
	"sync"

	"github.com/gobuffalo/buffalo"
	"github.com/nxo/engine/internal/cache"
	"github.com/nxo/engine/internal/config"
	"github.com/nxo/engine/internal/database"
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
