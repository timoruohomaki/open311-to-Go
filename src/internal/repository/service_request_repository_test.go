package repository

import (
	"context"
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

func (m *mockServiceRequestRepo) FindByOrganization(ctx context.Context, organizationID string) ([]models.ServiceRequest, error) {
	var results []models.ServiceRequest
	for _, req := range m.data {
		if req.OrganizationID == organizationID {
			results = append(results, req)
		}
	}
	return results, nil
}

func TestFindByFeature(t *testing.T) {
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

	t.Run("find by featureId", func(t *testing.T) {
		results, err := repo.FindByFeature(context.Background(), featureID, "")
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, featureID, *results[0].FeatureID)
	})

	t.Run("find by featureGuid", func(t *testing.T) {
		results, err := repo.FindByFeature(context.Background(), "", featureGuid)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, featureGuid, *results[0].FeatureGuid)
	})

	t.Run("find by both", func(t *testing.T) {
		results, err := repo.FindByFeature(context.Background(), featureID, featureGuid)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, featureID, *results[0].FeatureID)
		assert.Equal(t, featureGuid, *results[0].FeatureGuid)
	})

	t.Run("find by neither", func(t *testing.T) {
		results, err := repo.FindByFeature(context.Background(), "", "")
		assert.NoError(t, err)
		assert.Len(t, results, 2)
	})
}

func TestFindByOrganization(t *testing.T) {
	org1 := "org-1"
	org2 := "org-2"
	repo := &mockServiceRequestRepo{
		data: []models.ServiceRequest{
			{ID: "1", OrganizationID: org1},
			{ID: "2", OrganizationID: org2},
			{ID: "3", OrganizationID: org1},
		},
	}
	t.Run("find by org1", func(t *testing.T) {
		results, err := repo.FindByOrganization(context.Background(), org1)
		assert.NoError(t, err)
		assert.Len(t, results, 2)
	})
	t.Run("find by org2", func(t *testing.T) {
		results, err := repo.FindByOrganization(context.Background(), org2)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
	})
	t.Run("find by unknown org", func(t *testing.T) {
		results, err := repo.FindByOrganization(context.Background(), "org-x")
		assert.NoError(t, err)
		assert.Len(t, results, 0)
	})
}
