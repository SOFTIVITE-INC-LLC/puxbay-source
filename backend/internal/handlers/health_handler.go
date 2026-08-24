package handlers

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// HealthHandler handles health check endpoints.
type HealthHandler struct {
	db          *gorm.DB
	redisClient *redis.Client
	startTime   time.Time
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(db *gorm.DB, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:          db,
		redisClient: redisClient,
		startTime:   time.Now(),
	}
}

// HealthCheck is a basic liveness probe.
// GET /health
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "puxbay-api",
		"uptime":  time.Since(h.startTime).String(),
	})
}

// ReadinessCheck verifies all dependencies are ready.
// GET /health/ready
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	checks := gin.H{}
	allHealthy := true

	// Check database
	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		checks["database"] = "unhealthy"
		allHealthy = false
	} else {
		checks["database"] = "healthy"
	}

	// Gap #38: Check Redis
	if h.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := h.redisClient.Ping(ctx).Err(); err != nil {
			checks["redis"] = "unhealthy"
			allHealthy = false
		} else {
			checks["redis"] = "healthy"
		}
	} else {
		checks["redis"] = "not_configured"
	}

	status := http.StatusOK
	statusText := "ready"
	if !allHealthy {
		status = http.StatusServiceUnavailable
		statusText = "not ready"
	}

	c.JSON(status, gin.H{
		"status": statusText,
		"checks": checks,
	})
}

// MetricsCheck returns system metrics.
// GET /health/metrics
func (h *HealthHandler) MetricsCheck(c *gin.Context) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	sqlDB, _ := h.db.DB()
	dbStats := sqlDB.Stats()

	c.JSON(http.StatusOK, gin.H{
		"uptime_seconds":  time.Since(h.startTime).Seconds(),
		"go_routines":     runtime.NumGoroutine(),
		"memory_alloc_mb": float64(memStats.Alloc) / 1024 / 1024,
		"memory_sys_mb":   float64(memStats.Sys) / 1024 / 1024,
		"gc_cycles":       memStats.NumGC,
		"db_open_conns":   dbStats.OpenConnections,
		"db_in_use":       dbStats.InUse,
		"db_idle":         dbStats.Idle,
	})
}
