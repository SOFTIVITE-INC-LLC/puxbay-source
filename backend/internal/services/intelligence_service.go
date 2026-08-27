package services

import (
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type IntelligenceService struct {
	db *gorm.DB
}

func NewIntelligenceService(db *gorm.DB) *IntelligenceService {
	return &IntelligenceService{db: db}
}

type Forecast struct {
	Product  models.Product `json:"product"`
	DaysLeft float64        `json:"days_left"`
	Velocity float64        `json:"velocity"`
	Status   string         `json:"status"`
}

func (s *IntelligenceService) InventoryForecast(tenantID, branchID string) ([]Forecast, error) {
	var products []models.Product
	prodQuery := s.db.Preload("Category").Where("is_active = ?", true)
	if branchID != "" {
		prodQuery = prodQuery.Where("branch_id = ?", branchID)
	}
	if err := prodQuery.Find(&products).Error; err != nil {
		return nil, err
	}

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	type SalesAggregate struct {
		ProductID string
		TotalSold float64
	}
	var salesAgg []SalesAggregate

	s.db.Session(&gorm.Session{NewDB: true}).Table("order_items").
		Select("order_items.product_id, SUM(order_items.quantity) as total_sold").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.status = ? AND orders.created_at >= ? AND orders.deleted_at IS NULL AND order_items.deleted_at IS NULL", "completed", thirtyDaysAgo).
		Group("order_items.product_id").
		Scan(&salesAgg)

	salesMap := make(map[string]float64)
	for _, agg := range salesAgg {
		salesMap[agg.ProductID] = agg.TotalSold
	}

	var forecasts []Forecast
	for _, p := range products {
		if p.CurrentStock <= 0 {
			continue
		}
		totalSold := salesMap[p.ID.String()]
		velocity := totalSold / 30.0

		days := 999.0
		if velocity > 0 {
			days = float64(p.CurrentStock) / velocity
		}

		status := "healthy"
		if days < 7 || p.CurrentStock <= 0 {
			status = "critical"
		} else if days < 30 {
			status = "warning"
		}

		forecasts = append(forecasts, Forecast{
			Product:  p,
			DaysLeft: days,
			Velocity: velocity,
			Status:   status,
		})
	}

	return forecasts, nil
}

func (s *IntelligenceService) POSRecommendations(tenantID, productIDs string) ([]models.Product, error) {
	if productIDs == "" {
		return []models.Product{}, nil
	}

	ids := strings.Split(productIDs, ",")
	var validIDs []string
	for _, id := range ids {
		if id != "" {
			validIDs = append(validIDs, id)
		}
	}

	var recommendedProducts []models.Product

	subQuery := s.db.Table("order_items").
		Select("order_id").
		Where("product_id IN ?", validIDs)
	if err := s.db.Table("order_items").
		Select("products.*").
		Joins("JOIN products ON products.id = order_items.product_id").
		Where("order_items.order_id IN (?)", subQuery).
		Where("order_items.product_id NOT IN ?", validIDs).
		Group("products.id").
		Order("SUM(order_items.quantity) DESC").
		Limit(6).
		Find(&recommendedProducts).Error; err != nil {
		return nil, err
	}

	return recommendedProducts, nil
}

type LeaderboardEntry struct {
	CashierID  string  `json:"cashier_id"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	TotalSales float64 `json:"total_sales"`
	OrderCount int64   `json:"order_count"`
}

func (s *IntelligenceService) StaffLeaderboard(tenantID, branchID string, days int) ([]LeaderboardEntry, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	query := s.db.Model(&models.Order{}).
		Where("status = ? AND created_at >= ? AND cashier_id IS NOT NULL", "completed", startDate)

	if branchID != "" {
		query = query.Where("branch_id = ?", branchID)
	}

	var leaderboard []LeaderboardEntry

	if err := query.Select("orders.cashier_id, u.first_name, u.last_name, SUM(orders.total) as total_sales, COUNT(orders.id) as order_count").
		Joins("LEFT JOIN public.users u ON u.id = orders.cashier_id").
		Group("orders.cashier_id, u.first_name, u.last_name").
		Order("total_sales DESC").
		Scan(&leaderboard).Error; err != nil {
		return nil, err
	}

	return leaderboard, nil
}

type SegmentStats struct {
	Count int `json:"count"`
}

type CustomerSegmentationResult struct {
	Segments       map[string]SegmentStats
	TotalCustomers int
}

func (s *IntelligenceService) CustomerSegmentation(tenantID string) (*CustomerSegmentationResult, error) {
	type CustomerAgg struct {
		CustomerID string
		TotalSpent float64
		OrderCount int
		LastOrder  time.Time
	}

	var aggs []CustomerAgg
	if err := s.db.Table("orders").
		Select("customer_id, SUM(total) as total_spent, COUNT(id) as order_count, MAX(created_at) as last_order").
		Where("status = ? AND customer_id IS NOT NULL AND deleted_at IS NULL", "completed").
		Group("customer_id").
		Scan(&aggs).Error; err != nil {
		return nil, err
	}

	now := time.Now()

	segments := map[string]SegmentStats{
		"VIP":     {Count: 0},
		"Loyal":   {Count: 0},
		"Recent":  {Count: 0},
		"At Risk": {Count: 0},
		"Lost":    {Count: 0},
	}

	for _, agg := range aggs {
		daysSinceLast := now.Sub(agg.LastOrder).Hours() / 24.0

		segment := "Lost"
		if daysSinceLast < 30 {
			if agg.TotalSpent > 1000 && agg.OrderCount > 5 {
				segment = "VIP"
			} else if agg.OrderCount > 2 {
				segment = "Loyal"
			} else {
				segment = "Recent"
			}
		} else if daysSinceLast < 90 {
			if agg.TotalSpent > 500 {
				segment = "At Risk"
			} else {
				segment = "Lost"
			}
		}

		stat := segments[segment]
		stat.Count++
		segments[segment] = stat
	}

	return &CustomerSegmentationResult{
		Segments:       segments,
		TotalCustomers: len(aggs),
	}, nil
}

// GenerateMarketingCampaign uses a mocked LLM stub to generate copy.
func (s *IntelligenceService) GenerateMarketingCampaign(segment string) string {
	switch segment {
	case "VIP":
		return "Subject: Exclusive early access just for you! \nHi there, as one of our top customers, we wanted you to be the first to know..."
	case "At Risk":
		return "Subject: We miss you! \nHi! It's been a while since your last visit. Here is a 15% discount code to welcome you back..."
	default:
		return "Subject: Check out our latest products! \nVisit our store today to see what's new."
	}
}

type PricingSuggestion struct {
	ProductID      string  `json:"product_id"`
	ProductName    string  `json:"product_name"`
	SKU            string  `json:"sku"`
	CategoryName   string  `json:"category_name"`
	CurrentPrice   float64 `json:"current_price"`
	CostPrice      float64 `json:"cost_price"`
	SuggestedPrice float64 `json:"suggested_price"`
	ChangePercent  float64 `json:"change_percent"`
	Strategy       string  `json:"strategy"` // "surge", "clearance", "margin_recovery", "overstock", "competitive"
	Reason         string  `json:"reason"`
	Velocity       float64 `json:"velocity"`
	CurrentStock   float64 `json:"current_stock"`
}

type PricingActionItem struct {
	ProductID string  `json:"product_id"`
	NewPrice  float64 `json:"new_price"`
}

// GetDynamicPricing generates intelligent price recommendations for inventory products based on velocity, stock, and margins.
func (s *IntelligenceService) GetDynamicPricing(branchID string) ([]PricingSuggestion, error) {
	var products []models.Product
	prodQuery := s.db.Preload("Category").Where("is_active = ?", true)
	if branchID != "" {
		prodQuery = prodQuery.Where("branch_id = ?", branchID)
	}
	if err := prodQuery.Find(&products).Error; err != nil {
		return nil, err
	}

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	type SalesAggregate struct {
		ProductID string
		TotalSold float64
	}
	var salesAgg []SalesAggregate

	s.db.Session(&gorm.Session{NewDB: true}).Table("order_items").
		Select("order_items.product_id, SUM(order_items.quantity) as total_sold").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.status = ? AND orders.created_at >= ? AND orders.deleted_at IS NULL AND order_items.deleted_at IS NULL", "completed", thirtyDaysAgo).
		Group("order_items.product_id").
		Scan(&salesAgg)

	salesMap := make(map[string]float64)
	for _, agg := range salesAgg {
		salesMap[agg.ProductID] = agg.TotalSold
	}

	var suggestions []PricingSuggestion

	for _, p := range products {
		if p.SellingPrice <= 0 {
			continue
		}

		totalSold := salesMap[p.ID.String()]
		velocity := totalSold / 30.0

		categoryName := "General"
		if p.Category != nil && p.Category.Name != "" {
			categoryName = p.Category.Name
		}

		currentPrice := p.SellingPrice
		costPrice := p.CostPrice
		stock := p.CurrentStock

		var suggestedPrice float64
		var strategy string
		var reason string

		// Rule 1: High Demand Surge - Velocity is high (>= 2 units/day) and stock is under 14 days runway
		if velocity >= 2.0 && stock > 0 && (stock/velocity) < 14 {
			suggestedPrice = math.Round(currentPrice*1.12*100) / 100 // +12%
			strategy = "surge"
			reason = "High sales velocity with low inventory runway (<14 days)"
		} else if velocity <= 0.05 && stock >= 5 {
			// Rule 2: Slow Moving / Stagnant Stock Clearance - Little to no sales in 30 days
			minFloor := costPrice * 1.05
			discounted := math.Round(currentPrice*0.85*100) / 100 // -15%
			if minFloor > 0 && discounted < minFloor {
				discounted = math.Round(minFloor*100) / 100
			}
			if discounted < currentPrice {
				suggestedPrice = discounted
				strategy = "clearance"
				reason = "Stagnant inventory (<0.05 units/day) with excess stock"
			}
		} else if costPrice > 0 && ((currentPrice-costPrice)/costPrice) < 0.12 {
			// Rule 3: Margin Recovery - Margin is thin under 12%
			suggestedPrice = math.Round(costPrice*1.22*100) / 100 // Target 22% margin
			strategy = "margin_recovery"
			reason = "Sub-optimal profit margin (<12% markup on cost)"
		} else if stock > 25 && velocity > 0 && (stock/velocity) > 90 {
			// Rule 4: Overstock Optimization - Over 90 days of inventory sitting
			discounted := math.Round(currentPrice*0.90*100) / 100 // -10%
			if costPrice > 0 && discounted < costPrice*1.05 {
				discounted = math.Round(costPrice*1.05*100) / 100
			}
			if discounted < currentPrice {
				suggestedPrice = discounted
				strategy = "overstock"
				reason = "Overstocked inventory runway exceeding 90 days"
			}
		} else if velocity >= 0.5 && costPrice > 0 && ((currentPrice-costPrice)/costPrice) >= 0.15 {
			// Rule 5: Healthy Demand Fine-Tuning (+5% optimization)
			suggestedPrice = math.Round(currentPrice*1.05*100) / 100
			strategy = "competitive"
			reason = "Steady sales velocity with margin expansion potential"
		}

		if suggestedPrice > 0 && math.Abs(suggestedPrice-currentPrice) >= 0.01 {
			pct := ((suggestedPrice - currentPrice) / currentPrice) * 100
			suggestions = append(suggestions, PricingSuggestion{
				ProductID:      p.ID.String(),
				ProductName:    p.Name,
				SKU:            p.SKU,
				CategoryName:   categoryName,
				CurrentPrice:   currentPrice,
				CostPrice:      costPrice,
				SuggestedPrice: suggestedPrice,
				ChangePercent:  math.Round(pct*10) / 10,
				Strategy:       strategy,
				Reason:         reason,
				Velocity:       math.Round(velocity*100) / 100,
				CurrentStock:   stock,
			})
		}
	}

	return suggestions, nil
}

// ApplyPricingAction updates the product's selling price in the database.
func (s *IntelligenceService) ApplyPricingAction(productID string, newPrice float64) error {
	pID, err := uuid.Parse(productID)
	if err != nil {
		return err
	}
	return s.db.Model(&models.Product{}).Where("id = ?", pID).Update("selling_price", newPrice).Error
}

// BulkApplyPricing updates multiple product selling prices in a single transaction.
func (s *IntelligenceService) BulkApplyPricing(items []PricingActionItem) (int, error) {
	count := 0
	err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			pID, err := uuid.Parse(item.ProductID)
			if err != nil {
				continue
			}
			if err := tx.Model(&models.Product{}).Where("id = ?", pID).Update("selling_price", item.NewPrice).Error; err != nil {
				return err
			}
			count++
		}
		return nil
	})
	return count, err
}
