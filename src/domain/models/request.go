package models

import (
	"encoding/xml"
	"time"
)

// Request represents a request in the system
type Request struct {
	ID          string    `json:"id" xml:"id"`
	Name        string    `json:"name" xml:"name"`
	Description string    `json:"description" xml:"description"`
	Price       float64   `json:"price" xml:"price"`
	CreatedAt   time.Time `json:"createdAt" xml:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" xml:"updatedAt"`
}

// Requests is a collection of Request items for XML marshaling
type Requests struct {
	XMLName xml.Name  `xml:"requests"`
	Items   []Request `xml:"request"`
}
