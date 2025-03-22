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

// ServiceHandler handles service-related requests
type ServiceHandler struct {
	BaseHandler
	repo repository.ServiceRepository
}

// NewServiceHandler creates a new ServiceHandler
func NewServiceHandler(log logger.Logger, repo repository.ServiceRepository) *ServiceHandler {
	return &ServiceHandler{
		BaseHandler: BaseHandler{log: log},
		repo:        repo,
	}
}

// GetServices returns all services
func (h *ServiceHandler) GetServices(w http.ResponseWriter, r *http.Request) {
	// Get services from repository
	services, err := h.repo.FindAll(r.Context())
	if err != nil {
		h.log.Errorf("Failed to get services: %v", err)
		h.SendError(w, r, http.StatusInternalServerError, "Failed to get services")
		return
	}

	// For XML responses, wrap in Services struct
	if strings.Contains(r.Header.Get("Accept"), "application/xml") {
		h.SendResponse(w, r, http.StatusOK, models.Services{Items: services})
	} else {
		h.SendResponse(w, r, http.StatusOK, services)
	}
}

// GetService returns a specific service
func (h *ServiceHandler) GetService(w http.ResponseWriter, r *http.Request) {
	id := httputil.GetPathParam(r, "id")
	if id == "" {
		h.SendError(w, r, http.StatusBadRequest, "Invalid service ID")
		return
	}

	// Get service from repository
	service, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			h.SendError(w, r, http.StatusNotFound, "Service not found")
		case errors.Is(err, repository.ErrInvalidID):
			h.SendError(w, r, http.StatusBadRequest, "Invalid service ID format")
		default:
			h.log.Errorf("Failed to get service: %v", err)
			h.SendError(w, r, http.StatusInternalServerError, "Failed to get service")
		}
		return
	}

	h.SendResponse(w, r, http.StatusOK, service)
}

// CreateService creates a new service
func (h *ServiceHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	var service models.Service

	if err := h.DecodeRequest(r, &service); err != nil {
		h.SendError(w, r, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Basic validation
	if service.ServiceName == "" || service.Description == "" {
		h.SendError(w, r, http.StatusBadRequest, "Name and description are required")
		return
	}

	// Create service in repository
	createdService, err := h.repo.Create(r.Context(), service)
	if err != nil {
		h.log.Errorf("Failed to create service: %v", err)
		h.SendError(w, r, http.StatusInternalServerError, "Failed to create service")
		return
	}

	h.SendResponse(w, r, http.StatusCreated, createdService)
}

// UpdateService updates an existing service
func (h *ServiceHandler) UpdateService(w http.ResponseWriter, r *http.Request) {
	id := httputil.GetPathParam(r, "id")
	if id == "" {
		h.SendError(w, r, http.StatusBadRequest, "Invalid service ID")
		return
	}

	var service models.Service
	if err := h.DecodeRequest(r, &service); err != nil {
		h.SendError(w, r, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Set ID from path parameter
	service.ID = id

	// Update service in repository
	updatedService, err := h.repo.Update(r.Context(), service)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			h.SendError(w, r, http.StatusNotFound, "Service not found")
		case errors.Is(err, repository.ErrInvalidID):
			h.SendError(w, r, http.StatusBadRequest, "Invalid service ID format")
		default:
			h.log.Errorf("Failed to update service: %v", err)
			h.SendError(w, r, http.StatusInternalServerError, "Failed to update service")
		}
		return
	}

	h.SendResponse(w, r, http.StatusOK, updatedService)
}

// DeleteService deletes a service
func (h *ServiceHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	id := httputil.GetPathParam(r, "id")
	if id == "" {
		h.SendError(w, r, http.StatusBadRequest, "Invalid service ID")
		return
	}

	// Delete service from repository
	err := h.repo.Delete(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			h.SendError(w, r, http.StatusNotFound, "Service not found")
		case errors.Is(err, repository.ErrInvalidID):
			h.SendError(w, r, http.StatusBadRequest, "Invalid service ID format")
		default:
			h.log.Errorf("Failed to delete service: %v", err)
			h.SendError(w, r, http.StatusInternalServerError, "Failed to delete service")
		}
		return
	}

	h.SendResponse(w, r, http.StatusOK, map[string]string{"message": "Service deleted successfully"})
}
