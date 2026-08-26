package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CSRFMiddleware implements Double Submit Cookie CSRF protection.
// It ensures that authenticated mutating requests (POST/PUT/DELETE/PATCH) provide a valid X-CSRF-Token header.
func CSRFMiddleware(rootDomain string) gin.HandlerFunc {
	return func(c *gin.Context) {
		isProduction := os.Getenv("APP_ENV") == "production" || os.Getenv("APP_ENV") == "staging"

		// Use the root domain for the cookie so hq.puxbay.com can read a cookie set by api.puxbay.com
		cookieDomain := ""
		if isProduction && rootDomain != "" {
			domain := rootDomain
			if idx := strings.LastIndex(domain, ":"); idx != -1 {
				domain = domain[:idx]
			}
			cookieDomain = "." + domain
		}

		cookie, err := c.Cookie("csrf_token")
		if err != nil || cookie == "" {
			b := make([]byte, 32)
			_, _ = rand.Read(b)
			cookie = hex.EncodeToString(b)

			c.SetCookie("csrf_token", cookie, 86400, "/", cookieDomain, isProduction, false) // NOT HttpOnly so JS can read it
		}

		// Always expose the token so the frontend can read it from the response header
		c.Header("X-CSRF-Token", cookie)

		// Validate token on mutating methods
		method := c.Request.Method
		if method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete || method == http.MethodPatch {
			path := c.Request.URL.Path

			// Public auth endpoints don't have an existing session to exploit via CSRF
			isPublicAuth := strings.HasPrefix(path, "/api/v1/auth/login") ||
				strings.HasPrefix(path, "/api/v1/auth/register") ||
				strings.HasPrefix(path, "/api/v1/auth/forgot-password") ||
				strings.HasPrefix(path, "/api/v1/auth/reset-password") ||
				strings.HasPrefix(path, "/api/v1/auth/change-temporary-password")

			// If this is not a public auth endpoint, and the client has a session cookie, enforce CSRF
			_, errSession := c.Cookie("pux_session")
			hasSession := errSession == nil
			if !isPublicAuth && hasSession {
				headerToken := c.GetHeader("X-CSRF-Token")
				if headerToken == "" || headerToken != cookie {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF token mismatch or missing"})
					return
				}
			}
		}

		c.Next()
	}
}

