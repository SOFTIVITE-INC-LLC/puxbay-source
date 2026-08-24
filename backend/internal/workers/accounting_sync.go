package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type AccountingSyncPayload struct {
	TenantID uuid.UUID `json:"tenant_id"`
	Provider string    `json:"provider"`
	Date     string    `json:"date"` // YYYY-MM-DD
}

func NewAccountingSyncTask(tenantID uuid.UUID, provider string, date time.Time) (*asynq.Task, error) {
	p, err := json.Marshal(AccountingSyncPayload{
		TenantID: tenantID,
		Provider: provider,
		Date:     date.Format("2006-01-02"),
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeAccountingSync, p,
		asynq.MaxRetry(3),
		asynq.Queue("default"),
		asynq.Timeout(2*time.Minute),
	), nil
}

func (ws *WorkerServer) HandleAccountingSync(ctx context.Context, t *asynq.Task) error {
	var p AccountingSyncPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	targetDate, err := time.Parse("2006-01-02", p.Date)
	if err != nil {
		return fmt.Errorf("invalid date format: %w", err)
	}

	startOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	log.Printf("📊 Starting %s accounting sync for Tenant: %s (Date: %s)", p.Provider, p.TenantID, p.Date)

	var schemaName string
	if err := ws.db.Table("public.tenants").Select("schema_name").Where("id = ?", p.TenantID).Scan(&schemaName).Error; err != nil || schemaName == "" {
		return fmt.Errorf("failed to find tenant schema: %w", err)
	}

	return ws.db.Transaction(func(tx *gorm.DB) error {
		// Set search path for this transaction
		tx.Exec("SET LOCAL search_path TO " + schemaName)

		var integration models.TenantIntegration
		if err := tx.Where("provider = ? AND is_active = true", p.Provider).First(&integration).Error; err != nil {
			log.Printf("⚠️  Accounting integration for %s not found or inactive, skipping sync.", p.Provider)
			return nil
		}

		// 1. Fetch Orders for the day
		var orders []models.Order
		if err := tx.Preload("Items").Where("created_at >= ? AND created_at < ? AND status = 'completed'", startOfDay, endOfDay).Find(&orders).Error; err != nil {
			return fmt.Errorf("failed to fetch orders: %w", err)
		}

		// 2. Aggregate Sales Data
		var totalRevenue, totalTax, totalDiscount, totalCOGS float64
		paymentTotals := make(map[string]float64)

		for _, order := range orders {
			totalRevenue += order.Total
			totalTax += order.Tax
			totalDiscount += order.Discount
			paymentTotals[order.PaymentMethod] += order.Total

			for _, item := range order.Items {
				// Calculate COGS based on cost price at time of sale
				totalCOGS += (item.CostPrice * item.Quantity)
			}
		}

		// 3. Prepare Journal Entry (Stub for Xero/QBO)
		journalEntry := map[string]interface{}{
			"Date":            p.Date,
			"TotalRevenue":    totalRevenue,
			"TotalTax":        totalTax,
			"TotalDiscount":   totalDiscount,
			"TotalCOGS":       totalCOGS,
			"PaymentTotals":   paymentTotals,
			"IntegrationInfo": "Uses decrypted AccessToken to send HTTP POST to " + p.Provider,
		}

		jeJSON, _ := json.MarshalIndent(journalEntry, "", "  ")
		log.Printf("✅ Generated Journal Entry for %s:\n%s", p.Provider, string(jeJSON))

		// Update last sync time
		now := time.Now()
		tx.Model(&integration).Update("last_sync_at", &now)

		return nil
	})
}
