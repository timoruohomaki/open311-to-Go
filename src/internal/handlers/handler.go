package handlers

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strings"

	"github.com/timoruohomaki/open311-to-Go/pkg/httputil"
	"github.com/timoruohomaki/open311-to-Go/pkg/logger"
)

// BaseHandler provides common handler functionality
type BaseHandler struct {
	log logger.Logger
}

// DecodeRequest decodes the request body based on Content-Type
func (h *BaseHandler) DecodeRequest(r *http.Request, v interface{}) error {
	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "application/json") {
		return json.NewDecoder(r.Body).Decode(v)
	} else if strings.Contains(contentType, "application/xml") {
		return xml.NewDecoder(r.Body).Decode(v)
	}

	return nil
}

// SendResponse sends a response in the appropriate format
func (h *BaseHandler) SendResponse(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	if err := httputil.Send(w, r, statusCode, data); err != nil {
		h.log.Errorf("Failed to send response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// SendError sends an error response
func (h *BaseHandler) SendError(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	if err := httputil.SendError(w, r, statusCode, message); err != nil {
		h.log.Errorf("Failed to send error response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
