// domain/models/service.go
package models

import (
	"encoding/xml"
	"time"
)

// Service represents a service in the Open311 system
type Service struct {
	ID          string             `json:"id" xml:"id"`
	XMLName     xml.Name           `xml:"service" json:"-"`
	ServiceCode string             `xml:"service_code" json:"service_code"`
	ServiceName string             `xml:"service_name" json:"service_name"`
	Description string             `xml:"description" json:"description"`
	Metadata    bool               `xml:"metadata" json:"metadata"`
	Type        string             `xml:"type" json:"type"`
	Keywords    string             `xml:"keywords" json:"keywords"`
	Group       string             `xml:"group" json:"group"`
	Attributes  []ServiceAttribute `json:"attributes,omitempty" xml:"attributes>attribute,omitempty" bson:"attributes,omitempty"`
	CreatedAt   time.Time          `json:"createdAt" xml:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" xml:"updatedAt"`
}

// ServiceAttribute represents a custom attribute for a service
type ServiceAttribute struct {
	Variable    bool     `json:"variable" xml:"variable" bson:"variable"`
	Code        string   `json:"code" xml:"code" bson:"code"`
	DataType    string   `json:"datatype" xml:"datatype" bson:"datatype"`
	Required    bool     `json:"required" xml:"required" bson:"required"`
	Description string   `json:"description" xml:"description" bson:"description"`
	Order       int      `json:"order" xml:"order" bson:"order"`
	Values      []string `json:"values,omitempty" xml:"values>value,omitempty" bson:"values,omitempty"`
}

// Services is a collection of Service for XML marshaling
type Services struct {
	XMLName xml.Name  `xml:"services" json:"-"`
	Items   []Service `xml:"service" json:"services"`
}
