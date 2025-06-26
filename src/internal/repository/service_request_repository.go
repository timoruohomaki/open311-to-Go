package repository

import (
	"context"

	"github.com/timoruohomaki/open311-to-Go/domain/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ServiceRequestRepository interface {
	FindByFeature(ctx context.Context, featureID, featureGuid string) ([]models.ServiceRequest, error)
	FindByOrganization(ctx context.Context, organizationID string) ([]models.ServiceRequest, error)
}

type MongoServiceRequestRepository struct {
	db         *MongoDB
	collection *mongo.Collection
}

func NewMongoServiceRequestRepository(db *MongoDB) ServiceRequestRepository {
	return &MongoServiceRequestRepository{
		db:         db,
		collection: db.GetCollection("service_requests"),
	}
}

func (r *MongoServiceRequestRepository) FindByFeature(ctx context.Context, featureID, featureGuid string) ([]models.ServiceRequest, error) {
	filter := bson.M{}
	if featureID != "" {
		filter["featureid"] = featureID
	}
	if featureGuid != "" {
		filter["featureguid"] = featureGuid
	}
	cur, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var results []models.ServiceRequest
	for cur.Next(ctx) {
		var req models.ServiceRequest
		if err := cur.Decode(&req); err != nil {
			continue
		}
		results = append(results, req)
	}
	return results, nil
}

func (r *MongoServiceRequestRepository) FindByOrganization(ctx context.Context, organizationID string) ([]models.ServiceRequest, error) {
	filter := bson.M{"organizationid": organizationID}
	cur, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var results []models.ServiceRequest
	for cur.Next(ctx) {
		var req models.ServiceRequest
		if err := cur.Decode(&req); err != nil {
			continue
		}
		results = append(results, req)
	}
	return results, nil
}
