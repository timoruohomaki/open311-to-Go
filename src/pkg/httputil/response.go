// pkg/httputil/response.go
package httputil

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strings"
)

// wantsXML reports whether the client asked for XML via the Accept header.
func wantsXML(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "application/xml")
}

// Send writes data directly (no envelope) in the format chosen by the Accept
// header. The data value carries its own json/xml struct tags.
func Send(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) error {
	if wantsXML(r) {
		return SendXML(w, statusCode, data)
	}
	return SendJSON(w, statusCode, data)
}

// SendJSON writes data as JSON.
func SendJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

// SendXML writes data as XML.
func SendXML(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(statusCode)
	return xml.NewEncoder(w).Encode(data)
}

// APIError is a single error entry in the Open311 errors document.
type APIError struct {
	XMLName     xml.Name `json:"-" xml:"error"`
	Code        int      `json:"code" xml:"code"`
	Description string   `json:"description" xml:"description"`
}

// APIErrors is the Open311 GeoReport v2 error envelope:
// {"errors":[{"code":400,"description":"..."}]} / <errors><error>...</error></errors>.
type APIErrors struct {
	XMLName xml.Name   `json:"-" xml:"errors"`
	Errors  []APIError `json:"errors" xml:"error"`
}

// SendError writes an error response in the Open311 errors format.
func SendError(w http.ResponseWriter, r *http.Request, statusCode int, message string) error {
	payload := APIErrors{Errors: []APIError{{Code: statusCode, Description: message}}}
	if wantsXML(r) {
		return SendXML(w, statusCode, payload)
	}
	return SendJSON(w, statusCode, payload)
}
