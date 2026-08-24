package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// IdempotencyKey prevents duplicate processing of requests (e.g., payments).
func IdempotencyKey(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("Idempotency-Key")
		if key == "" {
			c.Next()
			return
		}

		ctx := context.Background()
		cacheKey := "idempotency:" + key

		// Try to lock the key
		success, err := redisClient.SetNX(ctx, cacheKey, "processing", 24*time.Hour).Result()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to check idempotency key"})
			return
		}

		if !success {
			// Key already exists, indicating a duplicate request
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "Duplicate request detected"})
			return
		}

		c.Next()
	}
}
