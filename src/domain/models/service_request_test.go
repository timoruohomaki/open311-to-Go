package models

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceRequestPropertiesJSON(t *testing.T) {
	sr := ServiceRequest{
		ServiceRequestID: "BCS-00059336",
		ServiceCode:      "Domestic Animal Issue",
		Status:           "closed",
		Properties: Properties{
			"closure_reason": "Resolved",
			"on_time":        "ONTIME",
			"pwd_district":   "1B",
		},
	}

	data, err := json.Marshal(sr)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"properties":`)
	assert.Contains(t, string(data), `"closure_reason":"Resolved"`)

	var out ServiceRequest
	assert.NoError(t, json.Unmarshal(data, &out))
	assert.Equal(t, "Resolved", out.Properties["closure_reason"])
	assert.Equal(t, "1B", out.Properties["pwd_district"])
}

func TestServiceRequestPropertiesXMLRoundTrip(t *testing.T) {
	sr := ServiceRequest{
		ServiceRequestID: "BCS-00059336",
		Properties: Properties{
			"closure_reason": "Resolved",
			"on_time":        "ONTIME",
		},
	}

	data, err := xml.Marshal(sr)
	assert.NoError(t, err)
	// Stable, ordered <property key="...">value</property> elements.
	assert.Contains(t, string(data), `<properties>`)
	assert.Contains(t, string(data), `<property key="closure_reason">Resolved</property>`)

	var out ServiceRequest
	assert.NoError(t, xml.Unmarshal(data, &out))
	assert.Equal(t, "Resolved", out.Properties["closure_reason"])
	assert.Equal(t, "ONTIME", out.Properties["on_time"])
}

func TestServiceRequestEmptyPropertiesOmitted(t *testing.T) {
	sr := ServiceRequest{ServiceRequestID: "BCS-1"}

	jsonData, err := json.Marshal(sr)
	assert.NoError(t, err)
	assert.NotContains(t, string(jsonData), "properties")

	xmlData, err := xml.Marshal(sr)
	assert.NoError(t, err)
	assert.NotContains(t, string(xmlData), "properties")
}
