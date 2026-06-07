package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/timoruohomaki/open311-to-Go/domain/models"
	"github.com/timoruohomaki/open311-to-Go/internal/repository"
	"github.com/timoruohomaki/open311-to-Go/pkg/router"
)

// withPathParam injects a router path parameter into the request context, the
// same way the router does at runtime, so handlers reading GetPathParam work.
func withPathParam(r *http.Request, key, value string) *http.Request {
	ctx := context.WithValue(r.Context(), router.PathParamKey(key), value)
	return r.WithContext(ctx)
}

type mockServiceRequestRepo struct {
	data    []models.ServiceRequest
	created []models.ServiceRequest
}

func (m *mockServiceRequestRepo) Find(ctx context.Context, q repository.ServiceRequestQuery) ([]models.ServiceRequest, error) {
	return m.data, nil
}

func (m *mockServiceRequestRepo) FindByServiceRequestID(ctx context.Context, id string) (models.ServiceRequest, error) {
	for _, req := range m.data {
		if req.ServiceRequestID == id {
			return req, nil
		}
	}
	return models.ServiceRequest{}, repository.ErrNotFound
}

func (m *mockServiceRequestRepo) Create(ctx context.Context, req models.ServiceRequest) (models.ServiceRequest, error) {
	if req.ServiceRequestID == "" {
		req.ServiceRequestID = "generated-id"
	}
	if req.Status == "" {
		req.Status = "open"
	}
	m.created = append(m.created, req)
	return req, nil
}

func (m *mockServiceRequestRepo) FindByFeature(ctx context.Context, featureID, featureGuid string) ([]models.ServiceRequest, error) {
	var results []models.ServiceRequest
	for _, req := range m.data {
		matchID := featureID == "" || (req.FeatureID != nil && *req.FeatureID == featureID)
		matchGuid := featureGuid == "" || (req.FeatureGuid != nil && *req.FeatureGuid == featureGuid)
		if matchID && matchGuid {
			results = append(results, req)
		}
	}
	return results, nil
}

func (m *mockServiceRequestRepo) FindByOrganization(ctx context.Context, organizationID string) ([]models.ServiceRequest, error) {
	var results []models.ServiceRequest
	for _, req := range m.data {
		if req.OrganizationID == organizationID {
			results = append(results, req)
		}
	}
	return results, nil
}

func TestSearchServiceRequestsByFeature(t *testing.T) {
	featureID := "https://example.com/ogcapi/collections/parks/items/park-42"
	featureGuid := "park-42"
	otherFeatureID := "https://example.com/ogcapi/collections/parks/items/park-99"
	otherFeatureGuid := "park-99"

	repo := &mockServiceRequestRepo{
		data: []models.ServiceRequest{
			{FeatureID: &featureID, FeatureGuid: &featureGuid},
			{FeatureID: &otherFeatureID, FeatureGuid: &otherFeatureGuid},
		},
	}
	handler := NewServiceRequestHandler(nil, repo)

	t.Run("find by featureId", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/service_requests/search?featureId="+featureID, nil)
		w := httptest.NewRecorder()
		handler.SearchServiceRequestsByFeature(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		var results []models.ServiceRequest
		json.NewDecoder(w.Body).Decode(&results)
		assert.Len(t, results, 1)
		assert.Equal(t, featureID, *results[0].FeatureID)
	})

	t.Run("find by featureGuid", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/service_requests/search?featureGuid="+featureGuid, nil)
		w := httptest.NewRecorder()
		handler.SearchServiceRequestsByFeature(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		var results []models.ServiceRequest
		json.NewDecoder(w.Body).Decode(&results)
		assert.Len(t, results, 1)
		assert.Equal(t, featureGuid, *results[0].FeatureGuid)
	})

	t.Run("find by both", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/service_requests/search?featureId="+featureID+"&featureGuid="+featureGuid, nil)
		w := httptest.NewRecorder()
		handler.SearchServiceRequestsByFeature(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		var results []models.ServiceRequest
		json.NewDecoder(w.Body).Decode(&results)
		assert.Len(t, results, 1)
		assert.Equal(t, featureID, *results[0].FeatureID)
		assert.Equal(t, featureGuid, *results[0].FeatureGuid)
	})

	t.Run("find by neither", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/service_requests/search", nil)
		w := httptest.NewRecorder()
		handler.SearchServiceRequestsByFeature(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		var results []models.ServiceRequest
		json.NewDecoder(w.Body).Decode(&results)
		assert.Len(t, results, 2)
	})
}

func TestSearchServiceRequestsByOrganization(t *testing.T) {
	org1 := "org-1"
	org2 := "org-2"
	repo := &mockServiceRequestRepo{
		data: []models.ServiceRequest{
			{ID: "1", OrganizationID: org1},
			{ID: "2", OrganizationID: org2},
			{ID: "3", OrganizationID: org1},
		},
	}
	handler := NewServiceRequestHandler(nil, repo)

	t.Run("find by org1", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/service_requests/by_organization?organizationId="+org1, nil)
		w := httptest.NewRecorder()
		handler.SearchServiceRequestsByOrganization(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		var results []models.ServiceRequest
		json.NewDecoder(w.Body).Decode(&results)
		assert.Len(t, results, 2)
	})

	t.Run("find by org2", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/service_requests/by_organization?organizationId="+org2, nil)
		w := httptest.NewRecorder()
		handler.SearchServiceRequestsByOrganization(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		var results []models.ServiceRequest
		json.NewDecoder(w.Body).Decode(&results)
		assert.Len(t, results, 1)
	})

	t.Run("missing organizationId", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/service_requests/by_organization", nil)
		w := httptest.NewRecorder()
		handler.SearchServiceRequestsByOrganization(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetServiceRequest(t *testing.T) {
	repo := &mockServiceRequestRepo{
		data: []models.ServiceRequest{
			{ServiceRequestID: "sr-1", ServiceCode: "POTHOLE"},
		},
	}
	handler := NewServiceRequestHandler(nil, repo)

	t.Run("found", func(t *testing.T) {
		r := withPathParam(httptest.NewRequest(http.MethodGet, "/open311/v2/requests/sr-1", nil), "id", "sr-1")
		w := httptest.NewRecorder()
		handler.GetServiceRequest(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		r := withPathParam(httptest.NewRequest(http.MethodGet, "/open311/v2/requests/missing", nil), "id", "missing")
		w := httptest.NewRecorder()
		handler.GetServiceRequest(w, r)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestCreateServiceRequest(t *testing.T) {
	handler := NewServiceRequestHandler(nil, &mockServiceRequestRepo{})

	jsonReq := func(body string) *http.Request {
		r := httptest.NewRequest(http.MethodPost, "/open311/v2/requests", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		return r
	}

	t.Run("valid with lat/long", func(t *testing.T) {
		w := httptest.NewRecorder()
		handler.CreateServiceRequest(w, jsonReq(`{"service_code":"POTHOLE","lat":42.36,"long":-71.05}`))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("valid with address", func(t *testing.T) {
		w := httptest.NewRecorder()
		handler.CreateServiceRequest(w, jsonReq(`{"service_code":"POTHOLE","address":"1 City Hall Sq"}`))
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("missing service_code", func(t *testing.T) {
		w := httptest.NewRecorder()
		handler.CreateServiceRequest(w, jsonReq(`{"lat":42.36,"long":-71.05}`))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing location", func(t *testing.T) {
		w := httptest.NewRecorder()
		handler.CreateServiceRequest(w, jsonReq(`{"service_code":"POTHOLE"}`))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetServiceRequests(t *testing.T) {
	repo := &mockServiceRequestRepo{
		data: []models.ServiceRequest{
			{ServiceRequestID: "sr-1"},
			{ServiceRequestID: "sr-2"},
		},
	}
	handler := NewServiceRequestHandler(nil, repo)

	r := httptest.NewRequest(http.MethodGet, "/open311/v2/requests?status=open", nil)
	w := httptest.NewRecorder()
	handler.GetServiceRequests(w, r)
	assert.Equal(t, http.StatusOK, w.Code)

	var results []models.ServiceRequest
	json.NewDecoder(w.Body).Decode(&results)
	assert.Len(t, results, 2)
}
