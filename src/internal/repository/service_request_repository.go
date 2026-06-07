package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/timoruohomaki/open311-to-Go/domain/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ServiceRequestQuery holds the filters for listing service requests. Zero-value
// fields are ignored. Mirrors the Open311 GeoReport v2 query parameters plus the
// Boston extensions (q, updated_*) and this project's feature/organization
// extensions.
type ServiceRequestQuery struct {
	ServiceRequestIDs []string
	ServiceCodes      []string
	Statuses          []string
	StartDate         *time.Time
	EndDate           *time.Time
	UpdatedAfter      *time.Time
	UpdatedBefore     *time.Time
	Q                 string
	FeatureID         string
	FeatureGuid       string
	OrganizationID    string
	Page              int
	PerPage           int
}

type ServiceRequestRepository interface {
	Find(ctx context.Context, q ServiceRequestQuery) ([]models.ServiceRequest, error)
	FindByServiceRequestID(ctx context.Context, serviceRequestID string) (models.ServiceRequest, error)
	Create(ctx context.Context, req models.ServiceRequest) (models.ServiceRequest, error)
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

func serviceRequestDocFromModel(m models.ServiceRequest) serviceRequestDoc {
	doc := serviceRequestDoc{
		ServiceRequestID:  m.ServiceRequestID,
		Status:            m.Status,
		StatusNotes:       m.StatusNotes,
		ServiceName:       m.ServiceName,
		ServiceCode:       m.ServiceCode,
		Description:       m.Description,
		AgencyResponsible: m.AgencyResponsible,
		ServiceNotice:     m.ServiceNotice,
		RequestedDatetime: m.RequestedDatetime,
		UpdatedDatetime:   m.UpdatedDatetime,
		ExpectedDatetime:  m.ExpectedDatetime,
		Address:           m.Address,
		AddressID:         m.AddressID,
		Zipcode:           m.Zipcode,
		Latitude:          m.Latitude,
		Longitude:         m.Longitude,
		MediaURL:          m.MediaURL,
		FeatureID:         m.FeatureID,
		FeatureGuid:       m.FeatureGuid,
		OrganizationID:    m.OrganizationID,
	}
	if m.ID != "" {
		if oid, err := primitive.ObjectIDFromHex(m.ID); err == nil {
			doc.ID = oid
		}
	}
	return doc
}

type MongoServiceRequestRepository struct {
	db         *MongoDB
	collection *mongo.Collection
}

func NewMongoServiceRequestRepository(db *MongoDB, collection string) ServiceRequestRepository {
	if collection == "" {
		collection = "service_requests"
	}
	return &MongoServiceRequestRepository{
		db:         db,
		collection: db.GetCollection(collection),
	}
}

func (r *MongoServiceRequestRepository) find(ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]models.ServiceRequest, error) {
	cur, err := r.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	defer cur.Close(ctx)

	var docs []serviceRequestDoc
	if err := cur.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	results := make([]models.ServiceRequest, 0, len(docs))
	for _, d := range docs {
		results = append(results, d.toModel())
	}
	return results, nil
}

// Find lists service requests matching the query, newest first, with pagination
// (PerPage defaults to 100 and is capped at 100; Page is 1-based).
func (r *MongoServiceRequestRepository) Find(ctx context.Context, q ServiceRequestQuery) ([]models.ServiceRequest, error) {
	filter := bson.M{}

	if len(q.ServiceRequestIDs) > 0 {
		filter["service_request_id"] = bson.M{"$in": q.ServiceRequestIDs}
	}
	if len(q.ServiceCodes) > 0 {
		filter["service_code"] = bson.M{"$in": q.ServiceCodes}
	}
	if len(q.Statuses) > 0 {
		filter["status"] = bson.M{"$in": q.Statuses}
	}
	if q.FeatureID != "" {
		filter["featureId"] = q.FeatureID
	}
	if q.FeatureGuid != "" {
		filter["featureGuid"] = q.FeatureGuid
	}
	if q.OrganizationID != "" {
		filter["organizationId"] = q.OrganizationID
	}
	if dr := dateRange(q.StartDate, q.EndDate); dr != nil {
		filter["requested_datetime"] = dr
	}
	if dr := dateRange(q.UpdatedAfter, q.UpdatedBefore); dr != nil {
		filter["updated_datetime"] = dr
	}
	if q.Q != "" {
		rx := primitive.Regex{Pattern: regexp.QuoteMeta(q.Q), Options: "i"}
		filter["$or"] = bson.A{
			bson.M{"description": rx},
			bson.M{"service_name": rx},
			bson.M{"address": rx},
		}
	}

	perPage := q.PerPage
	if perPage <= 0 || perPage > 100 {
		perPage = 100
	}
	page := q.Page
	if page < 1 {
		page = 1
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "requested_datetime", Value: -1}}).
		SetLimit(int64(perPage)).
		SetSkip(int64((page - 1) * perPage))

	return r.find(ctx, filter, opts)
}

func (r *MongoServiceRequestRepository) FindByServiceRequestID(ctx context.Context, serviceRequestID string) (models.ServiceRequest, error) {
	var doc serviceRequestDoc
	err := r.collection.FindOne(ctx, bson.M{"service_request_id": serviceRequestID}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.ServiceRequest{}, ErrNotFound
		}
		return models.ServiceRequest{}, fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	return doc.toModel(), nil
}

// Create inserts a new service request, assigning a service_request_id (the new
// ObjectID hex) and timestamps when not supplied, and defaulting status to open.
func (r *MongoServiceRequestRepository) Create(ctx context.Context, req models.ServiceRequest) (models.ServiceRequest, error) {
	oid := primitive.NewObjectID()
	now := time.Now().UTC()

	if req.ServiceRequestID == "" {
		req.ServiceRequestID = oid.Hex()
	}
	if req.Status == "" {
		req.Status = "open"
	}
	if req.RequestedDatetime.IsZero() {
		req.RequestedDatetime = now
	}
	req.UpdatedDatetime = now

	doc := serviceRequestDocFromModel(req)
	doc.ID = oid

	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		return models.ServiceRequest{}, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	req.ID = oid.Hex()
	return req, nil
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

// dateRange builds a {$gte, $lte} filter fragment, or nil if both bounds are nil.
func dateRange(from, to *time.Time) bson.M {
	r := bson.M{}
	if from != nil {
		r["$gte"] = *from
	}
	if to != nil {
		r["$lte"] = *to
	}
	if len(r) == 0 {
		return nil
	}
	return r
}
