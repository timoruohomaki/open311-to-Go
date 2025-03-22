// domain/models/service.go
package models

import (
	"encoding/xml"
	"time"
)

// Service represents a service in the Open311 system
type Service struct {
	ID          string             `json:"id" xml:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" xml:"name" bson:"name"`
	Description string             `json:"description" xml:"description" bson:"description"`
	Metadata    bool               `json:"metadata" xml:"metadata" bson:"metadata"`
	Keywords    []string           `json:"keywords,omitempty" xml:"keywords>keyword,omitempty" bson:"keywords,omitempty"`
	Group       string             `json:"group,omitempty" xml:"group,omitempty" bson:"group,omitempty"`
	Status      string             `json:"status" xml:"status" bson:"status"`
	Attributes  []ServiceAttribute `json:"attributes,omitempty" xml:"attributes>attribute,omitempty" bson:"attributes,omitempty"`
	CreatedAt   time.Time          `json:"createdAt" xml:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" xml:"updatedAt" bson:"updatedAt"`
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
	XMLName xml.Name  `xml:"services"`
	Items   []Service `xml:"service"`
}
