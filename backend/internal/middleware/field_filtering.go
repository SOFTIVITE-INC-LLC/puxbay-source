package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FieldFiltering extracts the 'fields' query parameter and stores it in the context
func FieldFiltering() gin.HandlerFunc {
	return func(c *gin.Context) {
		fieldsParam := c.Query("fields")
		if fieldsParam != "" {
			// Split by comma and trim spaces
			rawFields := strings.Split(fieldsParam, ",")
			var fields []string
			for _, f := range rawFields {
				trimmed := strings.TrimSpace(f)
				if trimmed != "" {
					// Extremely basic sanitization to prevent SQL injection in SELECT clause
					if !strings.ContainsAny(trimmed, " ;'\"-") {
						fields = append(fields, trimmed)
					}
				}
			}
			if len(fields) > 0 {
				c.Set("select_fields", fields)
			}
		}
		c.Next()
	}
}

// SelectScope is a GORM scope that applies the parsed fields to a query
func SelectScope(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		fields, ok := c.Get("select_fields")
		if ok {
			if f, isSlice := fields.([]string); isSlice && len(f) > 0 {
				return db.Select(f)
			}
		}
		return db
	}
}
