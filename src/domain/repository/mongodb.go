package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDB represents a MongoDB connection
type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
	config   struct {
		URI              string
		Database         string
		ConnectTimeout   int
		OperationTimeout int
	}
}

// NewMongoDBConnection creates a new MongoDB connection
func NewMongoDBConnection(cfg struct {
	URI              string `json:"uri"`
	Database         string `json:"database"`
	ConnectTimeout   int    `json:"connectTimeoutSeconds"`
	OperationTimeout int    `json:"operationTimeoutSeconds"`
}) (*MongoDB, error) {
	// Create MongoDB instance
	db := &MongoDB{
		config: struct {
			URI              string
			Database         string
			ConnectTimeout   int
			OperationTimeout int
		}{
			URI:              cfg.URI,
			Database:         cfg.Database,
			ConnectTimeout:   cfg.ConnectTimeout,
			OperationTimeout: cfg.OperationTimeout,
		},
	}

	// Connect to MongoDB
	if err := db.connect(); err != nil {
		return nil, err
	}

	return db, nil
}

// connect establishes a connection to MongoDB
func (db *MongoDB) connect() error {
	// Create MongoDB client options
	clientOptions := options.Client().ApplyURI(db.config.URI)

	// Set connect timeout
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(db.config.ConnectTimeout)*time.Second,
	)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping MongoDB to verify connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Set client and database
	db.client = client
	db.database = client.Database(db.config.Database)

	return nil
}

// Disconnect closes the MongoDB connection
func (db *MongoDB) Disconnect() error {
	if db.client != nil {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			time.Duration(db.config.ConnectTimeout)*time.Second,
		)
		defer cancel()

		if err := db.client.Disconnect(ctx); err != nil {
			return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
		}
	}

	return nil
}

// GetCollection returns a MongoDB collection
func (db *MongoDB) GetCollection(name string) *mongo.Collection {
	return db.database.Collection(name)
}

// GetContext returns a context with timeout for MongoDB operations
func (db *MongoDB) GetContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(
		context.Background(),
		time.Duration(db.config.OperationTimeout)*time.Second,
	)
}
