package services

import (
	"errors"
	"strings"

	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type TaxEngine interface {
	CalculateTax(subtotal float64, region string, taxExempt bool) (float64, error)
	CalculateItemTax(item models.OrderItem, region string, taxExempt bool) (float64, error)
}

type DefaultTaxEngine struct {
	db *gorm.DB
}

func NewDefaultTaxEngine(db *gorm.DB) *DefaultTaxEngine {
	return &DefaultTaxEngine{db: db}
}

// CalculateTax applies a standard region-based tax if the customer isn't tax exempt.
func (e *DefaultTaxEngine) CalculateTax(subtotal float64, region string, taxExempt bool) (float64, error) {
	if taxExempt || subtotal <= 0 {
		return 0, nil
	}

	var config models.TaxConfiguration
	if err := e.db.Where("LOWER(region) = ?", strings.ToLower(region)).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// fallback to a default tax rate of 0 if region is not configured
			return 0, nil
		}
		return 0, err
	}

	taxAmount := subtotal * (config.TaxRate / 100.0)
	return taxAmount, nil
}

// CalculateItemTax applies item-specific tax logic based on region.
func (e *DefaultTaxEngine) CalculateItemTax(item models.OrderItem, region string, taxExempt bool) (float64, error) {
	if taxExempt || item.Total <= 0 {
		return 0, nil
	}

	// This can be extended to use Product Tax Classes (e.g. standard, reduced, zero)
	return e.CalculateTax(item.Total, region, taxExempt)
}
