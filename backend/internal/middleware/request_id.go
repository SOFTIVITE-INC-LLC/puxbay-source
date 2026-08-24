package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID sets a unique identifier for each request and injects it into the response headers.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}

		// Set in Gin context
		c.Set("request_id", reqID)

		// Set in response header
		c.Header("X-Request-ID", reqID)

		c.Next()
	}
}
