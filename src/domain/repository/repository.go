package repository

import (
	"context"
	"errors"

	"github.com/timoruohomaki/open311-to-Go/domain/models"
)

var (
	// ErrNotFound is returned when an entity is not found
	ErrNotFound = errors.New("entity not found")
	// ErrInvalidID is returned when an invalid ID is provided
	ErrInvalidID = errors.New("invalid id")
	// ErrDatabase is returned when a database error occurs
	ErrDatabase = errors.New("database error")
)

// Repository is a generic repository interface
type Repository interface {
	Close() error
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	Repository
	FindAll(ctx context.Context) ([]models.User, error)
	FindByID(ctx context.Context, id string) (models.User, error)
	Create(ctx context.Context, user models.User) (models.User, error)
	Update(ctx context.Context, user models.User) (models.User, error)
	Delete(ctx context.Context, id string) error
}

// ServiceRepository defines the interface for service data access
type ServiceRepository interface {
	Repository
	FindAll(ctx context.Context) ([]models.Service, error)
	FindByID(ctx context.Context, id string) (models.Service, error)
	Create(ctx context.Context, product models.Service) (models.Service, error)
	Update(ctx context.Context, product models.Service) (models.Service, error)
	Delete(ctx context.Context, id string) error
}
