package handlers

import (
	"net/http"

	"github.com/timoruohomaki/open311-to-Go/internal/repository"
	"github.com/timoruohomaki/open311-to-Go/pkg/logger"
)

type ServiceRequestHandler struct {
	BaseHandler
	repo repository.ServiceRequestRepository
}

func NewServiceRequestHandler(log logger.Logger, repo repository.ServiceRequestRepository) *ServiceRequestHandler {
	return &ServiceRequestHandler{
		BaseHandler: BaseHandler{log: log},
		repo:        repo,
	}
}

// SearchServiceRequestsByFeature handles GET /api/v1/service_requests/search?featureId=...&featureGuid=...
func (h *ServiceRequestHandler) SearchServiceRequestsByFeature(w http.ResponseWriter, r *http.Request) {
	featureId := r.URL.Query().Get("featureId")
	featureGuid := r.URL.Query().Get("featureGuid")

	results, err := h.repo.FindByFeature(r.Context(), featureId, featureGuid)
	if err != nil {
		h.SendError(w, r, http.StatusInternalServerError, "Failed to search service requests")
		return
	}

	h.SendResponse(w, r, http.StatusOK, results)
}

// SearchServiceRequestsByOrganization handles GET /api/v1/service_requests/by_organization?organizationId=...
func (h *ServiceRequestHandler) SearchServiceRequestsByOrganization(w http.ResponseWriter, r *http.Request) {
	organizationId := r.URL.Query().Get("organizationId")
	if organizationId == "" {
		h.SendError(w, r, http.StatusBadRequest, "Missing organizationId parameter")
		return
	}
	results, err := h.repo.FindByOrganization(r.Context(), organizationId)
	if err != nil {
		h.SendError(w, r, http.StatusInternalServerError, "Failed to search service requests by organization")
		return
	}
	h.SendResponse(w, r, http.StatusOK, results)
}
