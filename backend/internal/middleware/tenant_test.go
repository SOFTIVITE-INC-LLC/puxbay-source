package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestExtractSubdomain(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		host        string
		header      string
		expectedSub string
	}{
		{
			name:        "Header preferred",
			host:        "tenant1.example.com",
			header:      "tenant2",
			expectedSub: "tenant2",
		},
		{
			name:        "Fallback to host",
			host:        "tenant1.example.com",
			header:      "",
			expectedSub: "tenant1",
		},
		{
			name:        "Ignore www in host",
			host:        "www.example.com",
			header:      "",
			expectedSub: "",
		},
		{
			name:        "Ignore api in host",
			host:        "api.example.com",
			header:      "",
			expectedSub: "",
		},
		{
			name:        "No subdomain",
			host:        "example.com",
			header:      "",
			expectedSub: "",
		},
		{
			name:        "Localhost with port",
			host:        "tenant3.localhost:8080",
			header:      "",
			expectedSub: "tenant3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			req.Host = tt.host
			if tt.header != "" {
				req.Header.Set("X-Tenant-Subdomain", tt.header)
			}
			c.Request = req

			subdomain := extractSubdomain(c)
			assert.Equal(t, tt.expectedSub, subdomain)
		})
	}
}
