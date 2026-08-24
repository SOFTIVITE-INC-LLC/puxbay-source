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

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/softivite/puxbay/internal/admin"
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
	if cfg.App.Env == "development" || cfg.App.Env == "production" {
		logger.Log.Info("📦 Running public schema auto-migrations...")

		// Ensure we are in the public schema
		database.ClearTenantSchema(db)

		// Public Models
		err = db.AutoMigrate(
			&models.Tenant{},
			&models.Domain{},
			&models.User{},
			&models.UserProfile{},
			&models.Plan{},
			&models.PricingPlan{},
			&models.PlanFeature{},
			&models.AdminRole{},
			&models.AdminUser{},
			&models.IPAllowlist{},
			&models.MasterAPIKey{},

			// RBAC Dynamic Permissions
			&models.Permission{},
			&models.Role{},

			&models.TelemetryLog{},

			&models.AuditLog{},
			&models.Subscription{},
			&models.BillingPayment{},
			&models.PromoCode{},
			&models.ReferralReward{},
			&models.BillingSettings{},
			&models.CrossTenantAuditLog{},
			&models.LegalDocument{},
			&models.BlogPost{},
			&models.DatabaseBackup{},
			&models.TenantMetrics{},
			&models.FeatureFlag{},
			&models.SEOSettings{},
			&models.Broadcast{},
		)
		if err != nil {
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

				return tx.AutoMigrate(
					&models.Branch{},
					&models.CashDrawerSession{},
					&models.Shift{},
					&models.StaffShift{},
					&models.PrintJob{},
					&models.Category{},
					&models.Product{},
					&models.ProductVariant{},
					&models.ProductComponent{},
					&models.ProductHistory{},
					&models.StockTransfer{},
					&models.StockTransferItem{},
					&models.PurchaseOrder{},
					&models.PurchaseOrderItem{},
					&models.StocktakeSession{},
					&models.StockMovement{},
					&models.Order{},
					&models.OrderItem{},
					&models.CustomerTier{},
					&models.Customer{},
					&models.LoyaltyTransaction{},
					&models.GiftCard{},
					&models.Supplier{},
					&models.AuditLog{},
					&models.SupplierProfile{},
					&models.SupplierProduct{},
					&models.SupplierLedgerEntry{},
					&models.ExpenseCategory{},
					&models.Expense{},
					&models.CustomerSegment{},
					&models.AbandonedCart{},

					&models.Domain{},
					&models.StockBatch{},
					&models.PaymentMethod{},
					&models.Payment{},
					&models.TaxConfiguration{},
					&models.Return{},
					&models.ReturnItem{},
					&models.CRMSettings{},
					&models.MarketingCampaign{},
					&models.Promotion{},
					&models.DiscountCode{},
					&models.CustomerFeedback{},
					&models.DiningTable{},
					&models.KDSTicket{},
					&models.SplitBillGroup{},
					&models.PayrollPeriod{},
					&models.StorefrontSettings{},
					&models.ProductReview{},
					&models.Wishlist{},
					&models.Coupon{},
					&models.NewsletterSubscription{},
					&models.AbandonedCart{},
					&models.ProductImageGallery{},
					&models.TenantIntegration{},
					&models.PayrollRecord{},
					&models.LeaveRequest{},
					&models.Attendance{},
					&models.ServiceCategory{},
					&models.Service{},
					&models.Appointment{},
					&models.ServiceCommissionRule{},
					&models.ServiceCommissionRecord{},
					&models.Quotation{},
					&models.QuotationItem{},
					&models.AuditLog{},
					&models.APIRequestLog{},
					&models.HoneypotAttempt{},
					&models.APIKey{},
					&models.ExternalSystem{},
					&models.WebhookEndpoint{},
					&models.WebhookEvent{},
					&models.Notification{},
					&models.StocktakeEntry{},
					&models.NotificationSetting{},
					&models.DeliveryDriver{},
					&models.DeliveryOrder{},
					&models.LedgerAccount{},
					&models.JournalEntry{},
					&models.LedgerLine{},
					&models.StockAlert{},
					&models.SupportTicket{},
					&models.TicketMessage{},
					&models.StockBatch{},
					&models.StockMovement{},
				)
			})
			if err != nil {
				log.Fatalf("Failed to run tenant migrations for %s: %v", schema, err)
			}
		}

		// Reset back to public
		database.ClearTenantSchema(db)

		log.Println("✅ All migrations complete")
	} else {
		// Run golang-migrate in staging/production
		logger.Log.Info("📦 Running database migrations via golang-migrate...")

		dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.DBName,
			cfg.Database.SSLMode,
		)

		m, err := migrate.New(
			"file://internal/db/migrations",
			dbURL,
		)
		if err != nil {
			log.Fatalf("Failed to initialize migrations: %v", err)
		}

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		logger.Log.Info("✅ Database migrations applied successfully")
	}

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

	// 5. Setup GoAdmin
	if err := admin.Setup(r, cfg); err != nil {
		log.Fatalf("Failed to setup GoAdmin: %v", err)
	}

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
