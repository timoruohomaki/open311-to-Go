// pkg/httputil/response.go
package httputil

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strings"
)

// Response is a generic API response structure
type Response struct {
	Status  string      `json:"status" xml:"status"`
	Message string      `json:"message,omitempty" xml:"message,omitempty"`
	Data    interface{} `json:"data,omitempty" xml:"data,omitempty"`
}

// XMLResponse is a wrapper to make XML responses work with arbitrary data
type XMLResponse struct {
	XMLName xml.Name    `xml:"response"`
	Status  string      `xml:"status"`
	Message string      `xml:"message,omitempty"`
	Data    interface{} `xml:"data,omitempty"`
}

// Send sends the response in the format specified by the Accept header
func Send(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) error {
	acceptHeader := r.Header.Get("Accept")

	response := Response{
		Status: http.StatusText(statusCode),
		Data:   data,
	}

	if strings.Contains(acceptHeader, "application/xml") {
		return SendXML(w, statusCode, response)
	}

	return SendJSON(w, statusCode, response)
}

// SendJSON sends a JSON response
func SendJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	return json.NewEncoder(w).Encode(data)
}

// SendXML sends an XML response
func SendXML(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(statusCode)

	// Convert to XMLResponse to ensure proper XML structure
	xmlData := XMLResponse{
		Status: http.StatusText(statusCode),
	}

	// If data is already a Response, extract its fields
	if resp, ok := data.(Response); ok {
		xmlData.Status = resp.Status
		xmlData.Message = resp.Message
		xmlData.Data = resp.Data
	} else {
		xmlData.Data = data
	}

	return xml.NewEncoder(w).Encode(xmlData)
}

// SendError sends an error response
func SendError(w http.ResponseWriter, r *http.Request, statusCode int, message string) error {
	response := Response{
		Status:  http.StatusText(statusCode),
		Message: message,
	}

	acceptHeader := r.Header.Get("Accept")

	if strings.Contains(acceptHeader, "application/xml") {
		return SendXML(w, statusCode, response)
	}

	return SendJSON(w, statusCode, response)
}
