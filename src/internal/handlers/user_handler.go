package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/timoruohomaki/open311-to-Go/domain/models"
	"github.com/timoruohomaki/open311-to-Go/internal/repository"
	"github.com/timoruohomaki/open311-to-Go/pkg/httputil"
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
	id := httputil.GetPathParam(r, "id")
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
