package services_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupBillingTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.Subscription{},
		&models.BillingPayment{},
		&models.Plan{},
	)
	require.NoError(t, err)
	return db
}

func TestBillingService_ListInvoices(t *testing.T) {
	db := setupBillingTestDB(t)
	svc := services.NewBillingService(db)

	tenantID := uuid.New()
	subID := uuid.New()
	planID := uuid.New()

	db.Create(&models.Plan{
		Base: models.Base{ID: planID},
		Name: "Pro Plan",
	})

	tNow := time.Now().AddDate(0, 1, 0)
	db.Create(&models.Subscription{
		Base:             models.Base{ID: subID},
		TenantID:         tenantID,
		PlanID:           &planID,
		Status:           "active",
		CurrentPeriodEnd: &tNow,
	})

	db.Create(&models.BillingPayment{
		Base:           models.Base{ID: uuid.New()},
		SubscriptionID: subID,
		Amount:         29.99,
		Status:         "paid",
	})
	db.Create(&models.BillingPayment{
		Base:           models.Base{ID: uuid.New()},
		SubscriptionID: subID,
		Amount:         29.99,
		Status:         "paid",
	})

	payments, err := svc.ListInvoices(tenantID)
	assert.NoError(t, err)
	assert.Len(t, payments, 2)
}

func TestBillingService_ListInvoices_NoSubscription(t *testing.T) {
	db := setupBillingTestDB(t)
	svc := services.NewBillingService(db)

	tenantID := uuid.New()

	payments, err := svc.ListInvoices(tenantID)
	assert.NoError(t, err)
	assert.Len(t, payments, 0)
}

func TestBillingService_GetSubscription(t *testing.T) {
	db := setupBillingTestDB(t)
	svc := services.NewBillingService(db)

	tenantID := uuid.New()
	subID := uuid.New()
	planID := uuid.New()

	db.Create(&models.Plan{
		Base: models.Base{ID: planID},
		Name: "Pro Plan",
	})

	tNow := time.Now().AddDate(0, 1, 0)
	db.Create(&models.Subscription{
		Base:             models.Base{ID: subID},
		TenantID:         tenantID,
		PlanID:           &planID,
		Status:           "active",
		CurrentPeriodEnd: &tNow,
	})

	sub, err := svc.GetSubscription(tenantID)
	assert.NoError(t, err)
	assert.NotNil(t, sub)
	assert.Equal(t, subID, sub.ID)
	assert.NotNil(t, sub.Plan)
	assert.Equal(t, "Pro Plan", sub.Plan.Name)
}
