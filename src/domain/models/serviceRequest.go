package models

import (
	"encoding/xml"
	"sort"
	"time"
)

// Properties is an open set of jurisdiction-specific key/value pairs carried on
// a service request (the inline "properties" extension). It holds fields that
// have no Open311 equivalent — e.g. Boston's assigned_team, closure_reason,
// districts, ward — and PSK 5970 case/event annotations.
//
// JSON marshals as an object; XML as <properties><property key="k">v</property>…
type Properties map[string]string

type propertyXML struct {
	XMLName xml.Name `xml:"property"`
	Key     string   `xml:"key,attr"`
	Value   string   `xml:",chardata"`
}

// MarshalXML renders the map as a stable, ordered list of <property> elements.
func (p Properties) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(p) == 0 {
		return nil
	}
	start.Name = xml.Name{Local: "properties"}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	keys := make([]string, 0, len(p))
	for k := range p {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if err := e.Encode(propertyXML{Key: k, Value: p[k]}); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}

// UnmarshalXML reads <property key="k">v</property> children into the map.
func (p *Properties) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	m := Properties{}
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "property" {
				var prop propertyXML
				if err := d.DecodeElement(&prop, &t); err != nil {
					return err
				}
				m[prop.Key] = prop.Value
			}
		case xml.EndElement:
			if t.Name == start.Name {
				*p = m
				return nil
			}
		}
	}
}

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
	OrganizationID    string    `json:"organizationId,omitempty" xml:"organization_id,omitempty"`
	// Properties carries jurisdiction-specific fields with no Open311 equivalent
	// (e.g. Boston extras) and PSK 5970 annotations. See dictionaries/.
	Properties Properties `json:"properties,omitempty" xml:"properties,omitempty"`
}

// Requests is a collection of Request items for XML marshaling
type ServiceRequests struct {
	XMLName xml.Name         `xml:"requests"`
	Items   []ServiceRequest `xml:"request"`
}
