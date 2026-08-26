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
// supply an Authorization: Bearer token WITHOUT any cookies (mobile apps, API
// clients) are exempt, since an attacker cannot forge a cross-origin request
// with a custom header.
//
// Browser sessions now use the HttpOnly pux_session cookie. This means all
// browser-originated requests will have cookies and MUST pass CSRF validation.
func CSRFMiddleware(rootDomain string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF when no Cookie header is present at all.
		// Clients that never send cookies (mobile apps, API tools) are inherently
		// immune to CSRF attacks.
		if c.GetHeader("Cookie") == "" {
			c.Next()
			return
		}

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
			headerToken := c.GetHeader("X-CSRF-Token")
			if headerToken == "" || headerToken != cookie {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF token mismatch or missing"})
				return
			}
		}

		c.Next()
	}
}

