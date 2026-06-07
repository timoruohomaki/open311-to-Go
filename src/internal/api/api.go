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
	router       *router.Router
	config       *config.Config
	logger       logger.Logger
	accessLogger logger.Logger
}

// New creates a new API
func New(cfg *config.Config, log logger.Logger, accessLog logger.Logger, db *repository.MongoDB) *API {
	// Create router
	r := router.New()

	// Add middleware (outermost first): access log -> rate limit -> API key -> content type
	r.Use(middleware.LoggingMiddleware(accessLog))
	r.Use(middleware.RateLimitMiddleware(cfg.RateLimit.RequestsPerMinute))
	r.Use(middleware.APIKeyMiddleware(cfg.Auth.APIKeys))
	r.Use(middleware.ContentTypeMiddleware)

	if len(cfg.Auth.APIKeys) == 0 {
		log.Warnf("API_KEYS is not set; write endpoints (POST/PUT/DELETE) are unauthenticated")
	}
	if cfg.RateLimit.RequestsPerMinute <= 0 {
		log.Infof("Rate limiting disabled (RATE_LIMIT_RPM unset)")
	} else {
		log.Infof("Rate limiting enabled: %d requests/min per client", cfg.RateLimit.RequestsPerMinute)
	}

	// Initialize repositories
	userRepo := repository.NewMongoUserRepository(db)
	serviceRepo := repository.NewMongoServiceRepository(db)
	serviceRequestRepo := repository.NewMongoServiceRequestRepository(db, cfg.MongoDB.Collection)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(log, userRepo)
	serviceHandler := handlers.NewServiceHandler(log, serviceRepo)
	serviceRequestHandler := handlers.NewServiceRequestHandler(log, serviceRequestRepo)
	healthHandler := handlers.NewHealthHandler(log, db)

	api := &API{
		router:       r,
		config:       cfg,
		logger:       log,
		accessLogger: accessLog,
	}

	// Register routes
	api.registerRoutes(userHandler, serviceHandler, serviceRequestHandler, healthHandler)

	return api
}

// registerRoutes sets up all API routes
func (a *API) registerRoutes(userHandler *handlers.UserHandler, serviceHandler *handlers.ServiceHandler, serviceRequestHandler *handlers.ServiceRequestHandler, healthHandler *handlers.HealthHandler) {
	// Health check (public, used for liveness + MongoDB connectivity)
	a.router.Handle("GET", "/health", healthHandler.Health)

	// User routes
	a.router.Handle("GET", "/open311/v2/users", userHandler.GetUsers)
	a.router.Handle("GET", "/open311/v2/users/", userHandler.GetUsers) // Trailing slash version
	a.router.Handle("GET", "/open311/v2/users/{id}", userHandler.GetUser)
	// a.router.Handle("POST", "/open311/v2/users", userHandler.CreateUser)
	// a.router.Handle("PUT", "/open311/v2/users/{id}", userHandler.UpdateUser)
	// a.router.Handle("DELETE", "/open311/v2/users/{id}", userHandler.DeleteUser)

	// Service routes (Open311 service list & definition)
	a.router.Handle("GET", "/open311/v2/services", serviceHandler.GetServices)
	a.router.Handle("GET", "/open311/v2/services/", serviceHandler.GetServices) // Trailing slash version
	a.router.Handle("GET", "/open311/v2/services/{id}", serviceHandler.GetService)
	a.router.Handle("POST", "/open311/v2/services", serviceHandler.CreateService)
	a.router.Handle("PUT", "/open311/v2/services/{id}", serviceHandler.UpdateService)
	a.router.Handle("DELETE", "/open311/v2/services/{id}", serviceHandler.DeleteService)

	// Service Request routes (Open311 GeoReport v2).
	// Register the specific sub-paths before the {id} wildcard so they win.
	a.router.Handle("GET", "/open311/v2/requests/search", serviceRequestHandler.SearchServiceRequestsByFeature)
	a.router.Handle("GET", "/open311/v2/requests/by_organization", serviceRequestHandler.SearchServiceRequestsByOrganization)
	a.router.Handle("GET", "/open311/v2/requests", serviceRequestHandler.GetServiceRequests)
	a.router.Handle("GET", "/open311/v2/requests/", serviceRequestHandler.GetServiceRequests) // Trailing slash version
	a.router.Handle("POST", "/open311/v2/requests", serviceRequestHandler.CreateServiceRequest)
	a.router.Handle("GET", "/open311/v2/requests/{id}", serviceRequestHandler.GetServiceRequest)
}

// Handler returns the HTTP handler for the API
func (a *API) Handler() http.Handler {
	return a.router
}
