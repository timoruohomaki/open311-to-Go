package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/timoruohomaki/open311-to-Go/domain/models"
	"github.com/timoruohomaki/open311-to-Go/domain/repository"
	"github.com/timoruohomaki/open311-to-Go/pkg/logger"
	"github.com/timoruohomaki/open311-to-Go/pkg/router"
)

// ServiceHandler handles product-related requests
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

	// For XML responses, wrap in services struct
	if strings.Contains(r.Header.Get("Accept"), "application/xml") {
		h.SendResponse(w, r, http.StatusOK, models.Services{Items: services})
	} else {
		h.SendResponse(w, r, http.StatusOK, services)
	}
}

// GetProduct returns a specific product
func (h *ServiceHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := router.GetPathParam(r, "id")
	if id == "" {
		h.SendError(w, r, http.StatusBadRequest, "Invalid product ID")
		return
	}

	// Get product from repository
	product, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			h.SendError(w, r, http.StatusNotFound, "Product not found")
		case errors.Is(err, repository.ErrInvalidID):
			h.SendError(w, r, http.StatusBadRequest, "Invalid product ID format")
		default:
			h.log.Errorf("Failed to get product: %v", err)
			h.SendError(w, r, http.StatusInternalServerError, "Failed to get product")
		}
		return
	}

	h.SendResponse(w, r, http.StatusOK, product)
}

// CreateProduct creates a new product
func (h *ServiceHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Service

	if err := h.DecodeRequest(r, &product); err != nil {
		h.SendError(w, r, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Basic validation
	if product.Name == "" || product.Price <= 0 {
		h.SendError(w, r, http.StatusBadRequest, "Name and a positive price are required")
		return
	}

	// Create product in repository
	createdProduct, err := h.repo.Create(r.Context(), product)
	if err != nil {
		h.log.Errorf("Failed to create product: %v", err)
		h.SendError(w, r, http.StatusInternalServerError, "Failed to create product")
		return
	}

	h.SendResponse(w, r, http.StatusCreated, createdProduct)
}

// UpdateProduct updates an existing product
func (h *ServiceHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := router.GetPathParam(r, "id")
	if id == "" {
		h.SendError(w, r, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var product models.Service
	if err := h.DecodeRequest(r, &product); err != nil {
		h.SendError(w, r, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Set ID from path parameter
	product.ID = id

	// Update product in repository
	updatedProduct, err := h.repo.Update(r.Context(), product)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			h.SendError(w, r, http.StatusNotFound, "Product not found")
		case errors.Is(err, repository.ErrInvalidID):
			h.SendError(w, r, http.StatusBadRequest, "Invalid product ID format")
		default:
			h.log.Errorf("Failed to update product: %v", err)
			h.SendError(w, r, http.StatusInternalServerError, "Failed to update product")
		}
		return
	}

	h.SendResponse(w, r, http.StatusOK, updatedProduct)
}

// DeleteService deletes a service
func (h *ServiceHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	id := router.GetPathParam(r, "id")
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
