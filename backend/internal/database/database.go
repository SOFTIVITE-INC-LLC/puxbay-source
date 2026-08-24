package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/softivite/puxbay/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database instance.
var DB *gorm.DB

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const TenantSchemaKey contextKey = "tenant_schema"

// quoteIdentifier ensures schema names are safely quoted for SQL execution.
func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// Connect initializes the database connection with GORM.
func Connect(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	// Configure GORM logger based on environment
	logLevel := logger.Warn
	if os.Getenv("APP_ENV") == "development" {
		logLevel = logger.Info
	}

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger:                 gormLogger,
		PrepareStmt:            true,
		SkipDefaultTransaction: true, // Better performance for read-heavy workloads
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Register global callbacks for Audit Logging
	RegisterAuditCallbacks(db)

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Verify connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
	return db, nil
}

// SetTenantSchema sets the current PostgreSQL search path for Schema Isolation.
// This is typically not needed for requests if you use WithTenantSchema.
func SetTenantSchema(db *gorm.DB, schemaName string) *gorm.DB {
	if schemaName == "" {
		schemaName = "public"
	}
	return db.Exec(fmt.Sprintf("SET LOCAL search_path TO %s", quoteIdentifier(schemaName)))
}

// ClearTenantSchema removes the tenant schema and resets to public for the transaction.
func ClearTenantSchema(db *gorm.DB) *gorm.DB {
	return db.Exec("SET LOCAL search_path TO public")
}

// WithTenantSchema returns a new GORM session scoped to a specific tenant schema using Context.
func WithTenantSchema(db *gorm.DB, schemaName string) *gorm.DB {
	if schemaName == "" {
		schemaName = "public"
	}
	// Get the base context — db.Statement may be nil when called from middleware
	var baseCtx context.Context
	if db.Statement != nil && db.Statement.Context != nil {
		baseCtx = db.Statement.Context
	} else {
		baseCtx = context.Background()
	}
	ctx := context.WithValue(baseCtx, TenantSchemaKey, schemaName)
	return db.WithContext(ctx)
}

// CreateTenantSchema creates a new PostgreSQL schema for a tenant.
func CreateTenantSchema(db *gorm.DB, schemaName string) error {
	return db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", quoteIdentifier(schemaName))).Error
}
