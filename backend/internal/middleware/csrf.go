package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CSRFMiddleware implements a Double Submit Cookie CSRF protection.
// It ensures that mutating requests provide a valid X-CSRF-Token header.
//
// NOTE: CSRF attacks are only possible with cookie-based auth. Requests that
// supply an Authorization: Bearer token (mobile apps, API clients) are exempt,
// since an attacker cannot forge a cross-origin request with a custom header.
func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF for Bearer-token authenticated requests (mobile / API clients)
		if strings.HasPrefix(c.GetHeader("Authorization"), "Bearer ") {
			c.Next()
			return
		}

		// Skip CSRF when no Cookie header is present.
		// CSRF requires the browser to silently attach cookies to cross-origin requests;
		// clients that never send cookies (mobile apps, API tools) are inherently immune.
		if c.GetHeader("Cookie") == "" {
			c.Next()
			return
		}

		cookie, err := c.Cookie("csrf_token")
		if err != nil || cookie == "" {
			b := make([]byte, 32)
			_, _ = rand.Read(b)
			cookie = hex.EncodeToString(b)

			// Use secure cookie settings in production
			isProduction := os.Getenv("APP_ENV") == "production" || os.Getenv("APP_ENV") == "staging"
			c.SetCookie("csrf_token", cookie, 3600, "/", "", isProduction, isProduction)
		}

		// Always expose the token so the frontend can read it
		c.Header("X-CSRF-Token", cookie)

		// Validate token on mutating methods
		method := c.Request.Method
		if method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete || method == http.MethodPatch {
			headerToken := c.GetHeader("X-CSRF-Token")
			if headerToken == "" || headerToken != cookie {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF token mismatch or missing"})
				return
			}
		}

		c.Next()
	}
}
