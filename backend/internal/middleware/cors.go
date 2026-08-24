package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/logger"
)

// CORSMiddleware configures Cross-Origin Resource Sharing.
func CORSMiddleware(cfg *config.CORSConfig) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// Only allow localhost origins in non-production environments
			env := os.Getenv("APP_ENV")
			isProduction := env == "production" || env == "staging"

			if !isProduction {
				if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
					return true
				}
			}

			// Check the exact allowed origins list and wildcard subdomains
			for _, allowed := range cfg.AllowedOrigins {
				// Exact match
				if origin == allowed {
					return true
				}

				if strings.Contains(origin, "puxbay.com") || strings.Contains(origin, ".puxbay.com") {
					return true
				}

				// Subdomain match (e.g. if allowed is "https://puxbay.com", allow "https://tenant1.puxbay.com")
				allowedDomain := strings.TrimPrefix(allowed, "https://")
				allowedDomain = strings.TrimPrefix(allowedDomain, "http://")

				originDomain := strings.TrimPrefix(origin, "https://")
				originDomain = strings.TrimPrefix(originDomain, "http://")

				if strings.HasSuffix(originDomain, "."+allowedDomain) {
					return true
				}
			}
			return false
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Tenant-ID", "X-Tenant-Subdomain", "X-CSRF-Token", "X-Session-ID", "Idempotency-Key", "X-Branch-ID"},
		ExposeHeaders:    []string{"Content-Length", "Content-Disposition", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// SecurityHeaders adds security headers to all responses.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(self)")

		// CSP — strict configuration for API and static assets
		csp := strings.Join([]string{
			"default-src 'self'",
			"script-src 'self'",
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
			"font-src 'self' https://fonts.gstatic.com",
			"img-src 'self' data: blob: https:",
			"connect-src 'self' wss: https://api.stripe.com",
			"object-src 'none'",
			"base-uri 'self'",
			"frame-ancestors 'none'",
		}, "; ")
		c.Header("Content-Security-Policy", csp)

		c.Next()
	}
}

// RequestLogger logs incoming requests with timing.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// Skip health check logs to avoid noise
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/health/ready" || c.Request.URL.Path == "/healthz" || c.Request.URL.Path == "/readyz" {
			return
		}

		if statusCode >= 500 {
			logger.Log.Error("HTTP Request",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", statusCode,
				"duration", duration.String(),
				"ip", c.ClientIP(),
			)
		} else {
			logger.Log.Info("HTTP Request",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", statusCode,
				"duration", duration.String(),
				"ip", c.ClientIP(),
			)
		}
	}
}
