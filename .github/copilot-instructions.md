# Heimdall AI Coding Instructions

Heimdall is a plugin-based WiFi pentesting platform with a Go/Buffalo backend and Next.js 14 frontend. Backend plugins define JSON UI schemas that the frontend's `DynamicRenderer` interprets.

## Architecture: Plugin System

**Compile-time plugins (NOT runtime loaded):**
1. Create struct in `backend/internal/plugins/builtin/{name}/{name}.go` implementing `Plugin` interface
2. Register via blank import in [builtin/builtin.go](../backend/internal/plugins/builtin/builtin.go): `_ "github.com/nxo/engine/internal/plugins/builtin/{name}"`
3. Plugin `init()` calls `plugins.Register(p)` - routes auto-mount at `/api/v1/plugins/{key}`
4. Implement optional `PluginWithModels` interface to auto-migrate GORM models on startup

```go
// Required interface (registry.go)
type Plugin interface {
    Key() string; Version() string; Description() string
    RegisterRoutes(group *buffalo.App, deps Deps)
    Manifest() map[string]any  // menu_items, permissions
}
```

## Architecture: Backend-Driven UI

Plugins return JSON schemas via `/view` endpoints → frontend [DynamicRenderer.tsx](../frontend/src/components/ui/DynamicRenderer.tsx) renders them.

```go
// Example: wifi/view.go pattern
view := ui.NewView("Title").WithIcon("wifi").WithRefresh(10)
view.AddComponent(ui.Card("Header", ui.Table(columns, data)))
view.AddAction(ui.Action{ID: "scan", Label: "Scan", Variant: "primary"})
return writeJSON(c, 200, view)
```

Component types: `card`, `table`, `form`, `stats`, `tabs`, `grid`, `badge`, `progress`, `chart`, `modal`, `button`

**Custom React components** (only when schema insufficient): Register in [frontend/src/lib/plugins/registry.tsx](../frontend/src/lib/plugins/registry.tsx)

## Development Commands

```bash
docker compose up -d postgres redis    # Infrastructure (ports 5433, 6379)
make dev                               # Full stack (backend:8080, frontend:3000, worker)
make migrate                           # Run DB migrations
make docker-debug                      # + Adminer:8082, Redis Commander:8081
```

## Code Conventions

**Backend (Go):**
- Module: `github.com/nxo/engine` (legacy name, not heimdall)
- Use `*database.DB` wrapper, `models.JSON` for jsonb columns
- Handlers return `error`, use `writeJSON(c, status, payload)`
- JWT middleware in [internal/middleware/middleware.go](../backend/internal/middleware/middleware.go)

**Frontend (Next.js 14):**
- API client with JWT refresh: [lib/api/client.ts](../frontend/src/lib/api/client.ts)
- Auth state: Zustand store + `js-cookie` for tokens
- i18n: Messages in [lib/i18n/messages.ts](../frontend/src/lib/i18n/messages.ts) (fr/en)
- Plugin pages: `/admin/plugins/[key]` routes to dynamic renderer

## Key Files

| Purpose | Path |
|---------|------|
| API entrypoint | `backend/cmd/api/main.go` |
| Route setup & CORS | `backend/internal/server/server.go` |
| UI schema builders | `backend/internal/ui/schema.go`, `forms.go` |
| Plugin registry | `backend/internal/plugins/registry.go` |
| Dynamic UI renderer | `frontend/src/components/ui/DynamicRenderer.tsx` |
| Plugin page router | `frontend/src/app/admin/plugins/[key]/page.tsx` |
| WiFi plugin (example) | `backend/internal/plugins/builtin/wifi/` |

## Environment (defaults in config.go)

`DB_HOST=localhost` `DB_PORT=5433` `DB_NAME=engine_dev` `REDIS_PORT=6379` `APP_PORT=8080`

## Security Notes

- Designed for Kali/Parrot (pentest distros); WiFi features require root/monitor mode
- Default admin: `admin@heimdall.local` / `admin123` (migration 000002)
