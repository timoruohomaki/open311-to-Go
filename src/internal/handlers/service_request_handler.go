package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/timoruohomaki/open311-to-Go/domain/models"
	"github.com/timoruohomaki/open311-to-Go/internal/repository"
	"github.com/timoruohomaki/open311-to-Go/pkg/httputil"
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

// GetServiceRequests handles GET /open311/v2/requests — list with Open311
// filters (service_request_id, service_code, status, start_date/end_date),
// Boston extensions (q, updated_after/before, page/per_page), and this project's
// feature/organization extensions.
func (h *ServiceRequestHandler) GetServiceRequests(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	query := repository.ServiceRequestQuery{
		ServiceRequestIDs: splitCSV(q.Get("service_request_id")),
		ServiceCodes:      splitCSV(q.Get("service_code")),
		Statuses:          splitCSV(q.Get("status")),
		Q:                 q.Get("q"),
		FeatureID:         q.Get("featureId"),
		FeatureGuid:       q.Get("featureGuid"),
		OrganizationID:    q.Get("organizationId"),
		Page:              atoiDefault(q.Get("page"), 0),
		PerPage:           atoiDefault(q.Get("per_page"), 0),
	}

	var err error
	if query.StartDate, err = parseTimeParam(q.Get("start_date")); err != nil {
		h.SendError(w, r, http.StatusBadRequest, "invalid start_date (expected ISO 8601)")
		return
	}
	if query.EndDate, err = parseTimeParam(q.Get("end_date")); err != nil {
		h.SendError(w, r, http.StatusBadRequest, "invalid end_date (expected ISO 8601)")
		return
	}
	if query.UpdatedAfter, err = parseTimeParam(q.Get("updated_after")); err != nil {
		h.SendError(w, r, http.StatusBadRequest, "invalid updated_after (expected ISO 8601)")
		return
	}
	if query.UpdatedBefore, err = parseTimeParam(q.Get("updated_before")); err != nil {
		h.SendError(w, r, http.StatusBadRequest, "invalid updated_before (expected ISO 8601)")
		return
	}

	results, err := h.repo.Find(r.Context(), query)
	if err != nil {
		h.log.Errorf("Failed to list service requests: %v", err)
		h.SendError(w, r, http.StatusInternalServerError, "Failed to list service requests")
		return
	}

	h.sendServiceRequests(w, r, results)
}

// GetServiceRequest handles GET /open311/v2/requests/{id} where id is the
// service_request_id.
func (h *ServiceRequestHandler) GetServiceRequest(w http.ResponseWriter, r *http.Request) {
	id := httputil.GetPathParam(r, "id")
	if id == "" {
		h.SendError(w, r, http.StatusBadRequest, "Missing service_request_id")
		return
	}

	req, err := h.repo.FindByServiceRequestID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			h.SendError(w, r, http.StatusNotFound, "Service request not found")
		default:
			h.log.Errorf("Failed to get service request: %v", err)
			h.SendError(w, r, http.StatusInternalServerError, "Failed to get service request")
		}
		return
	}

	// GeoReport returns service requests as a collection even for a single id.
	h.sendServiceRequests(w, r, []models.ServiceRequest{req})
}

// CreateServiceRequest handles POST /open311/v2/requests. Accepts JSON or XML
// (this API does not accept GeoReport's form-urlencoded bodies).
func (h *ServiceRequestHandler) CreateServiceRequest(w http.ResponseWriter, r *http.Request) {
	var req models.ServiceRequest
	if err := h.DecodeRequest(r, &req); err != nil {
		h.SendError(w, r, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.ServiceCode == "" {
		h.SendError(w, r, http.StatusBadRequest, "service_code is required")
		return
	}
	// Location is required: lat+long, or address, or address_id.
	hasLatLong := req.Latitude != 0 && req.Longitude != 0
	if !hasLatLong && req.Address == "" && req.AddressID == "" {
		h.SendError(w, r, http.StatusBadRequest, "a location is required: provide lat and long, address, or address_id")
		return
	}

	created, err := h.repo.Create(r.Context(), req)
	if err != nil {
		h.log.Errorf("Failed to create service request: %v", err)
		h.SendError(w, r, http.StatusInternalServerError, "Failed to create service request")
		return
	}

	h.sendServiceRequests(w, r, []models.ServiceRequest{created}, http.StatusCreated)
}

// UpsertServiceRequest handles PUT /open311/v2/requests/{id} where id is the
// service_request_id. It inserts the request if absent or fully replaces it if
// present (idempotent), making bulk feeds re-runnable. Unlike POST, a supplied
// updated_datetime is preserved (defaulting to now only when absent), so the
// source's own update/close timestamps survive. The URL id is authoritative and
// overrides any service_request_id in the body. Returns 201 when created, 200
// when an existing request was updated. Accepts JSON or XML.
func (h *ServiceRequestHandler) UpsertServiceRequest(w http.ResponseWriter, r *http.Request) {
	id := httputil.GetPathParam(r, "id")
	if id == "" {
		h.SendError(w, r, http.StatusBadRequest, "Missing service_request_id")
		return
	}

	var req models.ServiceRequest
	if err := h.DecodeRequest(r, &req); err != nil {
		h.SendError(w, r, http.StatusBadRequest, "Invalid request payload")
		return
	}
	// The URL is the source of truth for the natural key.
	req.ServiceRequestID = id

	if req.ServiceCode == "" {
		h.SendError(w, r, http.StatusBadRequest, "service_code is required")
		return
	}
	// Location is required: lat+long, or address, or address_id.
	hasLatLong := req.Latitude != 0 && req.Longitude != 0
	if !hasLatLong && req.Address == "" && req.AddressID == "" {
		h.SendError(w, r, http.StatusBadRequest, "a location is required: provide lat and long, address, or address_id")
		return
	}

	stored, created, err := h.repo.Upsert(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrInvalidID):
			h.SendError(w, r, http.StatusBadRequest, "Missing service_request_id")
		default:
			h.log.Errorf("Failed to upsert service request: %v", err)
			h.SendError(w, r, http.StatusInternalServerError, "Failed to upsert service request")
		}
		return
	}

	code := http.StatusOK
	if created {
		code = http.StatusCreated
	}
	h.sendServiceRequests(w, r, []models.ServiceRequest{stored}, code)
}

// DeleteServiceRequest handles DELETE /open311/v2/requests/{id} where id is the
// service_request_id. Not part of GeoReport v2; provided for administrative
// cleanup (e.g. removing test or mis-imported records). Returns 200 on success,
// 404 when the request does not exist.
func (h *ServiceRequestHandler) DeleteServiceRequest(w http.ResponseWriter, r *http.Request) {
	id := httputil.GetPathParam(r, "id")
	if id == "" {
		h.SendError(w, r, http.StatusBadRequest, "Missing service_request_id")
		return
	}

	err := h.repo.Delete(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			h.SendError(w, r, http.StatusNotFound, "Service request not found")
		case errors.Is(err, repository.ErrInvalidID):
			h.SendError(w, r, http.StatusBadRequest, "Missing service_request_id")
		default:
			h.log.Errorf("Failed to delete service request: %v", err)
			h.SendError(w, r, http.StatusInternalServerError, "Failed to delete service request")
		}
		return
	}

	h.SendResponse(w, r, http.StatusOK, MessageResponse{Message: "Service request deleted successfully"})
}

// SearchServiceRequestsByFeature handles GET /open311/v2/requests/search?featureId=...&featureGuid=...
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

// SearchServiceRequestsByOrganization handles GET /open311/v2/requests/by_organization?organizationId=...
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

// sendServiceRequests writes a list of service requests, wrapping in the XML
// collection type when the client requested XML. An optional status code
// defaults to 200.
func (h *ServiceRequestHandler) sendServiceRequests(w http.ResponseWriter, r *http.Request, results []models.ServiceRequest, status ...int) {
	code := http.StatusOK
	if len(status) > 0 {
		code = status[0]
	}
	if httputil.WantsXML(r) {
		h.SendResponse(w, r, code, models.ServiceRequests{Items: results})
		return
	}
	h.SendResponse(w, r, code, results)
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}

// parseTimeParam parses an optional ISO 8601 timestamp. Returns (nil, nil) when
// the value is empty.
func parseTimeParam(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
