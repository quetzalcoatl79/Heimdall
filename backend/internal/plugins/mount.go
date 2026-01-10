package plugins

import (
	"fmt"

	"github.com/gobuffalo/buffalo"
)

// Mount registers all plugin routes under /api/v1/plugins/{pluginKey}.
// These routes are expected to be placed under the JWT-protected group.
func Mount(protected *buffalo.App, deps Deps) error {
	if protected == nil {
		return fmt.Errorf("plugins: protected group is nil")
	}

	pluginsGroup := protected.Group("/plugins")
	for _, p := range All() {
		g := pluginsGroup.Group("/" + p.Key())
		g.Use(RequireEnabled(deps, p.Key()))
		p.RegisterRoutes(g, deps)
	}

	return nil
}
