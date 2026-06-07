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

// userDoc is the persistence representation of a User. It carries the BSON tags
// and the ObjectID `_id`; the domain model (models.User) stays free of storage
// concerns and exposes the id as a hex string. Convert with toModel/fromModel.
type userDoc struct {
	ID            primitive.ObjectID            `bson:"_id,omitempty"`
	Email         string                        `bson:"email"`
	FirstName     string                        `bson:"firstName"`
	LastName      string                        `bson:"lastName"`
	Phone         string                        `bson:"phone,omitempty"`
	Organization  string                        `bson:"organization,omitempty"`
	OrgType       models.OrgType                `bson:"orgType,omitempty"`
	Organizations []models.UserOrganizationLink `bson:"organizations,omitempty"`
	CreatedAt     time.Time                     `bson:"createdAt"`
	UpdatedAt     time.Time                     `bson:"updatedAt"`
}

func (d userDoc) toModel() models.User {
	id := ""
	if !d.ID.IsZero() {
		id = d.ID.Hex()
	}
	return models.User{
		ID:            id,
		Email:         d.Email,
		FirstName:     d.FirstName,
		LastName:      d.LastName,
		Phone:         d.Phone,
		Organization:  d.Organization,
		OrgType:       d.OrgType,
		Organizations: d.Organizations,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
}

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
	var docs []userDoc
	if err := cursor.All(opCtx, &docs); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	users := make([]models.User, 0, len(docs))
	for _, d := range docs {
		users = append(users, d.toModel())
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
	var doc userDoc
	err = r.collection.FindOne(opCtx, bson.M{"_id": objectID}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.User{}, ErrNotFound
		}
		return models.User{}, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	return doc.toModel(), nil
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

	// Insert user. Leave _id unset (omitempty) so MongoDB generates the ObjectID.
	doc := userDoc{
		Email:         user.Email,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Phone:         user.Phone,
		Organization:  user.Organization,
		OrgType:       user.OrgType,
		Organizations: user.Organizations,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	result, err := r.collection.InsertOne(opCtx, doc)
	if err != nil {
		return models.User{}, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	// Set ID from the generated ObjectID
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		user.ID = oid.Hex()
	}

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

	// Update user. Field names match the userDoc BSON tags.
	update := bson.M{
		"$set": bson.M{
			"email":        user.Email,
			"firstName":    user.FirstName,
			"lastName":     user.LastName,
			"phone":        user.Phone,
			"organization": user.Organization,
			"orgType":      user.OrgType,
			"updatedAt":    user.UpdatedAt,
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
	var doc userDoc
	if err := result.Decode(&doc); err != nil {
		return models.User{}, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	return doc.toModel(), nil
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
