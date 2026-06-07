package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/timoruohomaki/open311-to-Go/domain/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// serviceDoc is the persistence representation of a Service. BSON tags and the
// ObjectID `_id` live here; the domain model (models.Service) stays free of
// storage concerns and exposes the id as a hex string.
type serviceDoc struct {
	ID          primitive.ObjectID        `bson:"_id,omitempty"`
	ServiceCode string                    `bson:"service_code"`
	ServiceName string                    `bson:"service_name"`
	Description string                    `bson:"description"`
	Metadata    bool                      `bson:"metadata"`
	Type        string                    `bson:"type"`
	Keywords    string                    `bson:"keywords"`
	Group       string                    `bson:"group"`
	Attributes  []models.ServiceAttribute `bson:"attributes,omitempty"`
	CreatedAt   time.Time                 `bson:"createdAt"`
	UpdatedAt   time.Time                 `bson:"updatedAt"`
}

func (d serviceDoc) toModel() models.Service {
	id := ""
	if !d.ID.IsZero() {
		id = d.ID.Hex()
	}
	return models.Service{
		ID:          id,
		ServiceCode: d.ServiceCode,
		ServiceName: d.ServiceName,
		Description: d.Description,
		Metadata:    d.Metadata,
		Type:        d.Type,
		Keywords:    d.Keywords,
		Group:       d.Group,
		Attributes:  d.Attributes,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

// MongoServiceRepository implements ServiceRepository interface using MongoDB
type MongoServiceRepository struct {
	db         *MongoDB
	collection *mongo.Collection
}

// NewMongoServiceRepository creates a new MongoServiceRepository
func NewMongoServiceRepository(db *MongoDB) ServiceRepository {
	return &MongoServiceRepository{
		db:         db,
		collection: db.GetCollection("services"),
	}
}

// FindAll retrieves all services from the database
func (r *MongoServiceRepository) FindAll(ctx context.Context) ([]models.Service, error) {
	// Create operation context with timeout
	opCtx, cancel := r.db.GetContext()
	defer cancel()

	// Use provided context if it's not nil
	if ctx != nil {
		opCtx = ctx
	}

	// Find all services
	cursor, err := r.collection.Find(opCtx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	defer cursor.Close(opCtx)

	// Decode services
	var docs []serviceDoc
	if err := cursor.All(opCtx, &docs); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	services := make([]models.Service, 0, len(docs))
	for _, d := range docs {
		services = append(services, d.toModel())
	}

	return services, nil
}

// FindByID retrieves a service by ID from the database
func (r *MongoServiceRepository) FindByID(ctx context.Context, id string) (models.Service, error) {
	// Create operation context with timeout
	opCtx, cancel := r.db.GetContext()
	defer cancel()

	// Use provided context if it's not nil
	if ctx != nil {
		opCtx = ctx
	}

	// Convert ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return models.Service{}, ErrInvalidID
	}

	// Find service by ID
	var doc serviceDoc
	err = r.collection.FindOne(opCtx, bson.M{"_id": objectID}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.Service{}, ErrNotFound
		}
		return models.Service{}, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	return doc.toModel(), nil
}

// Create adds a new service to the database
func (r *MongoServiceRepository) Create(ctx context.Context, service models.Service) (models.Service, error) {
	// Create operation context with timeout
	opCtx, cancel := r.db.GetContext()
	defer cancel()

	// Use provided context if it's not nil
	if ctx != nil {
		opCtx = ctx
	}

	// Set creation and update timestamps
	now := time.Now()
	service.CreatedAt = now
	service.UpdatedAt = now

	// Insert service. Leave _id unset (omitempty) so MongoDB generates the ObjectID.
	doc := serviceDoc{
		ServiceCode: service.ServiceCode,
		ServiceName: service.ServiceName,
		Description: service.Description,
		Metadata:    service.Metadata,
		Type:        service.Type,
		Keywords:    service.Keywords,
		Group:       service.Group,
		Attributes:  service.Attributes,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	result, err := r.collection.InsertOne(opCtx, doc)
	if err != nil {
		return models.Service{}, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	// Set ID from the generated ObjectID
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		service.ID = oid.Hex()
	}

	return service, nil
}

// Update modifies an existing service in the database
func (r *MongoServiceRepository) Update(ctx context.Context, service models.Service) (models.Service, error) {
	// Create operation context with timeout
	opCtx, cancel := r.db.GetContext()
	defer cancel()

	// Use provided context if it's not nil
	if ctx != nil {
		opCtx = ctx
	}

	// Convert ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(service.ID)
	if err != nil {
		return models.Service{}, ErrInvalidID
	}

	// Set update timestamp
	service.UpdatedAt = time.Now()

	// Update service. Field names match the serviceDoc BSON tags.
	update := bson.M{
		"$set": bson.M{
			"service_name": service.ServiceName,
			"description":  service.Description,
			"metadata":     service.Metadata,
			"type":         service.Type,
			"keywords":     service.Keywords,
			"group":        service.Group,
			"attributes":   service.Attributes,
			"updatedAt":    service.UpdatedAt,
		},
	}

	// Find and update service
	result := r.collection.FindOneAndUpdate(
		opCtx,
		bson.M{"_id": objectID},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	// Check for errors
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return models.Service{}, ErrNotFound
		}
		return models.Service{}, fmt.Errorf("%w: %v", ErrDatabase, result.Err())
	}

	// Decode updated service
	var doc serviceDoc
	if err := result.Decode(&doc); err != nil {
		return models.Service{}, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	return doc.toModel(), nil
}

// Delete removes a service from the database
func (r *MongoServiceRepository) Delete(ctx context.Context, id string) error {
	// Create operation context with timeout
	opCtx, cancel := r.db.GetContext()
	defer cancel()

	// Use provided context if it's not nil
	if ctx != nil {
		opCtx = ctx
	}

	// Convert ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidID
	}

	// Delete service
	result, err := r.collection.DeleteOne(opCtx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	// Check if service was found
	if result.DeletedCount == 0 {
		return ErrNotFound
	}

	return nil
}

// Close closes the repository
func (r *MongoServiceRepository) Close() error {
	return nil
}
