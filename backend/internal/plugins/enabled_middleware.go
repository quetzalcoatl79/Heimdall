package plugins

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/nxo/engine/internal/models"
)

type enabledCacheItem struct {
	enabled   bool
	expiresAt time.Time
}

type enabledCache struct {
	mu    sync.RWMutex
	ttl   time.Duration
	items map[string]enabledCacheItem
}

func newEnabledCache(ttl time.Duration) *enabledCache {
	return &enabledCache{ttl: ttl, items: map[string]enabledCacheItem{}}
}

func (c *enabledCache) get(key string) (bool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	it, ok := c.items[key]
	if !ok {
		return false, false
	}
	if time.Now().After(it.expiresAt) {
		return false, false
	}
	return it.enabled, true
}

func (c *enabledCache) set(key string, enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = enabledCacheItem{enabled: enabled, expiresAt: time.Now().Add(c.ttl)}
}

var globalEnabledCache = newEnabledCache(2 * time.Second)

// RequireEnabled blocks plugin routes if the plugin is disabled.
func RequireEnabled(deps Deps, pluginKey string) buffalo.MiddlewareFunc {
	return func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			if deps.DB == nil {
				return next(c)
			}

			if enabled, ok := globalEnabledCache.get(pluginKey); ok {
				if !enabled {
					return writeJSON(c, 404, map[string]string{"error": "Plugin disabled"})
				}
				return next(c)
			}

			var plugin models.Plugin
			err := deps.DB.Where("name = ? AND enabled = ?", pluginKey, true).First(&plugin).Error
			enabled := err == nil
			globalEnabledCache.set(pluginKey, enabled)
			if !enabled {
				return writeJSON(c, 404, map[string]string{"error": "Plugin disabled"})
			}
			return next(c)
		}
	}
}

func writeJSON(c buffalo.Context, status int, payload any) error {
	c.Response().Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Response().WriteHeader(status)
	_ = json.NewEncoder(c.Response()).Encode(payload)
	return nil
}
