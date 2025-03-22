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

// MongoUserRepository implements UserRepository interface using MongoDB
type MongoUserRepository struct {
	db         *MongoDB
	collection *mongo.Collection
}

// NewMongoUserRepository creates a new MongoUserRepository
func NewMongoUserRepository(db *MongoDB) UserRepository {
	return &MongoUserRepository{
		db:         db,
		collection: db.GetCollection("Users"),
	}
}

// FindAll retrieves all users from the database
func (r *MongoUserRepository) FindAll(ctx context.Context) ([]models.User, error) {
	// Create operation context with timeout
	opCtx, cancel := r.db.GetContext()
	defer cancel()

	// Use provided context if it's not nil
	if ctx != nil {
		opCtx = ctx
	}

	// Find all users
	cursor, err := r.collection.Find(opCtx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	defer cursor.Close(opCtx)

	// Decode users
	var users []models.User
	if err := cursor.All(opCtx, &users); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	return users, nil
}

// FindByID retrieves a user by ID from the database
func (r *MongoUserRepository) FindByID(ctx context.Context, id string) (models.User, error) {
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
		return models.User{}, ErrInvalidID
	}

	// Find user by ID
	var user models.User
	err = r.collection.FindOne(opCtx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.User{}, ErrNotFound
		}
		return models.User{}, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	return user, nil
}

// Create adds a new user to the database
func (r *MongoUserRepository) Create(ctx context.Context, user models.User) (models.User, error) {
	// Create operation context with timeout
	opCtx, cancel := r.db.GetContext()
	defer cancel()

	// Use provided context if it's not nil
	if ctx != nil {
		opCtx = ctx
	}

	// Set creation and update timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Insert user
	result, err := r.collection.InsertOne(opCtx, user)
	if err != nil {
		return models.User{}, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	// Set ID
	user.ID = result.InsertedID.(primitive.ObjectID).Hex()

	return user, nil
}

// Update modifies an existing user in the database
func (r *MongoUserRepository) Update(ctx context.Context, user models.User) (models.User, error) {
	// Create operation context with timeout
	opCtx, cancel := r.db.GetContext()
	defer cancel()

	// Use provided context if it's not nil
	if ctx != nil {
		opCtx = ctx
	}

	// Convert ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		return models.User{}, ErrInvalidID
	}

	// Set update timestamp
	user.UpdatedAt = time.Now()

	// Update user
	update := bson.M{
		"$set": bson.M{
			"email":     user.Email,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
			"updatedAt": user.UpdatedAt,
		},
	}

	// Find and update user
	result := r.collection.FindOneAndUpdate(
		opCtx,
		bson.M{"_id": objectID},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	// Check for errors
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return models.User{}, ErrNotFound
		}
		return models.User{}, fmt.Errorf("%w: %v", ErrDatabase, result.Err())
	}

	// Decode updated user
	var updatedUser models.User
	if err := result.Decode(&updatedUser); err != nil {
		return models.User{}, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	return updatedUser, nil
}

// Delete removes a user from the database
func (r *MongoUserRepository) Delete(ctx context.Context, id string) error {
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

	// Delete user
	result, err := r.collection.DeleteOne(opCtx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	// Check if user was found
	if result.DeletedCount == 0 {
		return ErrNotFound
	}

	return nil
}

// Close closes the repository
func (r *MongoUserRepository) Close() error {
	return nil
}
