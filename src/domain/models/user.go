package models

import (
	"encoding/xml"
	"time"
)

// OrgType represents the type of organization/user
// e.g., Subcontractor, Supervisor, Internal, External, etc.
type OrgType string

const (
	OrgTypeUnknown       OrgType = "unknown"
	OrgTypeSubcontractor OrgType = "subcontractor"
	OrgTypeSupervisor    OrgType = "supervisor"
	OrgTypeInternal      OrgType = "internal"
	OrgTypeExternal      OrgType = "external"
)

// Role represents the user's role in an organization
// e.g., Employee, Member, Manager, Contractor, etc.
type Role string

const (
	RoleUnknown    Role = "unknown"
	RoleEmployee   Role = "employee"
	RoleMember     Role = "member"
	RoleManager    Role = "manager"
	RoleContractor Role = "contractor"
)

// User represents a user in the system
// phone: optional phone number
// organization: organization name or identifier
// orgType: type of organization/user (see OrgType)
type User struct {
	ID            string                 `json:"id" xml:"id"`
	Email         string                 `json:"email" xml:"email"`
	FirstName     string                 `json:"firstName" xml:"firstName"`
	LastName      string                 `json:"lastName" xml:"lastName"`
	Phone         string                 `json:"phone,omitempty" xml:"phone,omitempty"`
	Organization  string                 `json:"organization,omitempty" xml:"organization,omitempty"`
	OrgType       OrgType                `json:"orgType,omitempty" xml:"orgType,omitempty"`
	Organizations []UserOrganizationLink `json:"organizations,omitempty"`
	CreatedAt     time.Time              `json:"createdAt" xml:"createdAt"`
	UpdatedAt     time.Time              `json:"updatedAt" xml:"updatedAt"`
}

// Users is a collection of User for XML marshaling
type Users struct {
	XMLName xml.Name `xml:"users"`
	Items   []User   `xml:"user"`
}

// UserOrganizationLink represents the relationship between a user and an organization
// Role: the user's role in the organization (see Role)
type UserOrganizationLink struct {
	OrganizationID string `json:"organizationId"`
	Role           Role   `json:"role"`
}
