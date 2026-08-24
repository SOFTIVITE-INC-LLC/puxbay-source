package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/softivite/puxbay/internal/middleware"
)

func TestRoleMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()

	// Mock setting the role in the context
	r.Use(func(c *gin.Context) {
		role := c.GetHeader("X-Mock-Role")
		if role != "" {
			c.Set(middleware.ContextKeyRole, role)
		}
		c.Next()
	})

	r.GET("/admin-only", middleware.RoleMiddleware("admin"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	t.Run("Allowed Role", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/admin-only", nil)
		req.Header.Set("X-Mock-Role", "admin")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status OK, got %v", w.Code)
		}
	})

	t.Run("Denied Role", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/admin-only", nil)
		req.Header.Set("X-Mock-Role", "user")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status Forbidden, got %v", w.Code)
		}
	})

	t.Run("No Role", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/admin-only", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status Forbidden, got %v", w.Code)
		}
	})
}
