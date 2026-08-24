package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type clientRecord struct {
	count        int
	resetTime    time.Time
	lockoutUntil time.Time
}

// RateLimitMiddleware provides in-memory IP-based rate limiting with strict lockout policies.
// It allows maxRequests per window. If exceeded, the IP is locked out for lockoutDuration.
//
// counters. This has been updated to use Redis if provided, solving Gap #36 & #37.
func RateLimitMiddleware(redisClient *redis.Client, maxRequests int, window time.Duration, lockoutDuration time.Duration) gin.HandlerFunc {
	// Fallback in-memory state
	var (
		mu      sync.Mutex
		clients = make(map[string]*clientRecord)
	)

	// Background cleanup for in-memory fallback
	if redisClient == nil {
		ticker := time.NewTicker(window)
		go func() {
			for range ticker.C {
				mu.Lock()
				now := time.Now()
				for ip, record := range clients {
					if now.After(record.resetTime) && now.After(record.lockoutUntil) {
						delete(clients, ip)
					}
				}
				mu.Unlock()
			}
		}()
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		if redisClient != nil {
			// Redis-backed rate limiter
			ctx := context.Background()
			lockKey := fmt.Sprintf("ratelimit:lockout:%s", ip)
			countKey := fmt.Sprintf("ratelimit:count:%s", ip)

			locked, _ := redisClient.Exists(ctx, lockKey).Result()
			if locked > 0 {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"error": "Account temporarily locked due to too many requests. Please try again later.",
				})
				return
			}

			count, err := redisClient.Incr(ctx, countKey).Result()
			if err != nil {
				// Fallback to allow if Redis is down
				c.Next()
				return
			}

			if count == 1 {
				redisClient.Expire(ctx, countKey, window)
			}

			if count > int64(maxRequests) {
				redisClient.Set(ctx, lockKey, 1, lockoutDuration)
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"error": "Too many requests. You have been temporarily locked out.",
				})
				return
			}

			c.Next()
			return
		}

		// Fallback in-memory rate limiter
		mu.Lock()
		record, exists := clients[ip]
		if !exists || (now.After(record.resetTime) && now.After(record.lockoutUntil)) {
			clients[ip] = &clientRecord{
				count:     1,
				resetTime: now.Add(window),
			}
			mu.Unlock()
			c.Next()
			return
		}

		if now.Before(record.lockoutUntil) {
			mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Account temporarily locked due to too many requests. Please try again later.",
			})
			return
		}

		if record.count >= maxRequests {
			record.lockoutUntil = now.Add(lockoutDuration)
			mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. You have been temporarily locked out.",
			})
			return
		}

		record.count++
		mu.Unlock()
		c.Next()
	}
}
