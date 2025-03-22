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
	var services []models.Service
	if err := cursor.All(opCtx, &services); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
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
	var service models.Service
	err = r.collection.FindOne(opCtx, bson.M{"_id": objectID}).Decode(&service)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.Service{}, ErrNotFound
		}
		return models.Service{}, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	return service, nil
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

	// Insert service
	result, err := r.collection.InsertOne(opCtx, service)
	if err != nil {
		return models.Service{}, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	// Set ID
	service.ID = result.InsertedID.(primitive.ObjectID).Hex()

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

	// Update service
	update := bson.M{
		"$set": bson.M{
			"name":        service.ServiceName,
			"description": service.Description,
			"metadata":    service.Metadata,
			"keywords":    service.Keywords,
			"group":       service.Group,
			"updatedAt":   service.UpdatedAt,
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
	var updatedService models.Service
	if err := result.Decode(&updatedService); err != nil {
		return models.Service{}, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	return updatedService, nil
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
