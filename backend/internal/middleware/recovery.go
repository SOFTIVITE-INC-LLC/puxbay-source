package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/softivite/puxbay/internal/logger"
)

// RecoveryMiddleware catches panics, logs the stack trace to Zap, and returns a 500.
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				if logger.Log != nil {
					logger.Log.Error("Panic recovered",
						slog.Any("error", err),
						slog.String("stack", string(debug.Stack())),
					)
				}

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
			}
		}()
		c.Next()
	}
}
