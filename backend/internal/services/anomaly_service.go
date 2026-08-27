package services

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

// Anomaly represents a detected anomaly in business metrics.
type Anomaly struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Type        string    `json:"type"`     // sales_drop, refund_spike, stock_shrinkage, unusual_login
	Severity    string    `json:"severity"` // warning, critical
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Metric      string    `json:"metric"`
	Baseline    float64   `json:"baseline"`  // expected value (7-day avg)
	Actual      float64   `json:"actual"`    // observed value
	Deviation   float64   `json:"deviation"` // percent deviation
	DetectedAt  time.Time `json:"detected_at"`
}

// DetectAnomalies runs all anomaly checks for a tenant and returns any detected anomalies.
func (s *IntelligenceService) DetectAnomalies(tenantID string) ([]Anomaly, error) {
	var anomalies []Anomaly

	salesAnomaly, err := s.detectSalesDrop(tenantID)
	if err == nil && salesAnomaly != nil {
		anomalies = append(anomalies, *salesAnomaly)
	}

	refundAnomaly, err := s.detectRefundSpike(tenantID)
	if err == nil && refundAnomaly != nil {
		anomalies = append(anomalies, *refundAnomaly)
	}

	shrinkageAnomalies, err := s.detectStockShrinkage(tenantID)
	if err == nil {
		anomalies = append(anomalies, shrinkageAnomalies...)
	}

	return anomalies, nil
}

// detectSalesDrop flags when today's revenue is >40% below the 7-day average.
func (s *IntelligenceService) detectSalesDrop(tenantID string) (*Anomaly, error) {
	type RevenueRow struct {
		Total float64
	}

	today := time.Now().Truncate(24 * time.Hour)
	sevenDaysAgo := today.AddDate(0, 0, -7)
	yesterday := today.AddDate(0, 0, -1)

	// Today's revenue
	var todayRow RevenueRow
	s.db.Table("orders").
		Select("COALESCE(SUM(total), 0) as total").
		Where("status = ? AND created_at >= ? AND deleted_at IS NULL", "completed", today).
		Scan(&todayRow)

	// 7-day average (excluding today)
	var historicalRow RevenueRow
	s.db.Table("orders").
		Select("COALESCE(SUM(total) / 7.0, 0) as total").
		Where("status = ? AND created_at >= ? AND created_at < ? AND deleted_at IS NULL",
			"completed", sevenDaysAgo, yesterday).
		Scan(&historicalRow)

	baseline := historicalRow.Total
	actual := todayRow.Total

	if baseline <= 0 {
		return nil, nil // not enough data
	}

	deviation := ((baseline - actual) / baseline) * 100
	if deviation < 40 {
		return nil, nil
	}

	severity := "warning"
	if deviation >= 70 {
		severity = "critical"
	}

	return &Anomaly{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		Type:        "sales_drop",
		Severity:    severity,
		Title:       "Significant Sales Drop Detected",
		Description: formatPct("Today's revenue is %.1f%% below the 7-day average. Review your POS and online store.", deviation),
		Metric:      "daily_revenue",
		Baseline:    math.Round(baseline*100) / 100,
		Actual:      math.Round(actual*100) / 100,
		Deviation:   math.Round(deviation*10) / 10,
		DetectedAt:  time.Now(),
	}, nil
}

// detectRefundSpike flags when today's refund count is >200% above 7-day average.
func (s *IntelligenceService) detectRefundSpike(tenantID string) (*Anomaly, error) {
	type CountRow struct {
		Count float64
	}

	today := time.Now().Truncate(24 * time.Hour)
	sevenDaysAgo := today.AddDate(0, 0, -7)
	yesterday := today.AddDate(0, 0, -1)

	var todayRow CountRow
	s.db.Table("orders").
		Select("COUNT(*) as count").
		Where("status IN ('refunded', 'partially_refunded') AND created_at >= ? AND deleted_at IS NULL", today).
		Scan(&todayRow)

	var historicalRow CountRow
	s.db.Table("orders").
		Select("COALESCE(COUNT(*) / 7.0, 0) as count").
		Where("status IN ('refunded', 'partially_refunded') AND created_at >= ? AND created_at < ? AND deleted_at IS NULL",
			sevenDaysAgo, yesterday).
		Scan(&historicalRow)

	baseline := historicalRow.Count
	actual := todayRow.Count

	if baseline <= 0 || actual < 3 {
		return nil, nil // ignore noise with very few refunds
	}

	deviation := ((actual - baseline) / baseline) * 100
	if deviation < 200 {
		return nil, nil
	}

	severity := "warning"
	if deviation >= 400 {
		severity = "critical"
	}

	return &Anomaly{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		Type:        "refund_spike",
		Severity:    severity,
		Title:       "Unusual Refund Activity Detected",
		Description: formatPct("Today's refund count (%.0f) is %.1f%% above the 7-day average. Investigate product quality or cashier activity.", actual, deviation),
		Metric:      "daily_refunds",
		Baseline:    math.Round(baseline*10) / 10,
		Actual:      math.Round(actual),
		Deviation:   math.Round(deviation*10) / 10,
		DetectedAt:  time.Now(),
	}, nil
}

// detectStockShrinkage finds products where stock decreased without a recorded sale or adjustment.
func (s *IntelligenceService) detectStockShrinkage(tenantID string) ([]Anomaly, error) {
	type ShrinkageRow struct {
		ProductID   string
		ProductName string
		Expected    float64
		Actual      float64
		Diff        float64
	}

	last24h := time.Now().Add(-24 * time.Hour)

	// Find products where total recorded outflow movements != the stock change
	var rows []ShrinkageRow
	s.db.Raw(`
		SELECT 
			p.id as product_id,
			p.name as product_name,
			COALESCE(SUM(ABS(sm.quantity)), 0) as expected,
			p.current_stock as actual,
			COALESCE(SUM(ABS(sm.quantity)), 0) - p.current_stock as diff
		FROM products p
		LEFT JOIN stock_movements sm ON sm.product_id = p.id 
			AND sm.quantity < 0
			AND sm.created_at >= ?
			AND sm.deleted_at IS NULL
		WHERE p.track_inventory = true 
			AND p.is_active = true
			AND p.deleted_at IS NULL
		GROUP BY p.id, p.name, p.current_stock
		HAVING COALESCE(SUM(ABS(sm.quantity)), 0) > 0 AND p.current_stock < 0
		LIMIT 10
	`, last24h).Scan(&rows)

	var anomalies []Anomaly
	for _, row := range rows {
		anomalies = append(anomalies, Anomaly{
			ID:          uuid.New().String(),
			TenantID:    tenantID,
			Type:        "stock_shrinkage",
			Severity:    "warning",
			Title:       "Possible Inventory Shrinkage: " + row.ProductName,
			Description: "Stock level went negative without a matching adjustment. Possible theft, damage, or data entry error.",
			Metric:      "stock_discrepancy",
			Baseline:    row.Expected,
			Actual:      row.Actual,
			Deviation:   row.Diff,
			DetectedAt:  time.Now(),
		})
	}
	return anomalies, nil
}

// GetAnomalyAlerts is the public-facing API endpoint helper.
func (s *IntelligenceService) GetAnomalyAlerts(tenantID string) ([]Anomaly, error) {
	return s.DetectAnomalies(tenantID)
}

// AnomalyStats provides a summary count for dashboard KPIs.
type AnomalyStats struct {
	Total    int `json:"total"`
	Critical int `json:"critical"`
	Warning  int `json:"warning"`
}

func (s *IntelligenceService) GetAnomalyStats(tenantID string) (*AnomalyStats, error) {
	anomalies, err := s.DetectAnomalies(tenantID)
	if err != nil {
		return &AnomalyStats{}, err
	}
	stats := &AnomalyStats{Total: len(anomalies)}
	for _, a := range anomalies {
		if a.Severity == "critical" {
			stats.Critical++
		} else {
			stats.Warning++
		}
	}
	return stats, nil
}

func formatPct(format string, args ...float64) string {
	if len(args) == 1 {
		return fmt.Sprintf(format, args[0])
	}
	return fmt.Sprintf(format, args[0], args[1])
}
