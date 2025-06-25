package app

import (
	"net/http"

	"github.com/timoruohomaki/open311-to-Go/config"
	"github.com/timoruohomaki/open311-to-Go/internal/handlers"
	"github.com/timoruohomaki/open311-to-Go/internal/repository"
	"github.com/timoruohomaki/open311-to-Go/pkg/logger"
	"github.com/timoruohomaki/open311-to-Go/pkg/middleware"
	"github.com/timoruohomaki/open311-to-Go/pkg/router"
)

// App represents the application
type App struct {
	Router *router.Router
	Config *config.Config
	Logger logger.Logger
}

// New creates a new application
func New(cfg *config.Config, log logger.Logger, userRepo repository.UserRepository, serviceRepo repository.ServiceRepository) *App {
	// Create router
	r := router.New()

	// Add middleware
	r.Use(middleware.LoggingMiddleware(log))
	r.Use(middleware.ContentTypeMiddleware)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(log, userRepo)
	serviceHandler := handlers.NewServiceHandler(log, serviceRepo)

	// Register routes
	// User routes
	r.Handle("GET", "/api/v1/users", userHandler.GetUsers)
	r.Handle("GET", "/api/v1/users/", userHandler.GetUsers) // Trailing slash version
	r.Handle("GET", "/api/v1/users/{id}", userHandler.GetUser)

	// Service routes
	r.Handle("GET", "/api/v1/services", serviceHandler.GetServices)
	r.Handle("GET", "/api/v1/services/", serviceHandler.GetServices) // Trailing slash version
	r.Handle("GET", "/api/v1/services/{id}", serviceHandler.GetService)
	r.Handle("POST", "/api/v1/services", serviceHandler.CreateService)
	r.Handle("PUT", "/api/v1/services/{id}", serviceHandler.UpdateService)
	r.Handle("DELETE", "/api/v1/services/{id}", serviceHandler.DeleteService)

	return &App{
		Router: r,
		Config: cfg,
		Logger: log,
	}
}

// Handler returns the HTTP handler for the application
func (a *App) Handler() http.Handler {
	return a.Router
}
