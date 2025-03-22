package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/timoruohomaki/open311-to-Go/api"
	"github.com/timoruohomaki/open311-to-Go/domain/models"
	"github.com/timoruohomaki/open311-to-Go/pkg/logger"
)

// UserHandler handles user-related requests
type UserHandler struct {
	BaseHandler
	repo repository.UserRepository
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(log logger.Logger, repo repository.UserRepository) *UserHandler {
	return &UserHandler{
		BaseHandler: BaseHandler{log: log},
		repo:        repo,
	}
}

// GetUsers returns all users
func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	// Get users from repository
	users, err := h.repo.FindAll(r.Context())
	if err != nil {
		h.log.Errorf("Failed to get users: %v", err)
		h.SendError(w, r, http.StatusInternalServerError, "Failed to get users")
		return
	}

	// For XML responses, wrap in Users struct
	if strings.Contains(r.Header.Get("Accept"), "application/xml") {
		h.SendResponse(w, r, http.StatusOK, models.Users{Items: users})
	} else {
		h.SendResponse(w, r, http.StatusOK, users)
	}
}

// GetUser returns a specific user
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := api.GetPathParam(r, "id")
	if id == "" {
		h.SendError(w, r, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Get user from repository
	user, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			h.SendError(w, r, http.StatusNotFound, "User not found")
		case errors.Is(err, repository.ErrInvalidID):
			h.SendError(w, r, http.StatusBadRequest, "Invalid user ID format")
		default:
			h.log.Errorf("Failed to get user: %v", err)
			h.SendError(w, r, http.StatusInternalServerError, "Failed to get user")
		}
		return
	}

	h.SendResponse(w, r, http.StatusOK, user)
}

// CreateUser creates a new user
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User

	if err := h.DecodeRequest(r, &user); err != nil {
		h.SendError(w, r, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Basic validation
	if user.Email == "" || user.FirstName == "" || user.LastName == "" {
		h.SendError(w, r, http.StatusBadRequest, "Email, first name, and last name are required")
		return
	}

	// Create user in repository
	createdUser, err := h.repo.Create(r.Context(), user)
	if err != nil {
		h.log.Errorf("Failed to create user: %v", err)
		h.SendError(w, r, http.StatusInternalServerError, "Failed to create user")
		return
	}

	h.SendResponse(w, r, http.StatusCreated, createdUser)

	h.SendResponse(w, r, http.StatusCreated, user)
}

// UpdateUser updates an existing user
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := api.GetPathParam(r, "id")
	if id == "" {
		h.SendError(w, r, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var user models.User
	if err := h.DecodeRequest(r, &user); err != nil {
		h.SendError(w, r, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Set ID from path parameter
	user.ID = id

	// Update user in repository
	updatedUser, err := h.repo.Update(r.Context(), user)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			h.SendError(w, r, http.StatusNotFound, "User not found")
		case errors.Is(err, repository.ErrInvalidID):
			h.SendError(w, r, http.StatusBadRequest, "Invalid user ID format")
		default:
			h.log.Errorf("Failed to update user: %v", err)
			h.SendError(w, r, http.StatusInternalServerError, "Failed to update user")
		}
		return
	}

	h.SendResponse(w, r, http.StatusOK, updatedUser)
}

// DeleteUser deletes a user
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := api.GetPathParam(r, "id")
	if id == "" {
		h.SendError(w, r, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Delete user from repository
	err := h.repo.Delete(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			h.SendError(w, r, http.StatusNotFound, "User not found")
		case errors.Is(err, repository.ErrInvalidID):
			h.SendError(w, r, http.StatusBadRequest, "Invalid user ID format")
		default:
			h.log.Errorf("Failed to delete user: %v", err)
			h.SendError(w, r, http.StatusInternalServerError, "Failed to delete user")
		}
		return
	}

	h.SendResponse(w, r, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}
