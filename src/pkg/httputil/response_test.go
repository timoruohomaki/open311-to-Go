package httputil

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWantsXML(t *testing.T) {
	cases := []struct {
		name   string
		accept string
		want   bool
	}{
		{"no accept header", "", false},
		{"json", "application/json", false},
		{"wildcard", "*/*", false},
		{"explicit application/xml", "application/xml", true},
		{"explicit text/xml", "text/xml", true},
		{"browser (text/html + xml)", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/health", nil)
			if tc.accept != "" {
				r.Header.Set("Accept", tc.accept)
			}
			assert.Equal(t, tc.want, WantsXML(r))
		})
	}
}
