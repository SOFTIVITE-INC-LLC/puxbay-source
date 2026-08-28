package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StockTransfer tracks moving inventory between branches.
type StockTransfer struct {
	TenantScoped
	ReferenceNo  string     `gorm:"size:50;uniqueIndex;not null" json:"reference_no"`
	FromBranchID uuid.UUID  `gorm:"type:uuid;not null;index" json:"from_branch_id"`
	ToBranchID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"to_branch_id"`
	Status       string     `gorm:"size:20;default:'pending';index" json:"status"` // pending, shipped, received, cancelled
	Notes        *string    `gorm:"type:text" json:"notes,omitempty"`
	CreatedByID  *uuid.UUID `gorm:"type:uuid" json:"created_by_id,omitempty"`
	CreatedBy    *User      `gorm:"foreignKey:CreatedByID" json:"createdby,omitempty"`
	ShippedAt    *time.Time `json:"shipped_at,omitempty"`
	ReceivedAt   *time.Time `json:"received_at,omitempty"`

	// Relations
	Items []StockTransferItem `gorm:"foreignKey:TransferID" json:"items,omitempty"`
}

// StockTransferItem is a line item in a transfer.
type StockTransferItem struct {
	Base
	TransferID uuid.UUID  `gorm:"type:uuid;not null;index" json:"transfer_id"`
	ProductID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"product_id"`
	Product    *Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	VariantID  *uuid.UUID `gorm:"type:uuid" json:"variant_id,omitempty"`
	Quantity   float64    `gorm:"type:decimal(10,4);not null" json:"quantity"`
}

// PurchaseOrder tracks ordering from suppliers.
type PurchaseOrder struct {
	BranchScoped
	PONumber     string     `gorm:"size:50;uniqueIndex;not null" json:"po_number"`
	SupplierID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"supplier_id"`
	Supplier     *Supplier  `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	Status       string     `gorm:"size:20;default:'draft';index" json:"status"` // draft, issued, partially_received, received, cancelled
	TotalAmount  float64    `gorm:"type:decimal(12,2);default:0" json:"total_amount"`
	ExpectedDate *time.Time `json:"expected_date,omitempty"`
	Notes        *string    `gorm:"type:text" json:"notes,omitempty"`

	// Relations
	Items []PurchaseOrderItem `gorm:"foreignKey:POID" json:"items,omitempty"`
}

// PurchaseOrderItem is a line item in a PO.
type PurchaseOrderItem struct {
	Base
	POID             uuid.UUID  `gorm:"type:uuid;not null;index" json:"po_id"`
	ProductID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"product_id"`
	Product          *Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	VariantID        *uuid.UUID `gorm:"type:uuid" json:"variant_id,omitempty"`
	QuantityOrdered  float64    `gorm:"type:decimal(10,4);not null" json:"quantity_ordered"`
	QuantityReceived float64    `gorm:"type:decimal(10,4);default:0" json:"quantity_received"`
	UnitCost         float64    `gorm:"type:decimal(12,2);not null" json:"unit_cost"`
}

// StocktakeSession tracks inventory counting.
type StocktakeSession struct {
	BranchScoped
	Name        string     `gorm:"size:100;not null" json:"name"`
	Status      string     `gorm:"size:20;default:'in_progress'" json:"status"` // in_progress, review, completed, cancelled
	Notes       *string    `gorm:"type:text" json:"notes,omitempty"`
	AccessToken *uuid.UUID `gorm:"type:uuid;uniqueIndex" json:"access_token,omitempty"`
	CreatedByID *uuid.UUID `gorm:"type:uuid" json:"created_by_id,omitempty"`
	CreatedBy   *User      `gorm:"foreignKey:CreatedByID" json:"createdby,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	Entries []StocktakeEntry `gorm:"foreignKey:SessionID" json:"entries,omitempty"`
}

// StockMovement tracks the audit trail of inventory changes.
type StockMovement struct {
	TenantScoped
	BranchID      *uuid.UUID `gorm:"type:uuid;index" json:"branch_id,omitempty"`
	Branch        *Branch    `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	ProductID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"product_id"`
	VariantID     *uuid.UUID `gorm:"type:uuid" json:"variant_id,omitempty"`
	Quantity      float64    `gorm:"type:decimal(10,4);not null" json:"quantity"` // positive (in) or negative (out)
	PreviousStock float64    `gorm:"type:decimal(10,4);not null" json:"previous_stock"`
	NewStock      float64    `gorm:"type:decimal(10,4);not null" json:"new_stock"`
	Reason        string     `gorm:"size:50;not null" json:"reason"`         // sale, return, transfer, adjustment, po_receipt
	ReferenceID   *string    `gorm:"size:100" json:"reference_id,omitempty"` // Order ID, PO ID, etc
	UserID        *uuid.UUID `gorm:"type:uuid" json:"user_id,omitempty"`
}

// StockBatch tracks batches/expiry dates.
type StockBatch struct {
	Base
	BranchID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"branch_id"`
	Branch          *Branch    `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	ProductID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"product_id"`
	Product         *Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	BatchNumber     string     `gorm:"size:100;not null" json:"batch_number"`
	Quantity        float64    `gorm:"type:decimal(10,4);not null" json:"quantity"`
	ExpiryDate      *time.Time `json:"expiry_date,omitempty"`
	ManufactureDate *time.Time `json:"manufacture_date,omitempty"`
}

// StocktakeEntry is a line item in a stocktake session.
type StocktakeEntry struct {
	Base
	SessionID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"session_id"`
	ProductID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"product_id"`
	Product       *Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	VariantID     *uuid.UUID `gorm:"type:uuid" json:"variant_id,omitempty"`
	ExpectedStock float64    `gorm:"type:decimal(10,4);not null" json:"expected_stock"`
	ActualStock   float64    `gorm:"type:decimal(10,4);not null" json:"actual_stock"`
	Difference    float64    `gorm:"type:decimal(10,4);not null" json:"difference"` // actual - expected
}

// InventoryRecommendation for AI/Algorithm suggested stock levels.
type InventoryRecommendation struct {
	Base
	BranchID         uuid.UUID `gorm:"type:uuid;not null;index" json:"branch_id"`
	Branch           *Branch   `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	ProductID        uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	RecommendedStock float64   `gorm:"type:decimal(10,4);not null" json:"recommended_stock"`
	Reason           string    `gorm:"type:text" json:"reason"` // E.g., "High sales velocity in past 7 days"
	IsApplied        bool      `gorm:"default:false" json:"is_applied"`
}

// StockAlert alerts users when stock falls below thresholds.
type StockAlert struct {
	Base
	BranchID   uuid.UUID `gorm:"type:uuid;not null;index" json:"branch_id"`
	Branch     *Branch   `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	ProductID  uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	Message    string    `gorm:"type:text" json:"message"`
	IsResolved bool      `gorm:"default:false" json:"is_resolved"`
}

// AfterSave generates a StockAlert if the product's stock is low after a movement.
func (sm *StockMovement) AfterSave(tx *gorm.DB) error {
	var product Product
	if err := tx.Where("id = ?", sm.ProductID).First(&product).Error; err != nil {
		// Ignore if product not found, might be deleted
		return nil
	}

	if product.TrackInventory && sm.NewStock <= product.ReorderLevel {
		branchID := uuid.Nil
		if sm.BranchID != nil {
			branchID = *sm.BranchID
		}
		var count int64
		tx.Model(&StockAlert{}).Where("product_id = ? AND branch_id = ? AND is_resolved = ?", sm.ProductID, branchID, false).Count(&count)
		if count == 0 {
			alert := StockAlert{
				BranchID:  branchID,
				ProductID: sm.ProductID,
				Message:   "Stock level has reached reorder level.",
			}
			tx.Create(&alert)
		}
	}
	return nil
}
