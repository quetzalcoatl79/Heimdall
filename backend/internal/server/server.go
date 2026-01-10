package server

import (
	"context"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	contenttype "github.com/gobuffalo/mw-contenttype"
	paramlogger "github.com/gobuffalo/mw-paramlogger"
	"github.com/nxo/engine/internal/cache"
	"github.com/nxo/engine/internal/config"
	"github.com/nxo/engine/internal/database"
	"github.com/nxo/engine/internal/handlers"
	"github.com/nxo/engine/internal/middleware"
	"github.com/nxo/engine/internal/plugins"
	_ "github.com/nxo/engine/internal/plugins/builtin"
	"github.com/nxo/engine/internal/services"
	"github.com/rs/cors"
)

// Server wraps the Buffalo app
type Server struct {
	app    *buffalo.App
	db     *database.DB
	cache  *cache.Redis
	config *config.Config
}

// New creates a new server instance
func New(cfg *config.Config) (*Server, error) {
	// Initialize database
	db, err := database.New(&cfg.Database)
	if err != nil {
		return nil, err
	}

	// Initialize Redis cache
	redisCache := cache.New(&cfg.Redis)

	// Create Buffalo app
	app := buffalo.New(buffalo.Options{
		Env:         cfg.App.Env,
		SessionName: "_engine_session",
		Addr:        "0.0.0.0:" + cfg.App.Port,
	})

	// JSON renderer
	r := render.New(render.Options{
		DefaultContentType: "application/json",
	})

	// Global middleware
	app.Use(paramlogger.ParameterLogger)
	app.Use(contenttype.Set("application/json"))
	
	// CORS middleware
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001", "http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
	})
	app.Use(func(next buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			corsMiddleware.HandlerFunc(c.Response(), c.Request())
			// Handle CORS preflight requests early.
			// Buffalo does not register OPTIONS routes by default; without this,
			// preflight can fall through and return 405 without CORS headers.
			if c.Request().Method == "OPTIONS" {
				c.Response().WriteHeader(204)
				return nil
			}
			return next(c)
		}
	})

	// Initialize services
	authService := services.NewAuthService(db, redisCache, cfg)
	userService := services.NewUserService(db)
	pluginService := services.NewPluginService(db)

	pluginDeps := plugins.Deps{DB: db, Cache: redisCache, Config: cfg}
	if err := plugins.SyncDiscovered(pluginDeps); err != nil {
		return nil, err
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, r)
	userHandler := handlers.NewUserHandler(userService, r)
	pluginHandler := handlers.NewPluginHandler(pluginService, r)
	healthHandler := handlers.NewHealthHandler(db, redisCache, r)
	workerHandler := handlers.NewWorkerHandler(db, redisCache)

	// Routes
	if err := setupRoutes(app, cfg, authHandler, userHandler, pluginHandler, healthHandler, workerHandler, pluginDeps); err != nil {
		return nil, err
	}

	return &Server{
		app:    app,
		db:     db,
		cache:  redisCache,
		config: cfg,
	}, nil
}

func setupRoutes(
	app *buffalo.App,
	cfg *config.Config,
	auth *handlers.AuthHandler,
	user *handlers.UserHandler,
	plugin *handlers.PluginHandler,
	health *handlers.HealthHandler,
	worker *handlers.WorkerHandler,
	pluginDeps plugins.Deps,
) error {
	// CORS preflight: Buffalo will otherwise answer 405 for OPTIONS on routes
	// that only declare GET/POST/etc, which breaks browsers.
	app.OPTIONS("/{path:.*}", func(c buffalo.Context) error {
		c.Response().WriteHeader(204)
		return nil
	})

	// Health check
	app.GET("/health", health.Check)
	app.GET("/ready", health.Ready)

	// API v1
	api := app.Group("/api/v1")

	// Auth routes (public)
	authGroup := api.Group("/auth")
	authGroup.POST("/register", auth.Register)
	authGroup.POST("/login", auth.Login)
	authGroup.POST("/refresh", auth.Refresh)
	authGroup.POST("/logout", auth.Logout)

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.JWT(cfg.JWT.Secret))

	// User routes
	protected.GET("/users/me", user.Me)
	protected.PUT("/users/me", user.UpdateMe)

	// Admin routes
	admin := protected.Group("/admin")
	admin.Use(middleware.RequireRole("admin"))

	// Admin: Users
	admin.GET("/users", user.List)
	admin.GET("/users/{id}", user.Get)
	admin.POST("/users", user.Create)
	admin.PUT("/users/{id}", user.Update)
	admin.DELETE("/users/{id}", user.Delete)

	// Admin: Plugins
	admin.GET("/plugins", plugin.List)
	admin.GET("/plugins/manifest", plugin.Manifest)
	admin.GET("/plugins/{id}", plugin.Get)
	admin.POST("/plugins/{id}/enable", plugin.Enable)
	admin.POST("/plugins/{id}/disable", plugin.Disable)

	// Admin: Workers & Jobs
	admin.GET("/workers/stats", worker.GetStats)
	admin.GET("/jobs", worker.ListJobs)
	admin.GET("/jobs/{id}", worker.GetJob)

	// Plugin runtime routes (JWT-protected)
	if err := plugins.Mount(protected, pluginDeps); err != nil {
		return err
	}

	return nil
}

// Start starts the server
func (s *Server) Start(addr string) error {
	return s.app.Serve()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.db.Close(); err != nil {
		return err
	}
	return s.cache.Close()
}
