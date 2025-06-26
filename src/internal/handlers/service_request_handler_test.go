package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/timoruohomaki/open311-to-Go/domain/models"
)

type mockServiceRequestRepo struct {
	data []models.ServiceRequest
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
