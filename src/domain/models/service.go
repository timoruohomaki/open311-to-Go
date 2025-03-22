package models

import (
	"encoding/xml"
	"time"
)

// Service represents a service in the system
type Service struct {
	ID          string    `json:"id" xml:"id"`
	Name        string    `json:"name" xml:"name"`
	Description string    `json:"description" xml:"description"`
	Price       float64   `json:"price" xml:"price"`
	CreatedAt   time.Time `json:"createdAt" xml:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" xml:"updatedAt"`
}

// Services is a collection of Service for XML marshaling
type services struct {
	XMLName xml.Name  `xml:"services"`
	Items   []Service `xml:"service"`
}
