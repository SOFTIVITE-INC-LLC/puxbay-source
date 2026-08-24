package services_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.Customer{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
		&models.KDSTicket{},
	)
	require.NoError(t, err)
	return db
}

func TestOrderService_VoidOrder(t *testing.T) {
	db := setupTestDB(t)
	svc := services.NewOrderService(db, nil)

	customerID := uuid.New()
	db.Create(&models.Customer{
		TenantScoped: models.TenantScoped{Base: models.Base{ID: customerID}},
		Name:         "Test Customer",
		TotalSpend:   100.0,
		OrderCount:   1,
		LoyaltyPts:   10.0,
	})

	productID := uuid.New()
	db.Create(&models.Product{
		TenantScoped:   models.TenantScoped{Base: models.Base{ID: productID}},
		Name:           "Test Product",
		TrackInventory: true,
		CurrentStock:   5,
		SellingPrice:   100.0,
	})

	orderID := uuid.New()
	db.Create(&models.Order{
		BranchScoped: models.BranchScoped{TenantScoped: models.TenantScoped{Base: models.Base{ID: orderID}}},
		CustomerID:   &customerID,
		Total:        100.0,
		Status:       "completed",
		Items: []models.OrderItem{
			{ProductID: productID, Quantity: 2},
		},
	})

	// Perform Void
	err := svc.VoidOrder(orderID.String())
	assert.NoError(t, err)

	// Verify order status
	var order models.Order
	db.First(&order, "id = ?", orderID)
	assert.Equal(t, "voided", order.Status)

	// Verify inventory reversed
	var prod models.Product
	db.First(&prod, "id = ?", productID)
	assert.Equal(t, 7.0, prod.CurrentStock)

	// Verify loyalty reversed
	var cust models.Customer
	db.First(&cust, "id = ?", customerID)
	assert.Equal(t, 0.0, cust.TotalSpend)
	assert.Equal(t, 0, cust.OrderCount)
	assert.Equal(t, 0.0, cust.LoyaltyPts)
}

func TestOrderService_CreateOrder_NegativeStockPrevention(t *testing.T) {
	db := setupTestDB(t)
	svc := services.NewOrderService(db, nil)

	productID := uuid.New()
	db.Create(&models.Product{
		TenantScoped:   models.TenantScoped{Base: models.Base{ID: productID}},
		Name:           "Test Product",
		TrackInventory: true,
		CurrentStock:   5, // Only 5 in stock
		SellingPrice:   100.0,
	})

	input := services.OrderCreateInput{
		Total: 600.0,
		Items: []services.OrderItemInput{
			{ProductID: productID, Quantity: 6}, // Trying to buy 6
		},
	}

	_, err := svc.CreateOrder(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient stock")

	// Verify stock didn't change
	var prod models.Product
	db.First(&prod, "id = ?", productID)
	assert.Equal(t, 5.0, prod.CurrentStock)
}
