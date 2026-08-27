package router

import (
	"log"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/handlers"
	"github.com/softivite/puxbay/internal/logger"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/services"
	"github.com/softivite/puxbay/internal/websocket"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
)

// Setup configures all routes for the application.
func Setup(cfg *config.Config, db *gorm.DB, hub *websocket.Hub) *gin.Engine {
	// Set Gin mode
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Gap #26: Set max request body size to 32MB to prevent OOM from large uploads
	r.MaxMultipartMemory = 32 << 20 // 32 MB

	// Initialize Redis client for idempotency and health checks (Gap #19, #38)
	var redisClient *redis.Client
	opt, redisErr := redis.ParseURL(cfg.Redis.URL)
	if redisErr == nil {
		redisClient = redis.NewClient(opt)
	}

	// Initialize handlers
	offlineHandler := handlers.NewOfflineHandler(db)
	kioskHandler := handlers.NewKioskHandler(db)

	// Global middleware
	r.Use(otelgin.Middleware("puxbay")) // OpenTelemetry distributed tracing
	r.Use(middleware.PrometheusMetrics()) // Prometheus metrics collection
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(middleware.RequestID())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.ErrorMiddleware())
	r.Use(logger.GinZapMiddleware())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORSMiddleware(&cfg.CORS))
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CSRFMiddleware(cfg.App.RootDomain))
	r.Use(middleware.RateLimitMiddleware(redisClient, 100, time.Minute, 5*time.Minute)) // 100 requests per minute

	// Initialize services
	var tokenStore services.TokenStore
	tokenStore, err := services.NewRedisTokenStore(cfg.Redis.URL)
	if err != nil {
		log.Printf("Warning: failed to connect to Redis for token store (%v). Using NoopTokenStore.", err)
		tokenStore = &services.NoopTokenStore{}
	}

	authService := services.NewAuthService(&cfg.JWT, db, tokenStore, cfg.App.RootDomain)

	// Wire email service into auth service so Register() can send verification emails
	authEmailService := services.NewEmailService(db, cfg.SMTP)
	authService.SetEmailService(authEmailService)

	smsService := services.NewSMSService(cfg.SMS)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, authService, cfg.SMTP, cfg.App.RootDomain, cfg.JWT)
	healthHandler := handlers.NewHealthHandler(db, redisClient)
	branchHandler := handlers.NewBranchHandler(db)
	productHandler := handlers.NewProductHandler(db)
	categoryHandler := handlers.NewCategoryHandler(db)
	customerHandler := handlers.NewCustomerHandler(db)
	profileHandler := handlers.NewUserProfileHandler(db)
	orderHandler := handlers.NewOrderHandler(db, hub, smsService, cfg.App.RootDomain)
	inventoryHandler := handlers.NewInventoryHandler(db)
	financialHandler := handlers.NewFinancialHandler(db)
	billingHandler := handlers.NewBillingHandler(db, &cfg.Paystack)
	emailService := services.NewEmailService(db, cfg.SMTP)
	staffHandler := handlers.NewStaffHandler(db, authService, smsService, emailService, cfg.App.RootDomain)
	b2bHandler := handlers.NewB2BHandler(db)
	supplierHandler := handlers.NewSupplierHandler(db)
	supplierPortalHandler := handlers.NewSupplierPortalHandler(db, authService)
	hrHandler := handlers.NewHRHandler(db)
	serviceHandler := handlers.NewServiceHandler(db)
	marketingHandler := handlers.NewMarketingHandler(db)
	fnbHandler := handlers.NewFnBHandler(db, hub)
	analyticsHandler := handlers.NewAnalyticsHandler(db)
	intelligenceHandler := handlers.NewIntelligenceHandler(db)
	exportHandler := handlers.NewExportHandler(db)
	barcodeHandler := handlers.NewBarcodeHandler(db)
	securityHandler := handlers.NewSecurityHandler(db)
	privacyHandler := handlers.NewPrivacyHandler(db)
	publicPortalHandler := handlers.NewPublicPortalHandler(db)
	adminHandler := handlers.NewAdminHandler(db, authService)
	storefrontHandler := handlers.NewStorefrontAPIHandler(db, authService, &cfg.Paystack, redisClient)
	webhookHandler := handlers.NewWebhookHandler(db)
	walletHandler := handlers.NewWalletAPIHandler(db)
	notificationHandler := handlers.NewNotificationHandler(db, hub)
	pushService := services.NewPushService(db, hub)
	deviceTokenHandler := handlers.NewDeviceTokenHandler(pushService)
	returnHandler := handlers.NewReturnHandler(db)
	cashDrawerHandler := handlers.NewCashDrawerHandler(db)
	giftCardHandler := handlers.NewGiftCardHandler(db)
	paymentMethodHandler := handlers.NewPaymentMethodHandler(db, &cfg.Paystack)
	crmHandler := handlers.NewCRMHandler(db)
	settingsHandler := handlers.NewSettingsHandler(db)
	contentHandler := handlers.NewContentHandler(db)
	copilotHandler := handlers.NewCopilotHandler(cfg, db)
	deliveryHandler := handlers.NewDeliveryHandler(db)
	omnichannelHandler := handlers.NewOmnichannelHandler(db)
	syncHandler := handlers.NewSyncHandler(db)
	integrationHandler := handlers.NewIntegrationHandler(cfg, db)
	roleHandler := handlers.NewRoleHandler(db)
	accountingHandler := handlers.NewAccountingHandler(db)
	scheduleHandler := handlers.NewScheduleHandler(db)
	deviceHandler := handlers.NewDeviceHandler(db, cfg)
	creditService := services.NewCreditService(db, smsService)
	creditHandler := handlers.NewCreditHandler(db, creditService)
	reportService := services.NewReportService(db, cfg.SMTP)
	reportScheduleHandler := handlers.NewReportScheduleHandler(db, reportService, emailService)
	// ──────────────────────────────────────────────
	// Health Check Endpoints (no auth required)
	// ──────────────────────────────────────────────
	r.GET("/healthz", healthHandler.HealthCheck)
	r.GET("/readyz", healthHandler.ReadinessCheck)
	r.GET("/health/metrics", healthHandler.MetricsCheck)
	r.GET("/metrics", gin.WrapH(promhttp.Handler())) // Prometheus scrape endpoint

	// ──────────────────────────────────────────────
	// API v1
	// ──────────────────────────────────────────────
	v1 := r.Group("/api/v1")
	{
		// ──────────────────────────────────────────────────────────
		// Public Kiosk Routes (tenant resolved via subdomain, no JWT)
		// ──────────────────────────────────────────────────────────
		publicKiosk := v1.Group("/kiosk")
		publicKiosk.Use(middleware.TenantMiddleware())
		publicKiosk.Use(middleware.BranchMiddleware())
		{
			publicKiosk.POST("/orders", kioskHandler.PlaceOrder)
			publicKiosk.POST("/customers", kioskHandler.RegisterCustomer)
		}

		// ──────────────────────────────────────────────────────────
		// Public Stocktake Portal Routes (no JWT)
		// ──────────────────────────────────────────────────────────
		publicStocktake := v1.Group("/public/stocktake")
		publicStocktake.Use(middleware.TenantMiddleware())
		{
			publicStocktake.GET("/:token", inventoryHandler.PublicStocktakeSession)
			publicStocktake.GET("/:token/scan", inventoryHandler.PublicStocktakeScan)
			publicStocktake.POST("/:token/update", inventoryHandler.PublicStocktakeUpdate)
		}

		// ──────────────────────────────────────────────────────────
		// Supplier Portal (public login + JWT-protected portal routes)
		// ──────────────────────────────────────────────────────────
		supplierPortal := v1.Group("/supplier-portal")
		supplierPortal.Use(middleware.TenantMiddleware())
		{
			supplierPortal.POST("/login", supplierPortalHandler.Login)

			portalProtected := supplierPortal.Group("")
			portalProtected.Use(handlers.SupplierAuthMiddleware(authService))
			{
				portalProtected.GET("/me", supplierPortalHandler.Me)
				portalProtected.GET("/purchase-orders", supplierPortalHandler.ListPurchaseOrders)
			}
		}

		// Auth endpoints (no auth required)
		auth := v1.Group("/auth")
		auth.Use(middleware.RateLimitMiddleware(redisClient, 10, time.Minute, 15*time.Minute)) // 10 requests per min, 15m lockout
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/register", authHandler.Register)
			auth.POST("/change-temporary-password", authHandler.ChangeTemporaryPassword)

			// Email verification (no rate-limit — users actively waiting)
			auth.POST("/verify-email", authHandler.VerifyEmail)
			auth.GET("/verify-email", authHandler.VerifyEmail) // magic-link click
			auth.POST("/verify-email-otp", authHandler.VerifyEmailOTP)
			auth.POST("/resend-verification", authHandler.ResendVerificationEmail)

			// Strict rate limit for password resets (3 requests per hour, 1 hour lockout)
			pwdGroup := auth.Group("")
			pwdGroup.Use(middleware.RateLimitMiddleware(redisClient, 3, time.Hour, time.Hour))
			{
				pwdGroup.POST("/forgot-password", authHandler.ForgotPassword)
				pwdGroup.POST("/reset-password", authHandler.ResetPassword)
			}
		}

		// Protected auth endpoints
		authProtected := v1.Group("/auth")
		authProtected.Use(middleware.AuthMiddleware(authService))
		{
			authProtected.POST("/logout", authHandler.Logout)
			authProtected.GET("/user", authHandler.CurrentUser)
			authProtected.GET("/session", authHandler.GetSession) // SSO session restore
			authProtected.PUT("/me", authHandler.UpdateMe)
		}

		// ──────────────────────────────────────────
		// Protected API routes (require JWT + Tenant + Branch)
		// ──────────────────────────────────────────
		api := v1.Group("")
		api.Use(middleware.AuthMiddleware(authService))
		api.Use(middleware.TenantMiddleware())
		api.Use(middleware.BillingMiddleware()) // Enforces the 7-day trial and subscription status
		api.Use(middleware.BranchMiddleware())
		api.Use(middleware.FieldFiltering())
		{
			// Profiles
			api.GET("/profiles", profileHandler.List)
			api.GET("/profiles/:id", profileHandler.Get)
			api.PUT("/profiles/:id", profileHandler.Update)
			api.PUT("/profiles/pos-pin", profileHandler.SetPOSPin)

			// Products
			api.GET("/products", productHandler.List)
			api.POST("/products", middleware.RequirePermission("products:create"), productHandler.Create)
			api.POST("/products/import", middleware.RequirePermission("products:create"), productHandler.ImportExcel)
			api.GET("/products/:id", productHandler.Get)
			api.PUT("/products/:id", middleware.RequirePermission("products:update"), productHandler.Update)
			api.DELETE("/products/:id", middleware.RequirePermission("products:delete"), productHandler.Delete)
			api.GET("/products/:id/barcode", barcodeHandler.GenerateProductBarcode) // alias for frontend

			// Categories
			api.GET("/categories", categoryHandler.List)
			api.POST("/categories", middleware.RequirePermission("products:create"), categoryHandler.Create)
			api.GET("/categories/:id", categoryHandler.Get)
			api.PUT("/categories/:id", middleware.RequirePermission("products:update"), categoryHandler.Update)
			api.DELETE("/categories/:id", middleware.RequirePermission("products:delete"), categoryHandler.Delete)

			// Orders
			api.GET("/orders", orderHandler.List)
			api.GET("/orders/summary", orderHandler.Summary)
			api.POST("/orders/pos", middleware.BranchRequiredMiddleware(), orderHandler.POS)
			if redisClient != nil {
				api.POST("/orders", middleware.IdempotencyKey(redisClient), orderHandler.Create)
			} else {
				api.POST("/orders", orderHandler.Create)
			}
			api.GET("/orders/:id", orderHandler.Get)
			api.POST("/orders/:id/void", middleware.RequirePermissionOrOverride(db, "orders:void"), orderHandler.VoidOrder)
			api.POST("/orders/:id/complete", middleware.RequirePermission("orders:update"), orderHandler.CompleteOrder)
			api.GET("/orders/:id/receipt", orderHandler.GetReceipt)

			// Barcodes (protected — Gap #6)
			api.GET("/barcodes/product/:id", barcodeHandler.GenerateProductBarcode)
			api.GET("/qrs/product/:id", barcodeHandler.GenerateProductQR)
			api.POST("/barcodes/bulk-generate", barcodeHandler.BulkGenerateBarcodes)

			// Inventory
			api.GET("/inventory/transfers", middleware.RequirePermission("inventory:read"), inventoryHandler.ListTransfers)
			api.POST("/inventory/transfers", middleware.RequirePermission("inventory:manage"), inventoryHandler.CreateTransfer)
			api.GET("/inventory/transfers/:id", middleware.RequirePermission("inventory:read"), inventoryHandler.GetTransfer)
			api.POST("/inventory/transfers/:id/approve", middleware.RequirePermission("inventory:manage"), inventoryHandler.ApproveTransfer)
			api.POST("/inventory/transfers/:id/ship", middleware.RequirePermission("inventory:manage"), inventoryHandler.ShipTransfer)
			api.POST("/inventory/transfers/:id/receive", middleware.RequirePermission("inventory:manage"), inventoryHandler.ReceiveTransfer)
			api.GET("/inventory/purchase-orders", middleware.RequirePermission("inventory:read"), inventoryHandler.ListPOs)
			api.POST("/inventory/purchase-orders", middleware.BranchRequiredMiddleware(), middleware.RequirePermission("inventory:manage"), inventoryHandler.CreatePO)
			api.GET("/inventory/purchase-orders/:id", middleware.RequirePermission("inventory:read"), inventoryHandler.GetPO)
			api.POST("/inventory/purchase-orders/:id/receive", middleware.RequirePermission("inventory:manage"), inventoryHandler.ReceivePO)
			api.GET("/inventory/stocktakes", middleware.RequirePermission("inventory:read"), inventoryHandler.ListStocktakes)
			api.GET("/inventory/stocktakes/:id", middleware.RequirePermission("inventory:read"), inventoryHandler.GetStocktake)
			api.POST("/inventory/stocktakes", middleware.BranchRequiredMiddleware(), middleware.RequirePermission("inventory:manage"), inventoryHandler.CreateStocktake)
			api.POST("/inventory/stocktakes/:id/finalize", middleware.RequirePermission("inventory:manage"), inventoryHandler.FinalizeStocktake)
			api.GET("/inventory/movements", middleware.RequirePermission("inventory:read"), inventoryHandler.ListMovements)
			api.POST("/inventory/receive", inventoryHandler.ReceiveStock)
			api.GET("/inventory/low-stock", inventoryHandler.LowStockAlerts)
			api.GET("/inventory/products/:id/history", inventoryHandler.GetProductHistory)
			api.GET("/inventory/products/:id/components", inventoryHandler.GetProductComponents)
			// Batch & Expiry
			api.GET("/inventory/products/:id/batches", inventoryHandler.ListBatches)
			api.POST("/inventory/products/:id/batches", inventoryHandler.CreateBatch)
			api.PUT("/inventory/batches/:batchId", inventoryHandler.UpdateBatch)
			api.DELETE("/inventory/batches/:batchId", inventoryHandler.DeleteBatch)
			api.GET("/inventory/expiring-batches", inventoryHandler.ListExpiringBatches)

			// Offline Sync
			api.POST("/offline/sync", offlineHandler.SyncBatchedTransactions)

			// Kiosk
			api.GET("/kiosk/:branch_id/config", kioskHandler.GetConfig)
			api.GET("/kiosk/:branch_id/menu", kioskHandler.GetMenu)

			// Billing
			billingGroup := api.Group("/billing")
			billingGroup.Use(middleware.RequirePermission("billing:manage"))
			{
				billingGroup.GET("/subscription", billingHandler.GetSubscription)
				billingGroup.GET("/invoices", billingHandler.ListInvoices)
				billingGroup.POST("/validate-promo", billingHandler.ValidatePromo)
				billingGroup.GET("/plans", billingHandler.ListPlans)
				// Verify payment after Paystack redirect (dev fallback since webhooks can't reach localhost)
				billingGroup.GET("/verify/:reference", billingHandler.VerifyPayment)
			}
			// Gap #19: Wire idempotency middleware on payment endpoints
			if redisClient != nil {
				api.POST("/billing/checkout/:plan_id", middleware.IdempotencyKey(redisClient), billingHandler.ProcessPayment)
			} else {
				api.POST("/billing/checkout/:plan_id", billingHandler.ProcessPayment)
			}

			// Financial (admin/manager only)
			financial := api.Group("/financial")
			financial.Use(middleware.RequirePermission("finance:manage"))
			{
				financial.GET("/expenses", financialHandler.ListExpenses)
				financial.POST("/expenses", financialHandler.CreateExpense)
				financial.PUT("/expenses/:id", financialHandler.UpdateExpense)
				financial.DELETE("/expenses/:id", financialHandler.DeleteExpense)
				financial.GET("/expense-categories", financialHandler.ListExpenseCategories)
				financial.POST("/expense-categories", financialHandler.CreateExpenseCategory)
				financial.GET("/profit-and-loss", financialHandler.GetProfitAndLoss)
				financial.GET("/taxes/config", financialHandler.GetTaxConfig)
				financial.PUT("/taxes/config", financialHandler.UpdateTaxConfig)
				financial.GET("/taxes/report", financialHandler.GetTaxReport)
				financial.GET("/returns", financialHandler.ListReturns)
				financial.POST("/returns/:id/refund", financialHandler.ProcessReturnRefund)

				financial.GET("/ledger", accountingHandler.ListLedgerAccounts)
				financial.POST("/ledger", accountingHandler.CreateLedgerAccount)
				financial.GET("/journal-entries", accountingHandler.ListJournalEntries)
				financial.POST("/journal-entries", accountingHandler.CreateJournalEntry)
			}

			// Customers (CRM)
			api.GET("/customers", customerHandler.List)
			api.POST("/customers", customerHandler.Create)
			api.GET("/customers/:id", customerHandler.Get)
			api.PUT("/customers/:id", customerHandler.Update)
			api.DELETE("/customers/:id", customerHandler.Delete)
			api.POST("/customers/:id/payment", customerHandler.RecordPayment)
			// Customer Store Credit & BNPL
			api.GET("/customers/:id/credit-account", creditHandler.GetCreditAccount)
			api.POST("/customers/:id/credit-limit", creditHandler.SetCreditLimit)
			api.POST("/customers/:id/credit/drawdown", creditHandler.DrawdownCredit)
			api.POST("/customers/:id/credit/repay", creditHandler.RecordRepayment)
			api.POST("/customers/:id/credit/send-reminder", creditHandler.SendRepaymentReminder)
			api.GET("/credit/overdue", creditHandler.GetOverdueAccounts)

			// Customer Tiers & Loyalty
			api.GET("/customer-tiers", crmHandler.ListTiers)
			api.POST("/customer-tiers", crmHandler.CreateTier)
			api.PUT("/customer-tiers/:id", crmHandler.UpdateTier)
			api.DELETE("/customer-tiers/:id", crmHandler.DeleteTier)
			api.GET("/loyalty-transactions", crmHandler.ListLoyaltyTransactions)
			api.GET("/crm/loyalty", crmHandler.ListLoyaltyTransactions) // alias
			api.GET("/store-credit-transactions", crmHandler.ListStoreCreditTransactions)
			api.GET("/feedback", crmHandler.ListFeedback)
			api.POST("/feedback", crmHandler.CreateFeedback)
			api.DELETE("/feedback/:id", crmHandler.DeleteFeedback)
			// CRM-prefixed aliases for frontend compatibility
			api.GET("/crm/feedback", crmHandler.ListFeedback)
			api.POST("/crm/feedback", crmHandler.CreateFeedback)
			api.DELETE("/crm/feedback/:id", crmHandler.DeleteFeedback)
			api.GET("/crm/customers", customerHandler.List)
			api.POST("/crm/customers", customerHandler.Create)
			api.GET("/crm/customers/:id", customerHandler.Get)
			api.PUT("/crm/customers/:id", customerHandler.Update)

			// Support Tickets / Helpdesk
			api.GET("/crm/tickets", crmHandler.ListTickets)
			api.POST("/crm/tickets", crmHandler.CreateTicket)
			api.GET("/crm/tickets/:id/messages", crmHandler.GetTicketMessages)
			api.POST("/crm/tickets/:id/reply", crmHandler.ReplyTicket)

			// Branches
			api.GET("/branches", branchHandler.List)
			api.POST("/branches", branchHandler.Create)
			api.GET("/branches/network/metrics", branchHandler.NetworkMetrics)
			api.GET("/branches/:id/metrics", branchHandler.BranchMetrics)
			api.GET("/branches/:id", branchHandler.Get)
			api.PUT("/branches/:id", branchHandler.Update)
			api.DELETE("/branches/:id", branchHandler.Delete)

			// Devices (Hardware Terminals)
			api.GET("/devices", deviceHandler.List)
			api.POST("/devices", deviceHandler.Create)
			api.GET("/devices/:id", deviceHandler.Get)
			api.PUT("/devices/:id", deviceHandler.Update)
			api.DELETE("/devices/:id", deviceHandler.Delete)
			api.POST("/devices/:id/heartbeat", deviceHandler.Heartbeat)
			api.POST("/devices/:id/print", deviceHandler.PrintDocument)

			// Suppliers
			api.GET("/suppliers", supplierHandler.List)
			api.POST("/suppliers", supplierHandler.Create)
			api.GET("/suppliers/:id", supplierHandler.Get)
			api.PUT("/suppliers/:id", supplierHandler.Update)
			api.DELETE("/suppliers/:id", supplierHandler.Delete)

			// Supplier Catalogs
			api.GET("/suppliers/:id/products", supplierHandler.ListProducts)
			api.POST("/suppliers/:id/products", supplierHandler.AddProduct)
			api.DELETE("/suppliers/:id/products/:productId", supplierPortalHandler.RemoveProduct)

			// Supplier Portal admin actions
			api.POST("/suppliers/:id/invite", supplierPortalHandler.InviteUser)

			// Supplier Ledger (Accounts Payable) — full version with date
			api.GET("/suppliers/:id/ledger", supplierHandler.ListLedger)
			api.POST("/suppliers/:id/ledger", supplierPortalHandler.AddLedgerFull)

			// Webhooks
			api.GET("/webhooks", webhookHandler.List)
			api.POST("/webhooks", webhookHandler.Create)
			api.DELETE("/webhooks/:id", webhookHandler.Delete)

			// Cash Drawers
			api.GET("/cash-drawers", cashDrawerHandler.List)
			api.POST("/cash-drawers/open", cashDrawerHandler.Open)
			api.POST("/cash-drawers/close", cashDrawerHandler.Close)
			api.GET("/cash-drawers/reports/:id", cashDrawerHandler.GetReport)

			// Notifications
			api.GET("/notifications", notificationHandler.List)
			api.GET("/notifications/latest", notificationHandler.GetLatest)
			api.POST("/notifications/:id/read", notificationHandler.MarkAsRead)
			api.POST("/notifications/read-all", notificationHandler.MarkAllAsRead)
			api.GET("/notifications/settings", notificationHandler.GetSettings)
			api.PUT("/notifications/settings", notificationHandler.UpdateSettings)
			api.DELETE("/notifications/:id", notificationHandler.Delete)

			// Device Push Tokens
			api.POST("/devices/register", deviceTokenHandler.RegisterDeviceToken)
			api.DELETE("/devices/unregister", deviceTokenHandler.UnregisterDeviceToken)

			// Gift Cards
			api.GET("/gift-cards", giftCardHandler.List)
			api.POST("/gift-cards", giftCardHandler.Create)
			api.GET("/gift-cards/check", giftCardHandler.CheckBalance)
			api.GET("/gift-cards/:id", giftCardHandler.Get)
			api.POST("/gift-cards/:id/disable", giftCardHandler.Disable)
			api.POST("/gift-cards/redeem", giftCardHandler.Redeem)

			// Returns
			api.GET("/returns", returnHandler.List)
			api.POST("/returns", returnHandler.Create)
			api.GET("/returns/:id", returnHandler.Get)
			api.POST("/returns/:id/approve", returnHandler.Approve)
			api.POST("/returns/:id/reject", returnHandler.Reject)
			api.POST("/returns/:id/refund", returnHandler.ProcessRefund)

			// Payment Methods
			api.GET("/payment-methods", paymentMethodHandler.List)
			api.POST("/payment-methods", paymentMethodHandler.Create)
			api.PUT("/payment-methods/:id", paymentMethodHandler.Update)
			api.DELETE("/payment-methods/:id", paymentMethodHandler.Delete)
			api.GET("/payment-methods/paystack/subaccounts", paymentMethodHandler.ListPaystackSubaccounts)
			api.GET("/payment-methods/paystack/subaccounts/verify/:code", paymentMethodHandler.VerifyPaystackSubaccount)
			api.GET("/payment-methods/paystack/countries", paymentMethodHandler.ListPaystackCountries)
			api.GET("/payment-methods/paystack/banks", paymentMethodHandler.ListPaystackBanks)
			api.GET("/payment-methods/paystack/resolve-account", paymentMethodHandler.ResolvePaystackAccount)
			api.POST("/payment-methods/paystack/create-subaccount", paymentMethodHandler.CreatePaystackSubaccount)
			api.GET("/payment-methods/paystack/verify/:reference", paymentMethodHandler.VerifyTransaction)
			api.GET("/pos/verify-payment", paymentMethodHandler.VerifyTransaction)

			// Staff (read: all authenticated, write: admin only)
			api.GET("/staff", staffHandler.List)
			api.GET("/staff/:id", staffHandler.Get)
			staffAdmin := api.Group("/staff")
			staffAdmin.Use(middleware.RequirePermission("staff:manage"))
			{
				staffAdmin.POST("", staffHandler.Create)
				staffAdmin.PUT("/:id", staffHandler.Update)
				staffAdmin.DELETE("/:id", staffHandler.Delete)
			}

			// F&B
			api.GET("/fnb/tables", fnbHandler.ListTables)
			api.POST("/fnb/tables", fnbHandler.CreateTable)
			api.PUT("/fnb/tables/:id/status", fnbHandler.UpdateTableStatus)
			api.GET("/fnb/kds", fnbHandler.ListKDS)
			api.PUT("/fnb/kds/:id/advance", fnbHandler.AdvanceTicketStatus)
			api.POST("/fnb/split-bills", fnbHandler.CreateSplitBill)
			api.GET("/fnb/split-bills/:id", fnbHandler.GetSplitBill)

			// HR (All Staff)
			hrGeneral := api.Group("/hr")
			{
				hrGeneral.GET("/attendance", hrHandler.ListAttendance)
				hrGeneral.POST("/attendance/clock_in", hrHandler.ClockIn)
				hrGeneral.POST("/attendance/clock_out", hrHandler.ClockOut)
				hrGeneral.GET("/leave-requests", hrHandler.ListLeaveRequests)
				hrGeneral.POST("/leave-requests", hrHandler.CreateLeaveRequest)
			}

			// HR (admin/manager only)
			hrRoutes := api.Group("/hr")
			hrRoutes.Use(middleware.RequirePermission("hr:manage"))
			{
				hrRoutes.PATCH("/attendance/:id/correct", hrHandler.CorrectAttendance)
				hrRoutes.DELETE("/attendance/:id", hrHandler.DeleteAttendance)
				hrRoutes.PUT("/leave-requests/:id/approve", hrHandler.ApproveLeaveRequest)
				hrRoutes.PUT("/leave-requests/:id/reject", hrHandler.RejectLeaveRequest)
				hrRoutes.GET("/payroll/periods", hrHandler.ListPayrollPeriods)
				hrRoutes.GET("/payroll/periods/:id", hrHandler.GetPayrollPeriod)
				hrRoutes.POST("/payroll/periods/:id/process", hrHandler.ProcessPayroll)
				hrRoutes.GET("/payslips/:id", hrHandler.GetPayslip)

				hrRoutes.GET("/commission-rules", hrHandler.ListCommissionRules)
				hrRoutes.POST("/commission-rules", hrHandler.CreateCommissionRule)
				hrRoutes.GET("/achievements", hrHandler.ListStaffAchievements)
				hrRoutes.POST("/achievements", hrHandler.CreateStaffAchievement)
				hrRoutes.GET("/shift-swaps", hrHandler.ListShiftSwapRequests)
				hrRoutes.POST("/shift-swaps", hrHandler.CreateShiftSwapRequest)

				hrRoutes.GET("/roster", scheduleHandler.ListShifts)
				hrRoutes.POST("/roster", scheduleHandler.CreateShift)
			}

			// Services
			api.GET("/services", serviceHandler.List)
			api.POST("/services", serviceHandler.Create)
			api.GET("/services/appointments", serviceHandler.ListAppointments)
			api.POST("/services/appointments", serviceHandler.CreateAppointment)
			api.GET("/services/appointments/:id", serviceHandler.GetAppointment)
			api.GET("/services/commissions", serviceHandler.ListCommissions)
			api.POST("/services/commissions/mark-paid", serviceHandler.MarkCommissionsPaid)

			// Campaigns
			api.GET("/campaigns", marketingHandler.ListCampaigns)
			api.POST("/campaigns", marketingHandler.CreateCampaign)
			api.GET("/campaigns/:id", marketingHandler.GetCampaign)
			api.PUT("/campaigns/:id", marketingHandler.UpdateCampaign)
			api.DELETE("/campaigns/:id", marketingHandler.DeleteCampaign)
			api.POST("/campaigns/:id/send", marketingHandler.SendCampaign)
			api.POST("/campaigns/:id/open", marketingHandler.RecordCampaignOpen)
			api.POST("/campaigns/:id/convert", marketingHandler.RecordCampaignConversion)
			// Alias routes with /marketing/ prefix (used by frontend)
			api.GET("/marketing/campaigns", marketingHandler.ListCampaigns)
			api.POST("/marketing/campaigns", marketingHandler.CreateCampaign)
			api.PUT("/marketing/campaigns/:id", marketingHandler.UpdateCampaign)
			api.DELETE("/marketing/campaigns/:id", marketingHandler.DeleteCampaign)
			api.POST("/marketing/campaigns/:id/send", marketingHandler.SendCampaign)
			api.POST("/marketing/campaigns/:id/open", marketingHandler.RecordCampaignOpen)
			api.POST("/marketing/campaigns/:id/convert", marketingHandler.RecordCampaignConversion)
			api.POST("/marketing/trigger", marketingHandler.TriggerEventCampaigns)

			// Segments
			api.GET("/marketing/segments", marketingHandler.ListSegments)
			api.POST("/marketing/segments", marketingHandler.CreateSegment)
			api.PUT("/marketing/segments/:id", marketingHandler.UpdateSegment)
			api.DELETE("/marketing/segments/:id", marketingHandler.DeleteSegment)

			// Promotions
			api.GET("/promotions", marketingHandler.ListPromotions)
			api.POST("/promotions", marketingHandler.CreatePromotion)
			api.GET("/marketing/promotions", marketingHandler.ListPromotions)
			api.POST("/marketing/promotions", marketingHandler.CreatePromotion)
			api.PUT("/marketing/promotions/:id", marketingHandler.UpdatePromotion)
			api.DELETE("/marketing/promotions/:id", marketingHandler.DeletePromotion)

			// Discounts
			api.GET("/discounts", marketingHandler.ListDiscounts)
			api.POST("/discounts", marketingHandler.CreateDiscount)
			api.GET("/marketing/discounts", marketingHandler.ListDiscounts)
			api.POST("/marketing/discounts", marketingHandler.CreateDiscount)
			api.DELETE("/marketing/discounts/:id", marketingHandler.DeleteDiscount)

			// Loyalty Redemption
			api.POST("/marketing/redeem-points", marketingHandler.RedeemPointsForDiscount)

			// B2B
			api.GET("/b2b/quotes", b2bHandler.ListQuotes)
			api.POST("/b2b/quotes", b2bHandler.CreateQuote)
			api.GET("/b2b/quotes/:id", b2bHandler.GetQuote)
			api.POST("/b2b/quotes/:id/update", b2bHandler.UpdateQuote)
			api.POST("/b2b/quotes/:id/convert", b2bHandler.ConvertQuoteToOrder)
			api.POST("/b2b/bulk-order", b2bHandler.BulkOrder)
			api.GET("/b2b/clients", b2bHandler.ListClients)

			// Analytics
			analyticsGroup := api.Group("/analytics")
			{
				analyticsGroup.GET("/dashboard", analyticsHandler.Dashboard)
				analyticsGroup.GET("/sales-trends", analyticsHandler.SalesTrends)
				analyticsGroup.GET("/sales", analyticsHandler.SalesTrends) // alias
				analyticsGroup.GET("/revenue-breakdown", analyticsHandler.RevenueBreakdown)
				analyticsGroup.GET("/revenue", analyticsHandler.RevenueBreakdown) // alias
				analyticsGroup.GET("/top-products", analyticsHandler.TopProducts)
				analyticsGroup.GET("/customer-metrics", analyticsHandler.CustomerMetrics)
				analyticsGroup.GET("/customers", analyticsHandler.CustomerMetrics) // alias
				analyticsGroup.GET("/real-time", analyticsHandler.RealTimeMetrics)
				analyticsGroup.GET("/heatmap", analyticsHandler.SalesHeatmap)
				analyticsGroup.POST("/report-builder", analyticsHandler.ReportBuilder)
				analyticsGroup.GET("/staff-performance", analyticsHandler.StaffPerformance)
				analyticsGroup.GET("/sales-goal", analyticsHandler.SalesGoalProgress)
			}

			// Intelligence
			api.GET("/intelligence/inventory-forecast", intelligenceHandler.InventoryForecast)
			api.GET("/intelligence/pos-recommendations", intelligenceHandler.POSRecommendations)
			api.GET("/intelligence/staff-leaderboard", intelligenceHandler.StaffLeaderboard)
			api.GET("/intelligence/customer-segmentation", intelligenceHandler.CustomerSegmentation)
			api.GET("/intelligence/dynamic-pricing", intelligenceHandler.DynamicPricing)
			api.POST("/intelligence/dynamic-pricing/apply", intelligenceHandler.ApplyPricingAction)
			api.POST("/intelligence/dynamic-pricing/apply-bulk", intelligenceHandler.BulkApplyPricingAction)
			api.POST("/intelligence/auto-po", intelligenceHandler.GenerateAutoPOs)
			api.GET("/intelligence/anomalies", intelligenceHandler.GetAnomalies)
			api.GET("/intelligence/anomalies/stats", intelligenceHandler.GetAnomalyStats)

			// Copilot
			api.POST("/copilot/chat", copilotHandler.Chat)

			// Onboarding
			onboardingHandler := handlers.NewOnboardingHandler(db)
			api.GET("/onboarding/status", onboardingHandler.OnboardingStatus)

			// Content CMS routes
			api.GET("/content", contentHandler.ListPages)

			// Storefront Admin API
			api.PUT("/storefront/config", storefrontHandler.UpdateSettings)
			api.PUT("/storefront/settings", storefrontHandler.UpdateSettings)

			// Wallet
			api.GET("/wallet/dashboard", walletHandler.Dashboard)
			api.POST("/wallet/lookup", walletHandler.LookupCustomer)
			api.GET("/wallet/balance", walletHandler.GetBalance)
			api.GET("/wallet/transactions", walletHandler.ListTransactions)
			api.POST("/wallet/topup", walletHandler.TopUp)
			api.POST("/wallet/transfer", walletHandler.Transfer)
			api.POST("/wallet/customers/:customer_id/loyalty/adjust", walletHandler.AdjustLoyaltyPoints)
			api.POST("/wallet/customers/:customer_id/store-credit/adjust", walletHandler.AdjustStoreCredit)
			api.GET("/wallet/customers/:customer_id/gift-cards", walletHandler.GetGiftCards)
			api.POST("/wallet/gift-cards", walletHandler.CreateGiftCard)

			// Settings / Domains
			api.GET("/settings", settingsHandler.GetSettings)
			api.PUT("/settings", settingsHandler.UpdateSettings)
			api.GET("/settings/domains", adminHandler.ListDomains)
			api.POST("/settings/domains", adminHandler.CreateDomain)
			api.DELETE("/settings/domains/:id", adminHandler.DeleteDomain)
			api.POST("/settings/domains/:id/verify", adminHandler.VerifyDomain)
			api.POST("/settings/domains/:id/primary", adminHandler.SetPrimaryDomain)

			// Automated Reports & Z-Reports
			api.GET("/reports/schedules", reportScheduleHandler.GetSchedules)
			api.POST("/reports/schedules", reportScheduleHandler.SaveSchedule)
			api.POST("/reports/send-test", reportScheduleHandler.SendTestReport)
			api.GET("/reports/daily-z", reportScheduleHandler.GetDailyZReportData)
		}

		// Public Storefront Endpoints (Requires Tenant, NO POS Auth)
		storefrontPublic := v1.Group("/storefront")
		storefrontPublic.Use(middleware.TenantMiddleware())
		storefrontPublic.Use(middleware.RequireTenantMiddleware())
		{
			storefrontPublic.GET("/config", storefrontHandler.GetSettings)
			storefrontPublic.GET("/settings", storefrontHandler.GetSettings)
			storefrontPublic.GET("/categories", storefrontHandler.ListCategories)
			storefrontPublic.GET("/products", storefrontHandler.ListProducts)
			storefrontPublic.GET("/products/:id", storefrontHandler.GetProduct)
			storefrontPublic.GET("/cart", storefrontHandler.GetCart)
			storefrontPublic.POST("/cart/add", storefrontHandler.AddToCart)
			storefrontPublic.PUT("/cart/update", storefrontHandler.UpdateCart)
			storefrontPublic.PUT("/cart/email", storefrontHandler.UpdateCartEmail)
			storefrontPublic.DELETE("/cart/remove/:id", storefrontHandler.RemoveFromCart)
			storefrontPublic.POST("/checkout/verify", storefrontHandler.VerifyPaystackCheckout)
			storefrontPublic.POST("/checkout/convert-guest", storefrontHandler.ConvertGuestToAccount)
			storefrontPublic.POST("/wishlist/toggle/:id", storefrontHandler.ToggleWishlist)
			storefrontPublic.POST("/reviews/:product_id", storefrontHandler.SubmitReview)
			storefrontPublic.POST("/coupons/apply", storefrontHandler.ApplyCoupon)
			storefrontPublic.POST("/coupon/apply", storefrontHandler.ApplyCoupon) // alias
			storefrontPublic.GET("/coupons", storefrontHandler.ListCoupons)       // alias
			storefrontPublic.POST("/coupons/remove", storefrontHandler.RemoveCoupon)
			storefrontPublic.POST("/products/:id/reviews", storefrontHandler.SubmitReview)
			storefrontPublic.POST("/products/:id/notify", storefrontHandler.SubscribeBackInStock)
			storefrontPublic.POST("/newsletter/subscribe", storefrontHandler.SubscribeNewsletter)

			// Customer Auth
			storefrontAuth := storefrontPublic.Group("/auth")
			storefrontAuth.Use(middleware.RateLimitMiddleware(redisClient, 10, time.Minute, 15*time.Minute))
			{
				storefrontAuth.POST("/register", storefrontHandler.RegisterCustomer)
				storefrontAuth.POST("/login", storefrontHandler.LoginCustomer)
			}
		}

		// Authenticated Storefront Customer Endpoints (Requires Tenant + Customer Auth)
		storefrontCustomer := v1.Group("/storefront/me")
		storefrontCustomer.Use(middleware.TenantMiddleware())
		storefrontCustomer.Use(middleware.RequireTenantMiddleware())
		storefrontCustomer.Use(middleware.AuthMiddleware(authService))
		{
			storefrontCustomer.Use(middleware.RoleMiddleware("customer"))
			storefrontCustomer.GET("", storefrontHandler.GetCustomerMe)
			storefrontCustomer.PUT("", storefrontHandler.UpdateCustomerMe)
			storefrontCustomer.GET("/orders", storefrontHandler.GetCustomerOrders)
			storefrontCustomer.GET("/wishlist", storefrontHandler.GetCustomerWishlist)
			storefrontCustomer.POST("/wishlist/:id", storefrontHandler.ToggleCustomerWishlist)
		}

		// WebSockets (Requires Auth)
		v1.GET("/ws", func(c *gin.Context) {
			websocket.ServeWs(hub, c)
		})

		// Webhooks (No Auth — Gap #7: These endpoints MUST verify webhook signatures
		// using the provider's signing secret before processing in production.)
		webhooksGroup := v1.Group("/webhooks")
		webhooksGroup.Use(middleware.RateLimitMiddleware(redisClient, 100, time.Minute, 5*time.Minute))
		{
			webhooksGroup.POST("/stripe", billingHandler.StripeWebhook)

			paystackWebhookHandler := handlers.NewPaystackWebhookHandler(db)
			webhooksGroup.POST("/paystack", paystackWebhookHandler.Handle)
		}

		// Exports (require auth + tenant)
		exports := v1.Group("/export")
		exports.Use(middleware.AuthMiddleware(authService))
		exports.Use(middleware.TenantMiddleware())
		{
			exports.GET("/orders", exportHandler.ExportOrdersCSV)
			exports.GET("/products", exportHandler.ExportProductsCSV)
			exports.GET("/sales", exportHandler.ExportSalesCSV)
			exports.GET("/inventory", exportHandler.ExportInventoryCSV)
			exports.GET("/customers", exportHandler.ExportCustomersCSV)
			exports.GET("/order-items", exportHandler.ExportOrderItemsCSV)
		}

		// Delivery & Fleet
		delivery := v1.Group("/delivery")
		delivery.Use(middleware.AuthMiddleware(authService))
		delivery.Use(middleware.TenantMiddleware())
		{
			delivery.GET("/drivers", deliveryHandler.ListDrivers)
			delivery.POST("/drivers", deliveryHandler.AddDriver)
			delivery.POST("/dispatch", deliveryHandler.DispatchOrder)
			delivery.GET("/orders", deliveryHandler.ListDispatchedOrders)
			delivery.POST("/orders/:id/complete", deliveryHandler.CompleteDelivery)
		}

		// Barcodes (public product barcode alias — kept for backwards compat)
		// NOTE: Main barcode endpoints moved under auth (Gap #6).
		// These public aliases only serve pre-generated barcodes by product ID.
		barcodePublic := v1.Group("/barcode")
		barcodePublic.Use(middleware.RateLimitMiddleware(redisClient, 30, time.Minute, 15*time.Minute))
		barcodePublic.Use(middleware.TenantMiddleware())
		{
			barcodePublic.GET("/product/:id", barcodeHandler.GenerateProductBarcode)
		}

		// Security (require auth)
		security := api.Group("/security")
		{
			security.POST("/2fa/setup", securityHandler.Setup2FA)
			security.POST("/2fa/verify", securityHandler.Verify2FA)
			security.POST("/2fa/disable", securityHandler.Disable2FA)
			security.GET("/audit-logs", securityHandler.ListAuditLogs)
			security.GET("/backup", securityHandler.BackupDashboard)
			security.POST("/backup/restore", securityHandler.RestoreBackup)
		}

		// ──────────────────────────────────────────────
		// Roles & Permissions API
		// ──────────────────────────────────────────────
		roles := api.Group("/roles")
		roles.Use(middleware.RequirePermission("roles:read"))
		{
			roles.GET("", roleHandler.ListRoles)
			roles.POST("", middleware.RequirePermission("roles:write"), roleHandler.CreateRole)
			roles.PUT("/:id", middleware.RequirePermission("roles:write"), roleHandler.UpdateRole)
			roles.DELETE("/:id", middleware.RequirePermission("roles:delete"), roleHandler.DeleteRole)
		}

		permissions := api.Group("/permissions")
		permissions.Use(middleware.RequirePermission("roles:read"))
		{
			permissions.GET("", roleHandler.ListPermissions)
		}

		// Privacy (require auth)
		privacy := v1.Group("/privacy")
		privacy.Use(middleware.AuthMiddleware(authService))
		privacy.Use(middleware.TenantMiddleware())
		{
			privacy.POST("/export", privacyHandler.ExportData)
			privacy.POST("/delete-account", privacyHandler.DeleteAccount)
			privacy.POST("/anonymize/:customer_id", privacyHandler.AnonymizeCustomer)
		}

		// Public Portal (no auth needed)
		public := v1.Group("/public")
		public.Use(middleware.RateLimitMiddleware(redisClient, 20, time.Minute, 15*time.Minute)) // Strict rate limit for tenant discovery
		
		// Global public routes that do not require a tenant
		public.GET("/receipts/:token", orderHandler.GetPublicReceipt)

		// Tenant-specific public routes
		public.Use(middleware.TenantMiddleware())
		{
			public.GET("/tenant-info", publicPortalHandler.GetTenantInfo)
			public.GET("/products", publicPortalHandler.ListProducts)
			public.GET("/track-order", publicPortalHandler.TrackOrder)
			public.POST("/feedback", publicPortalHandler.SubmitFeedback)
		}

		// Public Delivery Tracking (No Auth)
		publicDelivery := v1.Group("/track")
		publicDelivery.Use(middleware.RateLimitMiddleware(redisClient, 30, time.Minute, 15*time.Minute))
		publicDelivery.Use(middleware.TenantMiddleware())
		{
			publicDelivery.GET("/:token", deliveryHandler.TrackOrder)
		}

		// Integrations (auth + tenant required)
		integrations := v1.Group("/integrations")
		integrations.Use(middleware.AuthMiddleware(authService))
		integrations.Use(middleware.TenantMiddleware())
		{
			integrations.GET("", integrationHandler.ListIntegrations)
			integrations.GET("/xero/connect", integrationHandler.ConnectXero)
			integrations.GET("/xero/callback", integrationHandler.XeroCallback)
			integrations.POST("/:provider/sync", integrationHandler.TriggerSync)
		}

		// Offline-First Sync (CRDT)
		offlineSync := v1.Group("/sync")
		offlineSync.Use(middleware.AuthMiddleware(authService))
		offlineSync.Use(middleware.TenantMiddleware())
		{
			offlineSync.POST("", syncHandler.SyncData)
		}

		// Omnichannel (Social Commerce)
		omnichannel := v1.Group("/omnichannel")
		{
			// Webhooks don't use auth middleware (they verify their own signatures)
			omnichannel.POST("/whatsapp/webhook", omnichannelHandler.WhatsAppWebhook)

			// Authorized routes
			authorizedOmni := omnichannel.Group("")
			authorizedOmni.Use(middleware.AuthMiddleware(authService))
			authorizedOmni.Use(middleware.TenantMiddleware())
			{
				authorizedOmni.POST("/tiktok/sync", omnichannelHandler.SyncTikTokCatalog)
			}
		}

		// Admin-only routes
		admin := v1.Group("/admin")
		admin.Use(middleware.AdminAuthMiddleware(db, authService))
		{
			admin.GET("/health", adminHandler.SystemHealth)
			admin.GET("/dashboard", adminHandler.GetDashboardStats)

			// Tenant Management
			admin.GET("/tenants", middleware.RequireAdminPermission("tenants:read"), adminHandler.ListTenants)
			admin.GET("/tenants/search", middleware.RequireAdminPermission("tenants:read"), adminHandler.SearchTenants)
			admin.POST("/tenants", middleware.RequireAdminPermission("tenants:write"), adminHandler.CreateTenant)
			admin.GET("/tenants/:id", middleware.RequireAdminPermission("tenants:read"), adminHandler.GetTenantDetail)
			admin.PUT("/tenants/:id/notes", middleware.RequireAdminPermission("tenants:write"), adminHandler.UpdateTenantNotes)
			admin.POST("/tenants/:id/suspend", middleware.RequireAdminPermission("tenants:write"), adminHandler.SuspendTenant)
			admin.POST("/tenants/:id/impersonate", middleware.RequireAdminPermission("tenants:write"), adminHandler.ImpersonateTenant)

			admin.GET("/plans", adminHandler.ListPlans)
			admin.POST("/plans", adminHandler.CreatePlan)
			admin.GET("/broadcasts", adminHandler.ListBroadcasts)
			admin.POST("/broadcasts", adminHandler.CreateBroadcast)
			admin.POST("/feature-flags", adminHandler.UpdateFeatureFlags)

			admin.GET("/pricing-plans", adminHandler.ListPricingPlans)
			admin.POST("/pricing-plans", adminHandler.CreatePricingPlan)
			admin.PUT("/pricing-plans/:id", adminHandler.UpdatePricingPlan)
			admin.DELETE("/pricing-plans/:id", adminHandler.DeletePricingPlan)

			// Billing & Subscriptions
			admin.GET("/subscriptions", middleware.RequireAdminPermission("billing:read"), adminHandler.ListSubscriptions)
			admin.POST("/subscriptions/:id/override", middleware.RequireAdminPermission("billing:write"), adminHandler.OverrideSubscription)
			admin.GET("/subscriptions/upcoming-renewals", middleware.RequireAdminPermission("billing:read"), adminHandler.ListUpcomingRenewals)
			admin.GET("/payments", middleware.RequireAdminPermission("billing:read"), adminHandler.ListBillingPayments)
			admin.GET("/payments/failed", middleware.RequireAdminPermission("billing:read"), adminHandler.ListFailedPayments)
			admin.GET("/promo-codes", middleware.RequireAdminPermission("promo_codes:read"), adminHandler.ListPromoCodes)
			admin.POST("/promo-codes", middleware.RequireAdminPermission("promo_codes:write"), adminHandler.CreatePromoCode)
			admin.POST("/promo-codes/:id/toggle", middleware.RequireAdminPermission("promo_codes:write"), adminHandler.TogglePromoCode)

			// Content & Security
			admin.GET("/audit-logs", adminHandler.ListAuditLogs)
			admin.GET("/faqs", adminHandler.ListFAQs)
			admin.POST("/faqs", adminHandler.CreateFAQ)
			admin.POST("/faqs/:id/toggle", adminHandler.ToggleFAQ)

			// Legal Documents
			admin.GET("/legal", adminHandler.ListLegalDocuments)
			admin.PUT("/legal/:type", adminHandler.UpsertLegalDocument)

			// Blog Posts
			admin.GET("/blog", adminHandler.ListAdminBlogPosts)
			admin.POST("/blog", adminHandler.CreateBlogPost)
			admin.GET("/blog/:id", adminHandler.GetAdminBlogPost)
			admin.PUT("/blog/:id", adminHandler.UpdateBlogPost)
			admin.DELETE("/blog/:id", adminHandler.DeleteBlogPost)

			// Integrations & Systems
			admin.GET("/apps", adminHandler.ListExternalSystems)
			admin.POST("/apps/:id/toggle", adminHandler.ToggleExternalSystem)
			// System Backups
			admin.GET("/backups", adminHandler.ListDatabaseBackups)
			admin.POST("/backups/trigger", adminHandler.TriggerDatabaseBackup)
			admin.GET("/backups/:id/download", adminHandler.DownloadDatabaseBackup)

			// Growth
			admin.GET("/referrals", adminHandler.ListReferrals)

			// Webhook Event Log
			admin.GET("/webhook-events", adminHandler.ListWebhookEvents)

			// Security & Access Control
			admin.GET("/admin-roles", middleware.RequireAdminPermission("security:read"), adminHandler.ListAdminRoles)
			admin.POST("/admin-roles", middleware.RequireAdminPermission("security:write"), adminHandler.CreateAdminRole)
			admin.PUT("/admin-roles/:id", middleware.RequireAdminPermission("security:write"), adminHandler.UpdateAdminRole)

			admin.GET("/users", middleware.RequireAdminPermission("admin_users:read"), adminHandler.ListAdminUsers)
			admin.POST("/users", middleware.RequireAdminPermission("admin_users:write"), adminHandler.CreateAdminUser)
			admin.PUT("/users/:id/role", middleware.RequireAdminPermission("admin_users:write"), adminHandler.UpdateAdminUserRole)
			admin.DELETE("/users/:id", middleware.RequireAdminPermission("admin_users:write"), adminHandler.DeleteAdminUser)
			admin.GET("/ip-allowlist", middleware.RequireAdminPermission("security:read"), adminHandler.ListIPAllowlist)
			admin.POST("/ip-allowlist", middleware.RequireAdminPermission("security:write"), adminHandler.AddIPToAllowlist)
			admin.DELETE("/ip-allowlist/:id", middleware.RequireAdminPermission("security:write"), adminHandler.RemoveIPFromAllowlist)

			// API Keys
			admin.GET("/api-keys", middleware.RequireAdminPermission("api_keys:read"), adminHandler.ListMasterAPIKeys)
			admin.POST("/api-keys", middleware.RequireAdminPermission("api_keys:write"), adminHandler.CreateMasterAPIKey)
			admin.DELETE("/api-keys/:id", middleware.RequireAdminPermission("api_keys:write"), adminHandler.RevokeMasterAPIKey)

			// Domains
			admin.GET("/domains", middleware.RequireAdminPermission("domains:read"), adminHandler.SearchAllDomains)
			admin.POST("/domains/verify-bulk", middleware.RequireAdminPermission("domains:write"), adminHandler.BulkVerifyDomains)
			admin.GET("/domains/:id/diagnostics", middleware.RequireAdminPermission("domains:read"), adminHandler.GetDomainDiagnostics)

			// Telemetry Logs
			admin.GET("/system-traces", middleware.RequireAdminPermission("security:read"), adminHandler.ListTelemetryLogs)

			// Webhook retry
			admin.POST("/webhook-events/:id/retry", middleware.RequireAdminPermission("webhooks:write"), adminHandler.RetryWebhookEvent)

			// Gift Cards (Admin)
			admin.GET("/gift-cards", middleware.RequireAdminPermission("billing:read"), adminHandler.ListAllGiftCards)
			admin.POST("/gift-cards", middleware.RequireAdminPermission("billing:write"), adminHandler.CreateGiftCardAdmin)
			admin.POST("/gift-cards/:id/disable", middleware.RequireAdminPermission("billing:write"), adminHandler.DisableGiftCardAdmin)
		}

		// Public Marketing Endpoints
		publicMarketing := v1.Group("/marketing")
		{
			publicMarketing.GET("/pricing-plans", adminHandler.ListPricingPlans)
		}

		// Public Legal Endpoints (no auth)
		publicLegal := v1.Group("/public/legal")
		{
			publicLegal.GET("/:type", adminHandler.GetPublicLegalDocument)
		}

		// Public Blog Endpoints (no auth)
		publicBlog := v1.Group("/public/blog")
		{
			publicBlog.GET("", adminHandler.ListPublicBlogPosts)
			publicBlog.GET("/:slug", adminHandler.GetPublicBlogPost)
		}
	}

	return r
}
