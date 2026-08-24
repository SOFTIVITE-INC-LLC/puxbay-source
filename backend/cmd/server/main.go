package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/database"
	"github.com/softivite/puxbay/internal/logger"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/router"
	"github.com/softivite/puxbay/internal/validation"
	"github.com/softivite/puxbay/internal/websocket"
	"github.com/softivite/puxbay/internal/workers"
	"gorm.io/gorm"
)

// @title Puxbay API
// @version 1.0
// @description This is the Puxbay POS API server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.puxbay.com/support
// @contact.email support@puxbay.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:5000
// @BasePath /api/v1

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logger.InitLogger(cfg.App.Env)
	defer logger.Sync()
	logger.Log.Info("🚀 Starting " + cfg.App.Name + " in " + cfg.App.Env + " mode")

	// 2. Connect to database
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("%s", "Failed to connect to database: "+err.Error())
	}
	logger.Log.Info("✅ Database connected")

	// 2.5 Init Validation Translations
	validation.InitTranslations()

	// 3. Database Migrations
	logger.Log.Info("📦 Running public schema auto-migrations...")

	// Ensure we are in the public schema
	database.ClearTenantSchema(db)

	if err := models.MigratePublicModels(db); err != nil {
		log.Fatalf("Failed to run public migrations: %v", err)
	}

	// Seed RBAC Default Roles and Permissions
	if err := database.SeedRBAC(db); err != nil {
		log.Fatalf("Failed to seed RBAC: %v", err)
	}

	log.Println("📦 Running tenant schema auto-migrations...")
	var schemas []string
	db.Table("tenants").Select("schema_name").Where("schema_name != ''").Pluck("schema_name", &schemas)

	for _, schema := range schemas {
		log.Printf("Migrating schema: %s", schema)

		// Create schema if it doesn't exist
		database.CreateTenantSchema(db, schema)

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := database.SetTenantSchema(tx, schema).Error; err != nil {
				return err
			}
			return models.MigrateTenantModels(tx)
		})
		if err != nil {
			log.Fatalf("Failed to run tenant migrations for %s: %v", schema, err)
		}
	}

	// Reset back to public
	database.ClearTenantSchema(db)

	log.Println("✅ All migrations complete")

	// 4. Setup WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	// 4.5 Initialize OpenTelemetry Tracing with DB Exporter
	tp, err := logger.InitTracer(cfg.App.Name, db, hub)
	if err != nil {
		logger.Log.Warn("Failed to initialize OpenTelemetry tracer: " + err.Error())
	} else {
		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				logger.Log.Warn("Error shutting down tracer provider: " + err.Error())
			}
		}()
		logger.Log.Info("🔍 OpenTelemetry tracing initialized (DB Exporter)")
	}

	// 5. Setup router
	r := router.Setup(cfg, db, hub)

	// 6. Setup Asynq Worker Server
	workerServer := workers.NewWorkerServer(cfg, db)
	go func() {
		if err := workerServer.Start(); err != nil {
			log.Fatalf("Failed to start Asynq worker server: %v", err)
		}
	}()
	defer workerServer.Stop()

	// 6.5. Start lightweight background workers
	workers.StartCartRecoveryWorker(db, cfg.SMTP)
	workers.StartBillingWorker(db, cfg.SMTP)
	workers.StartTelemetryCleanupWorker(db)

	// 7. Start HTTP server with graceful shutdown
	addr := fmt.Sprintf(":%s", cfg.App.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		logger.Log.Info(fmt.Sprintf("🌐 Server listening on %s", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("%s", fmt.Sprintf("Failed to start server: %v", err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("%s", fmt.Sprintf("Server forced to shutdown: %v", err))
	}

	// Close DB connections
	if sqlDB, err := db.DB(); err == nil {
		logger.Log.Info("Closing database connections...")
		sqlDB.Close()
	}

	logger.Log.Info("Server exiting")
}
