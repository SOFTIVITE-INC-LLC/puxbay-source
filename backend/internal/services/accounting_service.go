package services

import (
	"log"

	"github.com/softivite/puxbay/internal/models"
)

type AccountingIntegrationService interface {
	SyncInvoice(order *models.Order) error
	SyncCustomer(customer *models.Customer) error
	SyncProduct(product *models.Product) error
}

// XeroIntegrationStub is a stub for pushing data to Xero.
type XeroIntegrationStub struct {
	Enabled bool
}

func NewXeroIntegrationStub() *XeroIntegrationStub {
	return &XeroIntegrationStub{Enabled: true}
}

func (x *XeroIntegrationStub) SyncInvoice(order *models.Order) error {
	if !x.Enabled {
		return nil
	}
	log.Printf("[XERO-STUB] Syncing Invoice for Order %s. Total: %.2f\n", order.OrderNumber, order.Total)
	return nil
}

func (x *XeroIntegrationStub) SyncCustomer(customer *models.Customer) error {
	if !x.Enabled {
		return nil
	}
	log.Printf("[XERO-STUB] Syncing Customer %s to Xero Contacts.\n", customer.Name)
	return nil
}

func (x *XeroIntegrationStub) SyncProduct(product *models.Product) error {
	if !x.Enabled {
		return nil
	}
	log.Printf("[XERO-STUB] Syncing Product %s (SKU: %s) to Xero Items.\n", product.Name, product.SKU)
	return nil
}
