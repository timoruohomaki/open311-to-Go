package app

import (
	"net/http"

	"github.com/timoruohomaki/open311-to-Go/api/handlers"
	"github.com/timoruohomaki/open311-to-Go/api/middleware"
	"github.com/timoruohomaki/open311-to-Go/config"
	"github.com/timoruohomaki/open311-to-Go/domain/repository"
	"github.com/timoruohomaki/open311-to-Go/pkg/logger"
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
	r.AddRoute("GET", "/api/v1/users", userHandler.GetUsers)
	r.AddRoute("GET", "/api/v1/users/", userHandler.GetUsers) // Trailing slash version
	r.AddRoute("GET", "/api/v1/users/{id}", userHandler.GetUser)
	r.AddRoute("POST", "/api/v1/users", userHandler.CreateUser)
	r.AddRoute("PUT", "/api/v1/users/{id}", userHandler.UpdateUser)
	r.AddRoute("DELETE", "/api/v1/users/{id}", userHandler.DeleteUser)

	// Service routes (previously Product)
	r.AddRoute("GET", "/api/v1/services", serviceHandler.GetServices)
	r.AddRoute("GET", "/api/v1/services/", serviceHandler.GetServices) // Trailing slash version
	r.AddRoute("GET", "/api/v1/services/{id}", serviceHandler.GetService)
	r.AddRoute("POST", "/api/v1/services", serviceHandler.CreateService)
	r.AddRoute("PUT", "/api/v1/services/{id}", serviceHandler.UpdateService)
	r.AddRoute("DELETE", "/api/v1/services/{id}", serviceHandler.DeleteService)

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
