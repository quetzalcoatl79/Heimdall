# Heimdall AI Coding Instructions

Heimdall is a plugin-based WiFi pentesting platform with a Go backend and Next.js frontend. The architecture enables dynamic UI generation where backend plugins define JSON schemas that the frontend renders.

## Architecture Overview

### Plugin System (Compile-Time, Not Runtime)
- **Critical**: Go plugins are compiled into the binary, not loaded at runtime (see [internal/plugins/registry.go](../backend/internal/plugins/registry.go))
- New plugins must be registered in `backend/internal/plugins/builtin/builtin.go` and imported to trigger `init()`
- Each plugin implements: `Key()`, `Version()`, `Description()`, `RegisterRoutes()`, `Manifest()`
- Routes are auto-mounted at `/api/v1/plugins/{pluginKey}` via [internal/plugins/mount.go](../backend/internal/plugins/mount.go)
- Middleware `RequireEnabled()` gates plugin access based on DB state ([internal/plugins/enabled_middleware.go](../backend/internal/plugins/enabled_middleware.go))

### UI Schema System (Backend-Driven)
- Backend plugins return JSON schemas via `/view` endpoints (see [internal/ui/schema.go](../backend/internal/ui/schema.go))
- Frontend [DynamicRenderer.tsx](../frontend/src/components/ui/DynamicRenderer.tsx) interprets schemas to render components
- Component types: `card`, `table`, `form`, `stats`, `alert`, `tabs`, `grid`, `badge`, `progress`
- Example: [wifi/wifi.go](../backend/internal/plugins/builtin/wifi/wifi.go) demonstrates `ui.Card()`, `ui.Form()`, `ui.Table()` builders

### Worker System (Redis-Based)
- Workers use Redis for job queuing and coordination (see [internal/workers/manager.go](../backend/internal/workers/manager.go))
- Jobs are enqueued with handlers registered via `RegisterHandler(jobType string, handler JobHandler)`
- Worker heartbeats stored at `workers:heartbeat` key with 30s expiry
- Start worker process via `cmd/worker/main.go`, separate from API server

## Development Workflow

### Local Setup Commands
```bash
# Start infrastructure (run once)
docker compose up -d postgres redis

# Development mode (backend + frontend + worker)
./run_heimdall.sh --dev

# Production mode (systemd services)
sudo ./run_heimdall.sh --prod
```

### Key Make Targets
- `make dev` - Full stack development mode
- `make migrate` - Run DB migrations
- `make backend` - Run Go API server (port 8080)
- `make frontend` - Run Next.js dev server (port 3000)

### Environment Variables
Default config in [backend/internal/config/config.go](../backend/internal/config/config.go):
- `DB_HOST=localhost`, `DB_PORT=5433`, `DB_NAME=engine_dev`
- `REDIS_HOST=localhost`, `REDIS_PORT=6379`
- `APP_PORT=8080`, `APP_ENV=development`

## Code Conventions

### Backend (Go)
- **Buffalo framework**: Routes use `buffalo.Context`, handlers return `error`
- **Module path**: `github.com/nxo/engine` (not heimdall - legacy rename)
- **DB access**: Use `*database.DB` wrapper, not raw GORM
- **JSON types**: Use `models.JSON` (jsonb) for flexible plugin data storage
- **Middleware**: JWT auth via [internal/middleware/middleware.go](../backend/internal/middleware/middleware.go)

### Frontend (Next.js 14)
- **API client**: [lib/api/client.ts](../frontend/src/lib/api/client.ts) has auto-refresh interceptor for JWT tokens
- **Auth state**: Zustand store at [lib/store/auth.ts](../frontend/src/lib/store/auth.ts), uses `js-cookie` for token storage
- **i18n**: Messages in [lib/i18n/messages.ts](../frontend/src/lib/i18n/messages.ts), switch via `LanguageSwitcher`
- **Admin layout**: Protected routes under `/admin` with sidebar navigation

### Plugin Development Pattern
1. Create plugin struct in `backend/internal/plugins/builtin/{name}/{name}.go`
2. Implement `Plugin` interface methods
3. Register in `builtin/builtin.go`: `plugins.Register(&yourplugin.YourPlugin{})`
4. Define routes that return UI schemas using `ui.NewView()` builders
5. Add menu items in `Manifest()` for navigation sidebar

### UI Schema Builder Functions
```go
// Create views with auto-refresh
view := ui.NewView("Title").WithDescription("desc").WithRefresh(5)

// Add components
view.AddComponent(ui.Card("Header", 
    ui.Table(columns, data),
    ui.Form("id", fields),
))

// Add actions (buttons)
view.AddAction(ui.Action{ID: "scan", Label: "Start Scan", Icon: "radar"})
```

## Critical Files

- [backend/cmd/api/main.go](../backend/cmd/api/main.go) - API server entrypoint
- [backend/cmd/worker/main.go](../backend/cmd/worker/main.go) - Background worker entrypoint  
- [backend/internal/server/server.go](../backend/internal/server/server.go) - Buffalo app initialization, CORS, routes
- [frontend/src/app/admin/plugins/[key]/page.tsx](../frontend/src/app/admin/plugins/[key]/page.tsx) - Dynamic plugin page renderer
- [docker-compose.yml](../docker-compose.yml) - Local infra (postgres:5433, redis:6379)

## Testing & Debugging

- **Debug containers**: `make docker-debug` starts Adminer (port 8082) and Redis Commander (port 8081)
- **Logs**: Development logs to stdout; production uses `/opt/heimdall/logs/`
- **PID files**: Production mode writes to `/var/run/heimdall/`

## Security Context

- Platform designed for Kali Linux/Parrot OS (pentest distros)
- WiFi plugin requires root for monitor mode operations
- JWT tokens stored in secure, httpOnly cookies
- Default admin password set via migration `000002_update_admin_password.up.sql`
