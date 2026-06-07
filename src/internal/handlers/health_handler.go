package handlers

import (
	"context"
	"encoding/xml"
	"net/http"
	"time"

	"github.com/timoruohomaki/open311-to-Go/internal/repository"
	"github.com/timoruohomaki/open311-to-Go/pkg/logger"
)

// healthResponse is the /health body. It is a struct (not a map) so it marshals
// to both JSON and XML — encoding/xml cannot marshal maps.
type healthResponse struct {
	XMLName   xml.Name `json:"-" xml:"health"`
	Status    string   `json:"status" xml:"status"`
	Database  string   `json:"database" xml:"database"`
	Timestamp string   `json:"timestamp" xml:"timestamp"`
}

// HealthHandler reports service liveness and MongoDB connectivity.
type HealthHandler struct {
	BaseHandler
	db *repository.MongoDB
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(log logger.Logger, db *repository.MongoDB) *HealthHandler {
	return &HealthHandler{
		BaseHandler: BaseHandler{log: log},
		db:          db,
	}
}

// Health handles GET /health. It pings MongoDB and returns 200 when reachable,
// 503 otherwise, so it doubles as a connectivity check for load balancers.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	resp := healthResponse{
		Status:    "healthy",
		Database:  "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if err := h.db.Ping(ctx); err != nil {
		h.log.Errorf("health check: MongoDB ping failed: %v", err)
		resp.Status = "unhealthy"
		resp.Database = "unreachable"
		h.SendResponse(w, r, http.StatusServiceUnavailable, resp)
		return
	}

	h.SendResponse(w, r, http.StatusOK, resp)
}
