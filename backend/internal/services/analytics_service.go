package services

import (
	"fmt"
	"time"

	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type AnalyticsService struct {
	db *gorm.DB
}

func NewAnalyticsService(db *gorm.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

type DailyData struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
}

type SalesTrendsResult struct {
	CurrentRevenue  float64     `json:"current_revenue"`
	CurrentOrders   int64       `json:"current_orders"`
	PreviousRevenue float64     `json:"previous_revenue"`
	PreviousOrders  int64       `json:"previous_orders"`
	RevenueGrowth   float64     `json:"revenue_growth"`
	OrderGrowth     float64     `json:"order_growth"`
	DailyData       []DailyData `json:"daily_data"`
	Period          string      `json:"period"`
}

func (s *AnalyticsService) SalesTrends(tenantID, branchID, from, to string) (*SalesTrendsResult, error) {
	now := time.Now()
	var startDate, endDate time.Time

	if from != "" {
		if t, err := time.Parse("2006-01-02", from); err == nil {
			startDate = t
		}
	} else {
		startDate = now.AddDate(0, 0, -30)
	}

	if to != "" {
		if t, err := time.Parse("2006-01-02", to); err == nil {
			endDate = t.Add(24*time.Hour - time.Second)
		}
	} else {
		endDate = now
	}

	daysCount := int(endDate.Sub(startDate).Hours() / 24)
	if daysCount <= 0 {
		daysCount = 1
	}
	previousStart := startDate.AddDate(0, 0, -daysCount)
	db := s.db.Session(&gorm.Session{})

	var currentRevenue float64
	qRev := db.Model(&models.Order{}).Where("status = ?", "completed").Where("created_at >= ? AND created_at <= ?", startDate, endDate)
	if branchID != "" {
		qRev = qRev.Where("branch_id = ?", branchID)
	}
	qRev.Select("COALESCE(SUM(total), 0)").Scan(&currentRevenue)

	var currentCount int64
	qCnt := db.Model(&models.Order{}).Where("status = ?", "completed").Where("created_at >= ? AND created_at <= ?", startDate, endDate)
	if branchID != "" {
		qCnt = qCnt.Where("branch_id = ?", branchID)
	}
	qCnt.Count(&currentCount)

	var previousRevenue float64
	qPrevRev := db.Model(&models.Order{}).Where("status = ?", "completed").Where("created_at >= ? AND created_at < ?", previousStart, startDate)
	if branchID != "" {
		qPrevRev = qPrevRev.Where("branch_id = ?", branchID)
	}
	qPrevRev.Select("COALESCE(SUM(total), 0)").Scan(&previousRevenue)

	var previousCount int64
	qPrevCnt := db.Model(&models.Order{}).Where("status = ?", "completed").Where("created_at >= ? AND created_at < ?", previousStart, startDate)
	if branchID != "" {
		qPrevCnt = qPrevCnt.Where("branch_id = ?", branchID)
	}
	qPrevCnt.Count(&previousCount)

	var revenueGrowth float64
	if previousRevenue > 0 {
		revenueGrowth = ((currentRevenue - previousRevenue) / previousRevenue) * 100
	}
	var orderGrowth float64
	if previousCount > 0 {
		orderGrowth = (float64(currentCount-previousCount) / float64(previousCount)) * 100
	}

	var dailyData []DailyData
	qDaily := db.Model(&models.Order{}).Where("status = ?", "completed").Where("created_at >= ? AND created_at <= ?", startDate, endDate)
	if branchID != "" {
		qDaily = qDaily.Where("branch_id = ?", branchID)
	}
	qDaily.Select("DATE(created_at) as date, COALESCE(SUM(total), 0) as revenue, COUNT(id) as orders").
		Group("DATE(created_at)").
		Order("date").
		Scan(&dailyData)

	return &SalesTrendsResult{
		CurrentRevenue:  currentRevenue,
		CurrentOrders:   currentCount,
		PreviousRevenue: previousRevenue,
		PreviousOrders:  previousCount,
		RevenueGrowth:   revenueGrowth,
		OrderGrowth:     orderGrowth,
		DailyData:       dailyData,
		Period:          fmt.Sprintf("%s to %s", from, to),
	}, nil
}

type PaymentData struct {
	Method  string  `json:"method"`
	Revenue float64 `json:"revenue"`
	Count   int     `json:"count"`
}

type CategoryData struct {
	Name    string  `json:"name"`
	Revenue float64 `json:"revenue"`
}

type RevenueBreakdownResult struct {
	ByCategory      []CategoryData `json:"by_category"`
	ByPaymentMethod []PaymentData  `json:"by_payment_method"`
}

func (s *AnalyticsService) RevenueBreakdown(tenantID, branchID, startDateStr, endDateStr string) (*RevenueBreakdownResult, error) {
	db := s.db.Session(&gorm.Session{})
	baseQuery := db.Model(&models.Order{}).Where("orders.status = ?", "completed")
	if branchID != "" {
		baseQuery = baseQuery.Where("orders.branch_id = ?", branchID)
	}
	if startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			baseQuery = baseQuery.Where("orders.created_at >= ?", t)
		}
	}
	if endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			baseQuery = baseQuery.Where("orders.created_at <= ?", t.Add(24*time.Hour))
		}
	}

	var paymentRevenue []PaymentData
	baseQuery.Session(&gorm.Session{}).
		Select("payment_method as method, SUM(total) as revenue, COUNT(id) as count").
		Group("payment_method").
		Order("revenue DESC").
		Scan(&paymentRevenue)

	var categoryRevenue []CategoryData

	baseQuery.Session(&gorm.Session{}).
		Select("categories.name as name, SUM(order_items.unit_price * order_items.quantity) as revenue").
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Joins("LEFT JOIN categories ON categories.id = products.category_id").
		Group("categories.name").
		Order("revenue DESC").
		Scan(&categoryRevenue)

	return &RevenueBreakdownResult{
		ByCategory:      categoryRevenue,
		ByPaymentMethod: paymentRevenue,
	}, nil
}

type TopProductData struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	SKU       string  `json:"sku"`
	Quantity  float64 `json:"quantity"` // decimal(10,3) in DB — SUM returns "2.000", not an int
	Revenue   float64 `json:"revenue"`
}

type TopProductsResult struct {
	ByQuantity []TopProductData `json:"by_quantity"`
	ByRevenue  []TopProductData `json:"by_revenue"`
}

func (s *AnalyticsService) TopProducts(tenantID, branchID, from, to string, limit int) (*TopProductsResult, error) {
	if limit <= 0 {
		limit = 10
	}

	db := s.db.Session(&gorm.Session{})

	var topByQuantity []TopProductData
	var topByRevenue []TopProductData

	baseQuery := db.Table("order_items").
		Select("products.id as product_id, products.name, products.sku, SUM(order_items.quantity) as quantity, SUM(order_items.unit_price * order_items.quantity) as revenue").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Where("orders.status = ?", "completed")

	if from != "" {
		if t, err := time.Parse("2006-01-02", from); err == nil {
			baseQuery = baseQuery.Where("orders.created_at >= ?", t)
		}
	}
	if to != "" {
		if t, err := time.Parse("2006-01-02", to); err == nil {
			baseQuery = baseQuery.Where("orders.created_at <= ?", t.Add(24*time.Hour))
		}
	}

	if branchID != "" {
		baseQuery = baseQuery.Where("orders.branch_id = ?", branchID)
	}

	baseQuery = baseQuery.Group("products.id, products.name, products.sku")

	qQuant := s.db.Session(&gorm.Session{}).Table("order_items").
		Select("products.id as product_id, products.name, products.sku, SUM(order_items.quantity) as quantity, SUM(order_items.unit_price * order_items.quantity) as revenue").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Where("orders.status = ?", "completed")

	if branchID != "" {
		qQuant = qQuant.Where("orders.branch_id = ?", branchID)
	}
	qQuant.Group("products.id, products.name, products.sku").
		Order("quantity DESC").Limit(limit).Scan(&topByQuantity)

	qRev := s.db.Session(&gorm.Session{}).Table("order_items").
		Select("products.id as product_id, products.name, products.sku, SUM(order_items.quantity) as quantity, SUM(order_items.unit_price * order_items.quantity) as revenue").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Where("orders.status = ?", "completed")

	if branchID != "" {
		qRev = qRev.Where("orders.branch_id = ?", branchID)
	}
	qRev.Group("products.id, products.name, products.sku").
		Order("revenue DESC").Limit(limit).Scan(&topByRevenue)

	_ = baseQuery // suppress unused warning

	return &TopProductsResult{
		ByQuantity: topByQuantity,
		ByRevenue:  topByRevenue,
	}, nil
}

type CustomerMetricsResult struct {
	TotalCustomers  int64
	ActiveCustomers int64
	AvgOrderValue   float64
}

func (s *AnalyticsService) CustomerMetrics(tenantID, branchID string) (*CustomerMetricsResult, error) {
	db := s.db.Session(&gorm.Session{})

	var totalCustomers int64
	qCust := db.Model(&models.Customer{})
	if branchID != "" {
		qCust = qCust.Joins("JOIN orders ON orders.customer_id = customers.id").Where("orders.branch_id = ?", branchID).Distinct("customers.id")
	}
	qCust.Count(&totalCustomers)

	var activeCustomers int64
	qAct := db.Model(&models.Order{}).Where("status = ? AND customer_id IS NOT NULL", "completed")
	if branchID != "" {
		qAct = qAct.Where("branch_id = ?", branchID)
	}
	qAct.Distinct("customer_id").Count(&activeCustomers)

	var avgOrderValue float64
	qAvg := db.Model(&models.Order{}).Where("status = ?", "completed")
	if branchID != "" {
		qAvg = qAvg.Where("branch_id = ?", branchID)
	}
	qAvg.Select("COALESCE(AVG(total), 0)").Scan(&avgOrderValue)

	return &CustomerMetricsResult{
		TotalCustomers:  totalCustomers,
		ActiveCustomers: activeCustomers,
		AvgOrderValue:   avgOrderValue,
	}, nil
}

type RealTimeMetricsResult struct {
	TodayRevenue    float64
	TodayOrders     int64
	InventoryValue  float64
	LowStockCount   int64
	OutOfStockCount int64
	TotalProducts   int64
}

func (s *AnalyticsService) RealTimeMetrics(tenantID, branchID string) (*RealTimeMetricsResult, error) {
	today := time.Now().Truncate(24 * time.Hour)

	db := s.db.Session(&gorm.Session{})

	var todayRevenue float64
	qRev := db.Model(&models.Order{}).Where("status = ?", "completed").Where("created_at >= ?", today)
	if branchID != "" {
		qRev = qRev.Where("branch_id = ?", branchID)
	}
	qRev.Select("COALESCE(SUM(total), 0)").Scan(&todayRevenue)

	var todayOrders int64
	qOrd := db.Model(&models.Order{}).Where("status = ?", "completed").Where("created_at >= ?", today)
	if branchID != "" {
		qOrd = qOrd.Where("branch_id = ?", branchID)
	}
	qOrd.Count(&todayOrders)

	var inventoryValue float64
	qInv := db.Model(&models.Product{}).Where("is_active = ?", true)
	if branchID != "" {
		qInv = qInv.Where("branch_id = ?", branchID)
	}
	qInv.Select("COALESCE(SUM(current_stock * selling_price), 0)").Scan(&inventoryValue)

	var totalProducts int64
	qProd := db.Model(&models.Product{}).Where("is_active = ?", true)
	if branchID != "" {
		qProd = qProd.Where("branch_id = ?", branchID)
	}
	qProd.Count(&totalProducts)

	var lowStockCount int64
	qLow := db.Model(&models.Product{}).Where("is_active = ?", true).Where("current_stock <= reorder_level AND current_stock > 0")
	if branchID != "" {
		qLow = qLow.Where("branch_id = ?", branchID)
	}
	qLow.Count(&lowStockCount)

	var outOfStockCount int64
	qOut := db.Model(&models.Product{}).Where("is_active = ?", true).Where("current_stock = 0")
	if branchID != "" {
		qOut = qOut.Where("branch_id = ?", branchID)
	}
	qOut.Count(&outOfStockCount)

	return &RealTimeMetricsResult{
		TodayRevenue:    todayRevenue,
		TodayOrders:     todayOrders,
		InventoryValue:  inventoryValue,
		LowStockCount:   lowStockCount,
		OutOfStockCount: outOfStockCount,
		TotalProducts:   totalProducts,
	}, nil
}

type DashboardTransaction struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Initials string `json:"initials"`
	Amount   string `json:"amount"`
	Time     string `json:"time"`
	Type     string `json:"type"`
}

type DashboardOverviewResult struct {
	TotalSales         float64                `json:"total_sales"`
	TotalOrders        int64                  `json:"total_orders"`
	ActiveCustomers    int64                  `json:"active_customers"`
	LowStockItems      int64                  `json:"low_stock_items"`
	SalesTrend         float64                `json:"sales_trend"`
	RevenueChart       []float64              `json:"revenue_chart"`
	RecentTransactions []DashboardTransaction `json:"recent_transactions"`
}

func (s *AnalyticsService) DashboardOverview(tenantID, branchID string) (*DashboardOverviewResult, error) {
	// Re-use some existing logic
	realTime, err := s.RealTimeMetrics(tenantID, branchID)
	if err != nil {
		return nil, err
	}

	salesTrend, err := s.SalesTrends(tenantID, branchID, "", "")
	if err != nil {
		return nil, err
	}

	custMetrics, err := s.CustomerMetrics(tenantID, branchID)
	if err != nil {
		return nil, err
	}

	// Calculate 7-day revenue chart
	chart := make([]float64, 7)
	now := time.Now()
	for i := 0; i < 7; i++ {
		targetDate := now.AddDate(0, 0, -(6 - i))
		startOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)

		var dailyRev float64
		q := s.db.Model(&models.Order{}).Where("status = ?", "completed").
			Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay)
		if branchID != "" {
			q = q.Where("branch_id = ?", branchID)
		}
		q.Select("COALESCE(SUM(total), 0)").Scan(&dailyRev)
		chart[i] = dailyRev
	}

	// Fetch recent orders for transactions list
	var orders []models.Order
	qOrd := s.db.Preload("Customer").Order("created_at DESC").Limit(5)
	if branchID != "" {
		qOrd = qOrd.Where("branch_id = ?", branchID)
	}
	qOrd.Find(&orders)

	var txs []DashboardTransaction
	for _, o := range orders {
		name := "Walk-in Customer"
		initials := "WK"

		if o.Customer != nil && o.Customer.Name != "" {
			name = o.Customer.Name
			initials = string(o.Customer.Name[0])
		}

		timeStr := "Just now"
		diff := time.Since(o.CreatedAt)
		if diff.Hours() > 24 {
			timeStr = fmt.Sprintf("%.0fd ago", diff.Hours()/24)
		} else if diff.Hours() >= 1 {
			timeStr = fmt.Sprintf("%.0fh ago", diff.Hours())
		} else if diff.Minutes() >= 1 {
			timeStr = fmt.Sprintf("%.0fm ago", diff.Minutes())
		}

		txType := "positive"
		amtPrefix := "+"
		if o.Status == "refunded" || o.Status == "cancelled" {
			txType = "negative"
			amtPrefix = "-"
		}

		amtStr := fmt.Sprintf("%s$%.2f", amtPrefix, o.Total)

		txs = append(txs, DashboardTransaction{
			ID:       o.OrderNumber,
			Name:     name,
			Initials: initials,
			Amount:   amtStr,
			Time:     timeStr,
			Type:     txType,
		})
	}

	return &DashboardOverviewResult{
		TotalSales:         realTime.TodayRevenue,
		TotalOrders:        realTime.TodayOrders,
		ActiveCustomers:    custMetrics.ActiveCustomers,
		LowStockItems:      realTime.LowStockCount,
		SalesTrend:         salesTrend.RevenueGrowth,
		RevenueChart:       chart,
		RecentTransactions: txs,
	}, nil
}

type HeatmapData struct {
	Hour    int     `json:"hour"`
	Day     int     `json:"day"` // 0 = Sunday, 1 = Monday, etc.
	Revenue float64 `json:"revenue"`
}

func (s *AnalyticsService) SalesHeatmap(tenantID, branchID string) ([]HeatmapData, error) {
	db := s.db.Session(&gorm.Session{})

	// Default to last 30 days
	startDate := time.Now().AddDate(0, 0, -30)

	query := db.Model(&models.Order{}).Where("status = ?", "completed").Where("created_at >= ?", startDate)
	if branchID != "" {
		query = query.Where("branch_id = ?", branchID)
	}

	var rawData []struct {
		Hour    int
		Day     int
		Revenue float64
	}

	// PostgreSQL extraction of hour and DOW
	query.Select("EXTRACT(HOUR FROM created_at) as hour, EXTRACT(DOW FROM created_at) as day, SUM(total) as revenue").
		Group("EXTRACT(HOUR FROM created_at), EXTRACT(DOW FROM created_at)").
		Scan(&rawData)

	var heatmap []HeatmapData
	for _, d := range rawData {
		heatmap = append(heatmap, HeatmapData{
			Hour:    d.Hour,
			Day:     d.Day,
			Revenue: d.Revenue,
		})
	}

	return heatmap, nil
}

type StaffPerformanceData struct {
	StaffID     string  `json:"staff_id"`
	StaffName   string  `json:"staff_name"`
	Revenue     float64 `json:"revenue"`
	OrdersCount int64   `json:"orders_count"`
}

func (s *AnalyticsService) StaffPerformanceReport(tenantID, branchID, from, to string) ([]StaffPerformanceData, error) {
	db := s.db.Session(&gorm.Session{})

	baseQuery := db.Table("orders").
		Select("public.users.id as staff_id, public.users.first_name || ' ' || public.users.last_name as staff_name, SUM(orders.total) as revenue, COUNT(orders.id) as orders_count").
		Joins("JOIN public.users ON public.users.id = orders.cashier_id").
		Where("orders.status = ?", "completed").
		Where("orders.cashier_id IS NOT NULL")

	if branchID != "" {
		baseQuery = baseQuery.Where("orders.branch_id = ?", branchID)
	}

	if from != "" {
		if t, err := time.Parse("2006-01-02", from); err == nil {
			baseQuery = baseQuery.Where("orders.created_at >= ?", t)
		}
	}
	if to != "" {
		if t, err := time.Parse("2006-01-02", to); err == nil {
			baseQuery = baseQuery.Where("orders.created_at <= ?", t.Add(24*time.Hour))
		}
	}

	var results []StaffPerformanceData
	baseQuery.Group("public.users.id, public.users.first_name, public.users.last_name").
		Order("revenue DESC").
		Scan(&results)

	return results, nil
}

type SalesGoalProgress struct {
	Goal     float64 `json:"goal"`
	Current  float64 `json:"current"`
	Progress float64 `json:"progress"` // percentage
}

func (s *AnalyticsService) SalesGoalProgress(tenantID, branchID string) (*SalesGoalProgress, error) {
	// Fetch the configured goal for the tenant
	var settings models.CRMSettings
	goal := 50000.0 // Default fallback

	if err := s.db.First(&settings).Error; err == nil {
		if settings.MonthlySalesGoal > 0 {
			goal = settings.MonthlySalesGoal
		}
	}

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var currentRevenue float64
	q := s.db.Model(&models.Order{}).Where("status = ?", "completed").Where("created_at >= ?", startOfMonth)
	if branchID != "" {
		q = q.Where("branch_id = ?", branchID)
	}
	q.Select("COALESCE(SUM(total), 0)").Scan(&currentRevenue)

	progress := 0.0
	if goal > 0 {
		progress = (currentRevenue / goal) * 100
	}
	if progress > 100 {
		progress = 100
	}

	return &SalesGoalProgress{
		Goal:     goal,
		Current:  currentRevenue,
		Progress: progress,
	}, nil
}

type CustomReportResult struct {
	Headers []string                 `json:"headers"`
	Rows    []map[string]interface{} `json:"rows"`
}

func (s *AnalyticsService) GenerateCustomReport(tenantID, branchID string, metrics []string, dimensions []string, from, to string) (*CustomReportResult, error) {
	if len(metrics) == 0 {
		metrics = []string{"revenue"}
	}
	if len(dimensions) == 0 {
		dimensions = []string{"date"}
	}

	db := s.db.Session(&gorm.Session{})

	baseQuery := db.Table("orders").
		Where("orders.status = ?", "completed")

	if branchID != "" {
		baseQuery = baseQuery.Where("orders.branch_id = ?", branchID)
	}

	if from != "" {
		if t, err := time.Parse("2006-01-02", from); err == nil {
			baseQuery = baseQuery.Where("orders.created_at >= ?", t)
		}
	}
	if to != "" {
		if t, err := time.Parse("2006-01-02", to); err == nil {
			baseQuery = baseQuery.Where("orders.created_at <= ?", t.Add(24*time.Hour))
		}
	}

	var selectFields []string
	var groupFields []string

	// Map dimensions to SQL
	for _, dim := range dimensions {
		switch dim {
		case "date":
			selectFields = append(selectFields, "DATE(orders.created_at) as date")
			groupFields = append(groupFields, "DATE(orders.created_at)")
		case "payment_method":
			selectFields = append(selectFields, "orders.payment_method as payment_method")
			groupFields = append(groupFields, "orders.payment_method")
		case "staff":
			baseQuery = baseQuery.Joins("LEFT JOIN public.users ON public.users.id = orders.cashier_id")
			selectFields = append(selectFields, "COALESCE(public.users.first_name || ' ' || public.users.last_name, 'Online') as staff")
			groupFields = append(groupFields, "public.users.first_name, public.users.last_name")
		}
	}

	// Map metrics to SQL
	for _, met := range metrics {
		switch met {
		case "revenue":
			selectFields = append(selectFields, "SUM(orders.total) as revenue")
		case "orders":
			selectFields = append(selectFields, "COUNT(orders.id) as orders")
		case "discounts":
			selectFields = append(selectFields, "SUM(orders.discount) as discounts")
		case "tax":
			selectFields = append(selectFields, "SUM(orders.tax) as tax")
		}
	}

	// Construct SELECT
	selectStr := ""
	for i, f := range selectFields {
		if i > 0 {
			selectStr += ", "
		}
		selectStr += f
	}
	baseQuery = baseQuery.Select(selectStr)

	// Construct GROUP BY
	for _, g := range groupFields {
		baseQuery = baseQuery.Group(g)
	}

	// Order by the first dimension (usually date) descending if date is present
	hasDate := false
	for _, dim := range dimensions {
		if dim == "date" {
			hasDate = true
			break
		}
	}
	if hasDate {
		baseQuery = baseQuery.Order("date DESC")
	}

	var rows []map[string]interface{}
	if err := baseQuery.Find(&rows).Error; err != nil {
		return nil, err
	}

	headers := append(dimensions, metrics...)

	return &CustomReportResult{
		Headers: headers,
		Rows:    rows,
	}, nil
}
