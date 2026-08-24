package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type ReturnService struct {
	db       *gorm.DB
	tenantID uuid.UUID
}

func NewReturnService(db *gorm.DB, tenantID uuid.UUID) *ReturnService {
	return &ReturnService{db: db, tenantID: tenantID}
}

type ReturnListParams struct {
	BranchID string
	Status   string
	Limit    int
	Offset   int
}

func (s *ReturnService) ListReturns(params ReturnListParams) ([]models.Return, int64, error) {
	query := s.db.Model(&models.Return{})
	if params.BranchID != "" {
		query = query.Where("branch_id = ?", params.BranchID)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	var total int64
	query.Count(&total)

	var returns []models.Return
	if err := query.Preload("Customer").Preload("Order").Preload("Items.Product").
		Order("created_at DESC").Offset(params.Offset).Limit(params.Limit).Find(&returns).Error; err != nil {
		return nil, 0, err
	}

	return returns, total, nil
}

type ReturnCreateInput struct {
	OrderID      uuid.UUID
	BranchID     *uuid.UUID
	CustomerID   *uuid.UUID
	Reason       string
	ReasonDetail string
	RefundMethod string
	RefundAmount float64
	Items        []models.ReturnItem
}

func (s *ReturnService) CreateReturn(input ReturnCreateInput) (*models.Return, error) {
	ret := models.Return{
		OrderID:      input.OrderID,
		CustomerID:   input.CustomerID,
		Reason:       input.Reason,
		ReasonDetail: &input.ReasonDetail,
		Status:       "pending",
		RefundMethod: input.RefundMethod,
		RefundAmount: input.RefundAmount,
	}
	ret.BranchScoped.BranchID = input.BranchID

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&ret).Error; err != nil {
			return err
		}

		for i := range input.Items {
			input.Items[i].ReturnID = ret.ID
			if err := tx.Create(&input.Items[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (s *ReturnService) GetReturn(id string) (*models.Return, error) {
	returnID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid return ID")
	}

	var ret models.Return
	if err := s.db.Where("id = ?", returnID).
		Preload("Customer").Preload("Order").Preload("Items.Product").
		First(&ret).Error; err != nil {
		return nil, errors.New("return not found")
	}

	return &ret, nil
}

func (s *ReturnService) ApproveReturn(id string) (*models.Return, error) {
	returnID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid return ID")
	}

	var ret models.Return
	if err := s.db.Where("id = ?", returnID).First(&ret).Error; err != nil {
		return nil, errors.New("return not found")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		ret.Status = "approved"
		ret.ApprovedAt = &now
		if err := tx.Save(&ret).Error; err != nil {
			return err
		}

		var items []models.ReturnItem
		if err := tx.Where("return_id = ? AND restock = ?", returnID, true).Find(&items).Error; err != nil {
			return err
		}
		for _, item := range items {
			if item.ProductID != nil {
				var product models.Product
				if err := tx.Where("id = ?", item.ProductID).First(&product).Error; err != nil {
					return err
				}

				// Check if the product belongs to the return branch
				var destProduct models.Product
				var newStock float64

				// Handle case where Return doesn't have a specific BranchID (unlikely but possible)
				branchID := s.tenantID // Fallback to tenant ID if branch is nil
				if ret.BranchID != nil {
					branchID = *ret.BranchID
				}

				if product.BranchID != nil && *product.BranchID == branchID {
					// The product already belongs to this branch
					newStock = product.CurrentStock + item.Quantity
					if err := tx.Model(&product).UpdateColumn("current_stock", newStock).Error; err != nil {
						return err
					}
					destProduct = product
				} else {
					// Need to find or create the product in the destination branch
					err := tx.Where("sku = ? AND branch_id = ?", product.SKU, branchID).First(&destProduct).Error
					if err != nil {
						// Create it
						destProduct = product
						destProduct.ID = uuid.New()
						destProduct.BranchID = &branchID
						destProduct.CurrentStock = item.Quantity
						if err := tx.Create(&destProduct).Error; err != nil {
							return err
						}
						newStock = item.Quantity
					} else {
						// Update existing
						newStock = destProduct.CurrentStock + item.Quantity
						if err := tx.Model(&destProduct).UpdateColumn("current_stock", newStock).Error; err != nil {
							return err
						}
					}
				}

				retIDStr := ret.ID.String()
				movement := models.StockMovement{
					TenantID:      s.tenantID,
					BranchID:      branchID,
					ProductID:     destProduct.ID,
					Quantity:      item.Quantity,
					PreviousStock: destProduct.CurrentStock,
					NewStock:      newStock,
					Reason:        "return",
					ReferenceID:   &retIDStr,
				}
				if err := tx.Create(&movement).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (s *ReturnService) RejectReturn(id string) (*models.Return, error) {
	returnID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid return ID")
	}

	var ret models.Return
	if err := s.db.Where("id = ?", returnID).First(&ret).Error; err != nil {
		return nil, errors.New("return not found")
	}

	ret.Status = "rejected"
	if err := s.db.Save(&ret).Error; err != nil {
		return nil, err
	}

	return &ret, nil
}

func (s *ReturnService) ProcessRefund(id string) (*models.Return, float64, error) {
	returnID, err := uuid.Parse(id)
	if err != nil {
		return nil, 0, errors.New("invalid return ID")
	}

	var ret models.Return
	if err := s.db.Where("id = ? AND status = ?", returnID, "approved").First(&ret).Error; err != nil {
		return nil, 0, errors.New("approved return not found")
	}

	netRefund := ret.RefundAmount - ret.RestockingFee

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if ret.RefundMethod == "store_credit" && ret.CustomerID != nil {
			if err := tx.Model(&models.Customer{}).Where("id = ?", ret.CustomerID).
				UpdateColumn("store_credit", gorm.Expr("store_credit + ?", netRefund)).Error; err != nil {
				return err
			}
		}

		now := time.Now()
		ret.Status = "completed"
		ret.CompletedAt = &now
		if err := tx.Save(&ret).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return &ret, netRefund, nil
}
