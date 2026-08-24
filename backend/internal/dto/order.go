package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
)

// OrderListResponse represents a summarized order for list views
// to prevent N+1 queries from loading all nested items and products.
type OrderListResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	OrderNumber   string  `json:"order_number"`
	Total         float64 `json:"total"`
	AmountPaid    float64 `json:"amount_paid"`
	Status        string  `json:"status"`
	PaymentStatus string  `json:"payment_status"`
	PaymentMethod string  `json:"payment_method"`
	OrderType     string  `json:"order_type"`
	ReceiptToken  string  `json:"receipt_token"`

	// Relationships summarized
	CustomerID   *uuid.UUID       `json:"customer_id,omitempty"`
	CustomerName string           `json:"customer_name,omitempty"`
	CashierID    *uuid.UUID       `json:"cashier_id,omitempty"`
	ItemCount    int              `json:"item_count"`
	Customer     *models.Customer `json:"customer,omitempty"` // Full customer if needed by frontend
}
