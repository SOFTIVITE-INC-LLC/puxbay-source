package logger

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// GinZapMiddleware logs gin HTTP requests using zap.
func GinZapMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		if Log == nil {
			return
		}

		end := time.Now()
		latency := end.Sub(start)

		if len(c.Errors) > 0 {
			// Append error field if this is an erroneous request
			for _, e := range c.Errors.Errors() {
				Log.Error(e)
			}
		} else {
			reqID, _ := c.Get("request_id")
			Log.Info(path,
				slog.Int("status", c.Writer.Status()),
				slog.String("method", c.Request.Method),
				slog.String("path", path),
				slog.String("query", query),
				slog.String("ip", c.ClientIP()),
				slog.String("user-agent", c.Request.UserAgent()),
				slog.Any("request_id", reqID),
				slog.Any("duration", latency),
			)
		}
	}
}
