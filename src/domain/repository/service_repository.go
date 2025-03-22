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

// MongoServiceRepository implements ProductRepository interface using MongoDB
type MongoServiceRepository struct {
	db         *MongoDB
	collection *mongo.Collection
}

// NewMongoProductRepository creates a new MongoProductRepository
func NewMongoServiceRepository(db *MongoDB) ServiceRepository {
	return &MongoServiceRepository{
		db:         db,
		collection: db.GetCollection("Services"),
	}
}

// FindAll retrieves all services from the database
func (r *MongoServiceRepository) FindAll(ctx context.Context) ([]models.Product, error) {
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
	var services []models.Product
	if err := cursor.All(opCtx, &services); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	return services, nil
}

// FindByID retrieves a product by ID from the database
func (r *MongoServiceRepository) FindByID(ctx context.Context, id string) (models.Product, error) {
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
		return models.Product{}, ErrInvalidID
	}

	// Find product by ID
	var product models.Product
	err = r.collection.FindOne(opCtx, bson.M{"_id": objectID}).Decode(&product)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.Product{}, ErrNotFound
		}
		return models.Product{}, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	return product, nil
}

// Create adds a new product to the database
func (r *MongoServiceRepository) Create(ctx context.Context, product models.Product) (models.Product, error) {
	// Create operation context with timeout
	opCtx, cancel := r.db.GetContext()
	defer cancel()

	// Use provided context if it's not nil
	if ctx != nil {
		opCtx = ctx
	}

	// Set creation and update timestamps
	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	// Insert product
	result, err := r.collection.InsertOne(opCtx, product)
	if err != nil {
		return models.Product{}, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	// Set ID
	product.ID = result.InsertedID.(primitive.ObjectID).Hex()

	return product, nil
}

// Update modifies an existing product in the database
func (r *MongoServiceRepository) Update(ctx context.Context, product models.Product) (models.Product, error) {
	// Create operation context with timeout
	opCtx, cancel := r.db.GetContext()
	defer cancel()

	// Use provided context if it's not nil
	if ctx != nil {
		opCtx = ctx
	}

	// Convert ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(product.ID)
	if err != nil {
		return models.Product{}, ErrInvalidID
	}

	// Set update timestamp
	product.UpdatedAt = time.Now()

	// Update product
	update := bson.M{
		"$set": bson.M{
			"name":        product.Name,
			"description": product.Description,
			"price":       product.Price,
			"updatedAt":   product.UpdatedAt,
		},
	}

	// Find and update product
	result := r.collection.FindOneAndUpdate(
		opCtx,
		bson.M{"_id": objectID},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	// Check for errors
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return models.Product{}, ErrNotFound
		}
		return models.Product{}, fmt.Errorf("%w: %v", ErrDatabase, result.Err())
	}

	// Decode updated product
	var updatedProduct models.Product
	if err := result.Decode(&updatedProduct); err != nil {
		return models.Product{}, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	return updatedProduct, nil
}

// Delete removes a product from the database
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

	// Delete product
	result, err := r.collection.DeleteOne(opCtx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	// Check if product was found
	if result.DeletedCount == 0 {
		return ErrNotFound
	}

	return nil
}

// Close closes the repository
func (r *MongoServiceRepository) Close() error {
	return nil
}
