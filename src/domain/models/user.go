package models

import (
	"encoding/xml"
	"time"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id" xml:"id"`
	Email     string    `json:"email" xml:"email"`
	FirstName string    `json:"firstName" xml:"firstName"`
	LastName  string    `json:"lastName" xml:"lastName"`
	CreatedAt time.Time `json:"createdAt" xml:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" xml:"updatedAt"`
}

// Users is a collection of User for XML marshaling
type Users struct {
	XMLName xml.Name `xml:"users"`
	Items   []User   `xml:"user"`
}
