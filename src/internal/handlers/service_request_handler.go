package handlers

import (
	"encoding/json"
	"encoding/xml"
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

// maxBulkRequests caps how many service requests one bulk call may carry, to
// bound request size and memory. Feeders should chunk larger inputs.
const maxBulkRequests = 1000

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

// BulkItemError reports one record rejected during a bulk upsert (either by
// pre-validation or by the database).
type BulkItemError struct {
	Index            int    `json:"index" xml:"index"`
	ServiceRequestID string `json:"service_request_id" xml:"service_request_id"`
	Message          string `json:"message" xml:"message"`
}

// BulkUpsertResponse summarizes a bulk upsert. A struct (not a map) so it
// marshals to both JSON and XML.
type BulkUpsertResponse struct {
	XMLName   xml.Name        `json:"-" xml:"bulk_result"`
	Requested int             `json:"requested" xml:"requested"`
	Created   int             `json:"created" xml:"created"`
	Updated   int             `json:"updated" xml:"updated"`
	Failed    int             `json:"failed" xml:"failed"`
	Errors    []BulkItemError `json:"errors,omitempty" xml:"errors>error,omitempty"`
}

// BulkUpsertServiceRequests handles POST /open311/v2/requests/bulk — a project
// extension for re-runnable bulk feeds. Accepts a JSON array of service requests
// (or an XML <requests> document) and upserts them in a single MongoDB
// BulkWrite, keyed on service_request_id. Each record requires service_request_id,
// service_code, and a location (lat+long, address, or address_id); invalid
// records are rejected and reported without aborting the batch. Like PUT, a
// supplied updated_datetime is preserved. Returns 200 with a per-batch summary;
// 400 only when the whole payload is malformed, empty, or exceeds the cap.
func (h *ServiceRequestHandler) BulkUpsertServiceRequests(w http.ResponseWriter, r *http.Request) {
	var incoming []models.ServiceRequest

	if strings.Contains(r.Header.Get("Content-Type"), "xml") {
		var wrapper models.ServiceRequests
		if err := xml.NewDecoder(r.Body).Decode(&wrapper); err != nil {
			h.SendError(w, r, http.StatusBadRequest, "Invalid request payload")
			return
		}
		incoming = wrapper.Items
	} else {
		if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
			h.SendError(w, r, http.StatusBadRequest, "Invalid request payload (expected a JSON array of service requests)")
			return
		}
	}

	if len(incoming) == 0 {
		h.SendError(w, r, http.StatusBadRequest, "no service requests provided")
		return
	}
	if len(incoming) > maxBulkRequests {
		h.SendError(w, r, http.StatusBadRequest, "batch too large: at most "+strconv.Itoa(maxBulkRequests)+" requests per call")
		return
	}

	// Pre-validate; keep valid records, collect rejects (indexes preserved).
	valid := make([]models.ServiceRequest, 0, len(incoming))
	var rejects []BulkItemError
	for i, req := range incoming {
		switch {
		case req.ServiceRequestID == "":
			rejects = append(rejects, BulkItemError{Index: i, Message: "service_request_id is required"})
		case req.ServiceCode == "":
			rejects = append(rejects, BulkItemError{Index: i, ServiceRequestID: req.ServiceRequestID, Message: "service_code is required"})
		case req.Latitude == 0 && req.Longitude == 0 && req.Address == "" && req.AddressID == "":
			rejects = append(rejects, BulkItemError{Index: i, ServiceRequestID: req.ServiceRequestID, Message: "a location is required: provide lat and long, address, or address_id"})
		default:
			valid = append(valid, req)
		}
	}

	result, err := h.repo.BulkUpsert(r.Context(), valid)
	if err != nil {
		h.log.Errorf("Bulk upsert failed: %v", err)
		h.SendError(w, r, http.StatusInternalServerError, "Failed to bulk upsert service requests")
		return
	}

	resp := BulkUpsertResponse{
		Requested: len(incoming),
		Created:   result.Created,
		Updated:   result.Updated,
		Failed:    result.Failed + len(rejects),
	}
	for _, e := range rejects {
		resp.Errors = append(resp.Errors, e)
	}
	for _, e := range result.Errors {
		resp.Errors = append(resp.Errors, BulkItemError{Index: e.Index, ServiceRequestID: e.ServiceRequestID, Message: e.Message})
	}

	h.SendResponse(w, r, http.StatusOK, resp)
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
