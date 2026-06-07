package repository

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/timoruohomaki/open311-to-Go/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDB represents a MongoDB connection
type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
	config   config.MongoDBConfig
}

// NewMongoDBConnection creates a new MongoDB connection
func NewMongoDBConnection(cfg config.MongoDBConfig) (*MongoDB, error) {
	db := &MongoDB{config: cfg}

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

	// For MONGODB-X509 authentication, load the client certificate (and an
	// optional CA bundle) into the TLS config. The auth mechanism and
	// authSource=$external come from the connection string itself.
	if db.config.TLSCertificateKeyFile != "" || db.config.TLSCAFile != "" {
		tlsConfig, err := buildTLSConfig(db.config.TLSCertificateKeyFile, db.config.TLSCAFile)
		if err != nil {
			return fmt.Errorf("failed to build TLS config: %w", err)
		}
		clientOptions.SetTLSConfig(tlsConfig)
	}

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

// buildTLSConfig loads an X.509 client certificate/key (combined PEM) and an
// optional CA bundle for MongoDB TLS connections.
func buildTLSConfig(certKeyFile, caFile string) (*tls.Config, error) {
	tlsConfig := &tls.Config{}

	if certKeyFile != "" {
		// The client certificate and private key may live in the same PEM file,
		// so the same path is passed for both the cert and the key.
		cert, err := tls.LoadX509KeyPair(certKeyFile, certKeyFile)
		if err != nil {
			return nil, fmt.Errorf("loading client certificate %q: %w", certKeyFile, err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if caFile != "" {
		caPEM, err := os.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("reading CA file %q: %w", caFile, err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("no certificates found in CA file %q", caFile)
		}
		tlsConfig.RootCAs = pool
	}

	return tlsConfig, nil
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

// Ping verifies the MongoDB connection is alive.
func (db *MongoDB) Ping(ctx context.Context) error {
	if db.client == nil {
		return fmt.Errorf("mongodb client is not initialized")
	}
	return db.client.Ping(ctx, readpref.Primary())
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
