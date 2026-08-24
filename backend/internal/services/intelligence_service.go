package services

import (
	"strings"
	"time"

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

// CalculateDynamicPricing suggests an optimal price based on velocity and base price.
func (s *IntelligenceService) CalculateDynamicPricing(product models.Product, velocity float64) float64 {
	// If product is flying off the shelves (velocity > 10 units/day), suggest a 5% markup
	if velocity > 10 {
		return product.SellingPrice * 1.05
	}
	// If product is dead stock (velocity < 1 unit/day) and we have stock, suggest 15% discount
	if velocity < 1 && product.CurrentStock > 10 {
		return product.SellingPrice * 0.85
	}
	return product.SellingPrice
}
