package repository

import (
	"context"
	"time"

	"github.com/timoruohomaki/open311-to-Go/domain/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ServiceRequestRepository interface {
	FindByFeature(ctx context.Context, featureID, featureGuid string) ([]models.ServiceRequest, error)
	FindByOrganization(ctx context.Context, organizationID string) ([]models.ServiceRequest, error)
}

// serviceRequestDoc is the persistence representation of a ServiceRequest. BSON
// tags and the ObjectID `_id` live here; the domain model exposes the id as a
// hex string. Field names mirror the model's JSON/XML tags.
type serviceRequestDoc struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`
	ServiceRequestID  string             `bson:"service_request_id"`
	Status            string             `bson:"status"`
	StatusNotes       string             `bson:"status_notes"`
	ServiceName       string             `bson:"service_name"`
	ServiceCode       string             `bson:"service_code"`
	Description       string             `bson:"description"`
	AgencyResponsible string             `bson:"agency_responsible"`
	ServiceNotice     string             `bson:"service_notice"`
	RequestedDatetime time.Time          `bson:"requested_datetime"`
	UpdatedDatetime   time.Time          `bson:"updated_datetime"`
	ExpectedDatetime  time.Time          `bson:"expected_datetime"`
	Address           string             `bson:"address"`
	AddressID         string             `bson:"address_id"`
	Zipcode           string             `bson:"zipcode"`
	Latitude          float64            `bson:"lat"`
	Longitude         float64            `bson:"long"`
	MediaURL          string             `bson:"media_url"`
	FeatureID         *string            `bson:"featureId,omitempty"`
	FeatureGuid       *string            `bson:"featureGuid,omitempty"`
	OrganizationID    string             `bson:"organizationId,omitempty"`
}

func (d serviceRequestDoc) toModel() models.ServiceRequest {
	id := ""
	if !d.ID.IsZero() {
		id = d.ID.Hex()
	}
	return models.ServiceRequest{
		ID:                id,
		ServiceRequestID:  d.ServiceRequestID,
		Status:            d.Status,
		StatusNotes:       d.StatusNotes,
		ServiceName:       d.ServiceName,
		ServiceCode:       d.ServiceCode,
		Description:       d.Description,
		AgencyResponsible: d.AgencyResponsible,
		ServiceNotice:     d.ServiceNotice,
		RequestedDatetime: d.RequestedDatetime,
		UpdatedDatetime:   d.UpdatedDatetime,
		ExpectedDatetime:  d.ExpectedDatetime,
		Address:           d.Address,
		AddressID:         d.AddressID,
		Zipcode:           d.Zipcode,
		Latitude:          d.Latitude,
		Longitude:         d.Longitude,
		MediaURL:          d.MediaURL,
		FeatureID:         d.FeatureID,
		FeatureGuid:       d.FeatureGuid,
		OrganizationID:    d.OrganizationID,
	}
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

func (r *MongoServiceRequestRepository) find(ctx context.Context, filter bson.M) ([]models.ServiceRequest, error) {
	cur, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var docs []serviceRequestDoc
	if err := cur.All(ctx, &docs); err != nil {
		return nil, err
	}

	results := make([]models.ServiceRequest, 0, len(docs))
	for _, d := range docs {
		results = append(results, d.toModel())
	}
	return results, nil
}

func (r *MongoServiceRequestRepository) FindByFeature(ctx context.Context, featureID, featureGuid string) ([]models.ServiceRequest, error) {
	filter := bson.M{}
	if featureID != "" {
		filter["featureId"] = featureID
	}
	if featureGuid != "" {
		filter["featureGuid"] = featureGuid
	}
	return r.find(ctx, filter)
}

func (r *MongoServiceRequestRepository) FindByOrganization(ctx context.Context, organizationID string) ([]models.ServiceRequest, error) {
	return r.find(ctx, bson.M{"organizationId": organizationID})
}
