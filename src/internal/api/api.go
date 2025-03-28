package api

import (
	"net/http"

	"github.com/timoruohomaki/open311-to-Go/config"
	"github.com/timoruohomaki/open311-to-Go/internal/handlers"
	"github.com/timoruohomaki/open311-to-Go/internal/repository"
	"github.com/timoruohomaki/open311-to-Go/pkg/logger"
	"github.com/timoruohomaki/open311-to-Go/pkg/middleware"
	"github.com/timoruohomaki/open311-to-Go/pkg/router"
)

// API represents the REST API
type API struct {
	router *router.Router
	config *config.Config
	logger logger.Logger
}

// New creates a new API
func New(cfg *config.Config, log logger.Logger, db *repository.MongoDB) *API {
	// Create router
	r := router.New()

	// Add middleware
	r.Use(middleware.LoggingMiddleware(log))
	r.Use(middleware.ContentTypeMiddleware)

	// Initialize repositories
	userRepo := repository.NewMongoUserRepository(db)
	serviceRepo := repository.NewMongoServiceRepository(db)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(log, userRepo)
	serviceHandler := handlers.NewServiceHandler(log, serviceRepo)

	api := &API{
		router: r,
		config: cfg,
		logger: log,
	}

	// Register routes
	api.registerRoutes(userHandler, serviceHandler)

	return api
}

// registerRoutes sets up all API routes
func (a *API) registerRoutes(userHandler *handlers.UserHandler, serviceHandler *handlers.ServiceHandler) {
	// User routes
	a.router.Handle("GET", "/api/v1/users", userHandler.GetUsers)
	a.router.Handle("GET", "/api/v1/users/", userHandler.GetUsers) // Trailing slash version
	a.router.Handle("GET", "/api/v1/users/{id}", userHandler.GetUser)
	// a.router.Handle("POST", "/api/v1/users", userHandler.CreateUser)
	// a.router.Handle("PUT", "/api/v1/users/{id}", userHandler.UpdateUser)
	// a.router.Handle("DELETE", "/api/v1/users/{id}", userHandler.DeleteUser)

	// Service routes
	a.router.Handle("GET", "/api/v1/services", serviceHandler.GetServices)
	a.router.Handle("GET", "/api/v1/services/", serviceHandler.GetServices) // Trailing slash version
	a.router.Handle("GET", "/api/v1/services/{id}", serviceHandler.GetService)
	a.router.Handle("POST", "/api/v1/services", serviceHandler.CreateService)
	a.router.Handle("PUT", "/api/v1/services/{id}", serviceHandler.UpdateService)
	a.router.Handle("DELETE", "/api/v1/services/{id}", serviceHandler.DeleteService)
}

// Handler returns the HTTP handler for the API
func (a *API) Handler() http.Handler {
	return a.router
}
