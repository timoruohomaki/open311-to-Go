package models

import (
	"encoding/json"
	"encoding/xml"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserFieldsAndMarshaling(t *testing.T) {
	now := time.Now().UTC()
	user := User{
		ID:           "u1",
		Email:        "test@example.com",
		FirstName:    "Test",
		LastName:     "User",
		Phone:        "+1234567890",
		Organization: "Acme Corp",
		OrgType:      OrgTypeSubcontractor,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	t.Run("marshal to JSON", func(t *testing.T) {
		data, err := json.Marshal(user)
		assert.NoError(t, err)
		var out map[string]interface{}
		json.Unmarshal(data, &out)
		assert.Equal(t, "u1", out["id"])
		assert.Equal(t, "test@example.com", out["email"])
		assert.Equal(t, "+1234567890", out["phone"])
		assert.Equal(t, "Acme Corp", out["organization"])
		assert.Equal(t, string(OrgTypeSubcontractor), out["orgType"])
	})

	t.Run("marshal to XML", func(t *testing.T) {
		data, err := xml.Marshal(user)
		assert.NoError(t, err)
		assert.Contains(t, string(data), "<phone>+1234567890</phone>")
		assert.Contains(t, string(data), "<organization>Acme Corp</organization>")
		assert.Contains(t, string(data), "<orgType>subcontractor</orgType>")
	})

	t.Run("unmarshal OrgType", func(t *testing.T) {
		var u User
		jsonStr := `{"orgType":"supervisor"}`
		err := json.Unmarshal([]byte(jsonStr), &u)
		assert.NoError(t, err)
		assert.Equal(t, OrgTypeSupervisor, u.OrgType)
	})

	t.Run("default OrgType is unknown", func(t *testing.T) {
		var u User
		assert.Equal(t, OrgTypeUnknown, u.OrgType)
	})
}
