package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EnsureIndexes creates the indexes the application relies on. It is idempotent:
// re-creating an existing index is a no-op. serviceRequestsCollection is the
// configured collection name for service requests (e.g. "open311-boston").
//
// Note: imported documents must carry a GeoJSON `location` field to be covered
// by the 2dsphere index; documents missing it are simply not geo-indexed.
func EnsureIndexes(ctx context.Context, db *MongoDB, serviceRequestsCollection string) error {
	if serviceRequestsCollection == "" {
		serviceRequestsCollection = "service_requests"
	}

	serviceRequestIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "service_request_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_service_request_id"),
		},
		{
			Keys:    bson.D{{Key: "location", Value: "2dsphere"}},
			Options: options.Index().SetName("geo_location"),
		},
		{Keys: bson.D{{Key: "status", Value: 1}}, Options: options.Index().SetName("status")},
		{Keys: bson.D{{Key: "organizationId", Value: 1}}, Options: options.Index().SetName("organizationId")},
		{Keys: bson.D{{Key: "featureId", Value: 1}}, Options: options.Index().SetName("featureId")},
		{Keys: bson.D{{Key: "requested_datetime", Value: -1}}, Options: options.Index().SetName("requested_datetime")},
		{Keys: bson.D{{Key: "updated_datetime", Value: -1}}, Options: options.Index().SetName("updated_datetime")},
	}
	if _, err := db.GetCollection(serviceRequestsCollection).Indexes().CreateMany(ctx, serviceRequestIndexes); err != nil {
		return fmt.Errorf("creating indexes on %q: %w", serviceRequestsCollection, err)
	}

	// services: unique service_code
	if _, err := db.GetCollection("services").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "service_code", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("uniq_service_code"),
	}); err != nil {
		return fmt.Errorf("creating indexes on \"services\": %w", err)
	}

	// users: unique email (sparse so documents without an email are allowed)
	if _, err := db.GetCollection("Users").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetSparse(true).SetName("uniq_email"),
	}); err != nil {
		return fmt.Errorf("creating indexes on \"Users\": %w", err)
	}

	return nil
}
