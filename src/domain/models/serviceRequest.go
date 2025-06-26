package models

import (
	"encoding/xml"
	"time"
)

// Service Request represents a request in the system

type ServiceRequest struct {
	ID                string    `json:"id" xml:"id"`
	ServiceRequestID  string    `xml:"service_request_id" json:"service_request_id"`
	Status            string    `xml:"status" json:"status"`
	StatusNotes       string    `xml:"status_notes" json:"status_notes"`
	ServiceName       string    `xml:"service_name" json:"service_name"`
	ServiceCode       string    `xml:"service_code" json:"service_code"`
	Description       string    `xml:"description" json:"description"`
	AgencyResponsible string    `xml:"agency_responsible" json:"agency_responsible"`
	ServiceNotice     string    `xml:"service_notice" json:"service_notice"`
	RequestedDatetime time.Time `xml:"requested_datetime" json:"requested_datetime"`
	UpdatedDatetime   time.Time `xml:"updated_datetime" json:"updated_datetime"`
	ExpectedDatetime  time.Time `xml:"expected_datetime" json:"expected_datetime"`
	Address           string    `xml:"address" json:"address"`
	AddressID         string    `xml:"address_id" json:"address_id"`
	Zipcode           string    `xml:"zipcode" json:"zipcode"`
	Latitude          float64   `xml:"lat" json:"lat"`
	Longitude         float64   `xml:"long" json:"long"`
	MediaURL          string    `xml:"media_url" json:"media_url"`
	FeatureID         *string   `json:"featureId,omitempty" xml:"feature_id,omitempty"`
	FeatureGuid       *string   `json:"featureGuid,omitempty" xml:"feature_guid,omitempty"`
}

// Requests is a collection of Request items for XML marshaling
type ServiceRequests struct {
	XMLName xml.Name         `xml:"requests"`
	Items   []ServiceRequest `xml:"request"`
}
