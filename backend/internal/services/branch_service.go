package services

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type BranchService struct {
	db *gorm.DB
}

func NewBranchService(db *gorm.DB) *BranchService {
	return &BranchService{db: db}
}

func (s *BranchService) ListBranches(limit, offset int) ([]models.Branch, int64, error) {
	var branches []models.Branch
	var total int64

	query := s.db.Model(&models.Branch{})
	query.Count(&total)

	if err := query.Order("created_at desc").Offset(offset).Limit(limit).Find(&branches).Error; err != nil {
		return nil, 0, err
	}
	return branches, total, nil
}

type BranchCreateInput struct {
	TenantID       uuid.UUID
	Name           string
	Address        string
	Phone          string
	PrimaryColor   string
	CurrencySymbol string
	CurrencyCode   string
	BranchType     string
}

func (s *BranchService) CreateBranch(input BranchCreateInput) (*models.Branch, error) {
	var maxBranches uint = 1

	if input.TenantID != uuid.Nil {
		var sub models.Subscription
		if err := s.db.Where("tenant_id = ? AND status IN ('active', 'trialing')", input.TenantID).First(&sub).Error; err != nil {
			return nil, errors.New("active subscription required to create a branch")
		}

		if sub.Status == "trialing" {
			maxBranches = 1
		} else if sub.PlanID != nil {
			var pricingPlan models.PricingPlan
			if err := s.db.Where("id = ?", *sub.PlanID).First(&pricingPlan).Error; err == nil {
				maxBranches = uint(pricingPlan.MaxBranches)
			}
		} else if sub.Plan != nil {
			maxBranches = sub.Plan.MaxBranches
		}

		if sub.CustomBranchesCount != nil {
			maxBranches = *sub.CustomBranchesCount
		}
	} else {
		// Superadmin / System bypass
		maxBranches = 999
	}

	var currentBranches int64
	s.db.Model(&models.Branch{}).Where("tenant_id = ?", input.TenantID).Count(&currentBranches)

	if currentBranches >= int64(maxBranches) {
		return nil, fmt.Errorf("branch limit reached. Your plan allows %d branches. Please upgrade your plan", maxBranches)
	}

	branch := models.Branch{
		TenantID:       input.TenantID,
		Name:           input.Name,
		PrimaryColor:   input.PrimaryColor,
		CurrencySymbol: input.CurrencySymbol,
		CurrencyCode:   input.CurrencyCode,
		BranchType:     input.BranchType,
	}

	if input.Address != "" {
		branch.Address = &input.Address
	}
	if input.Phone != "" {
		branch.Phone = &input.Phone
	}

	if err := s.db.Create(&branch).Error; err != nil {
		return nil, err
	}

	return &branch, nil
}

func (s *BranchService) GetBranch(id string) (*models.Branch, error) {
	var branch models.Branch
	if err := s.db.Where("id = ?", id).First(&branch).Error; err != nil {
		return nil, errors.New("branch not found")
	}
	return &branch, nil
}

type BranchUpdateInput struct {
	Name              string
	Address           string
	Phone             string
	PrimaryColor      string
	CurrencySymbol    string
	CurrencyCode      string
	ReceiptHeader     *string
	ReceiptFooter     *string
	LowStockThreshold *uint
}

func (s *BranchService) UpdateBranch(id string, input BranchUpdateInput) (*models.Branch, error) {
	branch, err := s.GetBranch(id)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		branch.Name = input.Name
	}
	if input.PrimaryColor != "" {
		branch.PrimaryColor = input.PrimaryColor
	}
	if input.CurrencySymbol != "" {
		branch.CurrencySymbol = input.CurrencySymbol
	}
	if input.CurrencyCode != "" {
		branch.CurrencyCode = input.CurrencyCode
	}
	if input.Address != "" {
		branch.Address = &input.Address
	}
	if input.Phone != "" {
		branch.Phone = &input.Phone
	}
	if input.ReceiptHeader != nil {
		branch.ReceiptHeader = input.ReceiptHeader
	}
	if input.ReceiptFooter != nil {
		branch.ReceiptFooter = input.ReceiptFooter
	}
	if input.LowStockThreshold != nil {
		branch.LowStockThreshold = *input.LowStockThreshold
	}

	if err := s.db.Save(branch).Error; err != nil {
		return nil, err
	}

	return branch, nil
}

func (s *BranchService) DeleteBranch(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.Branch{}).Error
}

type NetworkMetrics struct {
	TotalRevenue float64 `json:"total_revenue"`
	ActiveStaff  int64   `json:"active_staff"`
	OfflineCount int64   `json:"offline_count"`
}

func (s *BranchService) GetNetworkMetrics(tenantID string) (*NetworkMetrics, error) {
	var metrics NetworkMetrics

	// Today's total revenue across all branches
	s.db.Model(&models.Order{}).
		Where("status = ? AND DATE(created_at) = CURRENT_DATE", "completed").
		Select("COALESCE(SUM(total), 0)").Scan(&metrics.TotalRevenue)

	// Active staff (shifts with no end time)
	s.db.Model(&models.Shift{}).
		Joins("JOIN branches ON branches.id = shifts.branch_id").
		Where("shifts.end_time IS NULL").
		Count(&metrics.ActiveStaff)

	// Offline/Unhealthy branches
	s.db.Model(&models.Branch{}).
		Where("sync_status != 'healthy' OR pending_sync_count > 0").
		Count(&metrics.OfflineCount)

	return &metrics, nil
}

type BranchMetrics struct {
	TodayRevenue       float64                `json:"today_revenue"`
	TodayOrders        int64                  `json:"today_orders"`
	ActiveStaff        int64                  `json:"active_staff"`
	LowStockItems      int64                  `json:"low_stock_items"`
	TotalOrdersAlltime int64                  `json:"total_orders_alltime"`
	TotalRevenueAlltime float64               `json:"total_revenue_alltime"`
	RecentTransactions []DashboardTransaction `json:"recent_transactions"`
}

func (s *BranchService) GetBranchMetrics(branchID string) (*BranchMetrics, error) {
	var metrics BranchMetrics

	s.db.Model(&models.Order{}).
		Where("branch_id = ? AND status = ? AND DATE(created_at) = CURRENT_DATE", branchID, "completed").
		Select("COALESCE(SUM(total), 0)").Scan(&metrics.TodayRevenue)

	s.db.Model(&models.Order{}).
		Where("branch_id = ? AND status = ? AND DATE(created_at) = CURRENT_DATE", branchID, "completed").
		Count(&metrics.TodayOrders)

	s.db.Model(&models.Shift{}).
		Where("branch_id = ? AND end_time IS NULL", branchID).
		Count(&metrics.ActiveStaff)

	s.db.Model(&models.Product{}).
		Where("branch_id = ? AND is_active = ? AND current_stock <= reorder_level AND current_stock > 0", branchID, true).
		Count(&metrics.LowStockItems)

	s.db.Model(&models.Order{}).
		Where("branch_id = ? AND status = ?", branchID, "completed").
		Count(&metrics.TotalOrdersAlltime)

	s.db.Model(&models.Order{}).
		Where("branch_id = ? AND status = ?", branchID, "completed").
		Select("COALESCE(SUM(total), 0)").Scan(&metrics.TotalRevenueAlltime)

	var orders []models.Order
	s.db.Where("branch_id = ?", branchID).Preload("Customer").Order("created_at DESC").Limit(5).Find(&orders)

	var txs []DashboardTransaction
	for _, o := range orders {
		name := "Walk-in Customer"
		initials := "WK"

		if o.Customer != nil && o.Customer.Name != "" {
			name = o.Customer.Name
			initials = string(o.Customer.Name[0])
		}

		txType := "positive"
		if o.Status == "refunded" || o.Status == "cancelled" {
			txType = "negative"
		}

		txs = append(txs, DashboardTransaction{
			ID:       o.OrderNumber,
			Name:     name,
			Initials: initials,
			Amount:   fmt.Sprintf("%.2f", o.Total),
			Time:     o.CreatedAt.Format("03:04 PM"),
			Type:     txType,
		})
	}
	metrics.RecentTransactions = txs

	return &metrics, nil
}
