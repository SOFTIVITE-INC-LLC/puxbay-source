package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/softivite/puxbay/internal/logger"
)

// AppError represents a safe error to be returned to the client
type AppError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
	Log     error  `json:"-"` // Internal error to log, not returned to client
}

func (e *AppError) Error() string {
	if e.Log != nil {
		return e.Log.Error()
	}
	return e.Message
}

// NewAppError creates a new AppError
func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Log:     err,
	}
}

// ErrorMiddleware catches gin c.Errors and returns a sanitized response
func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			if appErr, ok := err.(*AppError); ok {
				if appErr.Log != nil {
					logger.Log.Error("API Error",
						slog.String("path", c.Request.URL.Path),
						slog.Any("error", appErr.Log),
					)
				}
				c.JSON(appErr.Code, gin.H{"error": appErr.Message})
				return
			}

			// Unhandled generic error
			logger.Log.Error("Unhandled API Error",
				slog.String("path", c.Request.URL.Path),
				slog.Any("error", err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An unexpected error occurred."})
		}
	}
}
