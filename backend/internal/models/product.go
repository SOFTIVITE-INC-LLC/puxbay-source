package models

import (
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func prependCDN(image *string) {
	if image != nil && *image != "" && !strings.HasPrefix(*image, "http") {
		cdnURL := os.Getenv("CDN_URL")
		if cdnURL != "" {
			fullURL := strings.TrimRight(cdnURL, "/") + "/" + strings.TrimLeft(*image, "/")
			*image = fullURL
		}
	}
}

// Category represents a product category.
type Category struct {
	TenantScoped
	Name        string  `gorm:"size:100;not null" json:"name"`
	Description string  `gorm:"type:text" json:"description,omitempty"`
	Image       *string `gorm:"size:255" json:"image,omitempty"`
	Color       string  `gorm:"size:20;default:'blue'" json:"color"`
}

func (c *Category) AfterFind(tx *gorm.DB) (err error) {
	prependCDN(c.Image)
	return
}

// Product represents an item for sale or use.
type Product struct {
	TenantScoped
	BranchID    *uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_tenant_branch_sku;index" json:"branch_id,omitempty"` // Nullable for global products
	Branch      *Branch    `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Name        string     `gorm:"size:200;not null" json:"name"`
	Description string     `gorm:"type:text" json:"description,omitempty"`
	SKU         string     `gorm:"size:100;uniqueIndex:idx_tenant_branch_sku;not null" json:"sku"`
	Barcode     *string    `gorm:"size:100;index" json:"barcode,omitempty"`
	CategoryID  *uuid.UUID `gorm:"type:uuid;index" json:"category_id,omitempty"`
	Category    *Category  `gorm:"foreignKey:CategoryID" json:"category,omitempty"`

	CostPrice      float64 `gorm:"type:decimal(12,2);default:0" json:"cost_price"`
	SellingPrice   float64 `gorm:"type:decimal(12,2);not null" json:"selling_price"`
	WholesalePrice float64 `gorm:"type:decimal(12,2);default:0" json:"wholesale_price"`

	TrackInventory bool    `gorm:"default:true" json:"track_inventory"`
	CurrentStock   float64 `gorm:"type:decimal(10,4);default:0" json:"current_stock"`
	ReorderLevel   float64 `gorm:"type:decimal(10,4);default:0" json:"reorder_level"`
	StockUnit      string  `gorm:"size:50;default:'pcs'" json:"stock_unit"`

	HasVariants bool    `gorm:"default:false;index" json:"has_variants"`
	IsComposite bool    `gorm:"default:false" json:"is_composite"`
	IsActive    bool    `gorm:"default:true;index" json:"is_active"`
	IsOnline    bool    `gorm:"default:true;index" json:"is_online"` // Sell online
	Version     int     `gorm:"default:1" json:"version"`            // Optimistic Locking
	Image       *string `gorm:"size:255" json:"image,omitempty"`

	SupplierID *uuid.UUID `gorm:"type:uuid;index" json:"supplier_id,omitempty"`
	Supplier   *Supplier  `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`

	LastReceivedDate  *time.Time `json:"last_received_date,omitempty"`
	ExpiryDate        *time.Time `json:"expiry_date,omitempty"`
	ManufacturingDate *time.Time `json:"manufacturing_date,omitempty"`

	MinimumWholesaleQuantity float64 `gorm:"type:decimal(10,4);default:1" json:"minimum_wholesale_quantity"`
	BatchNumber              string  `gorm:"size:100" json:"batch_number,omitempty"`
	InvoiceWaybillNumber     string  `gorm:"size:100" json:"invoice_waybill_number,omitempty"`
	CountryOfOrigin          string  `gorm:"size:100" json:"country_of_origin,omitempty"`
	ManufacturerName         string  `gorm:"size:200" json:"manufacturer_name,omitempty"`
	ManufacturerAddress      string  `gorm:"type:text" json:"manufacturer_address,omitempty"`

	// Relations
}

func (p *Product) AfterFind(tx *gorm.DB) (err error) {
	prependCDN(p.Image)
	return
}

// ProductVariant represents a variation of a product (e.g. size/color).
type ProductVariant struct {
	Base
	ProductID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"product_id"`
	Name          string         `gorm:"size:100;not null" json:"name"`
	SKU           string         `gorm:"size:100;uniqueIndex;not null" json:"sku"`
	Barcode       *string        `gorm:"size:100;index" json:"barcode,omitempty"`
	PriceOverride *float64       `gorm:"type:decimal(12,2)" json:"price_override,omitempty"`
	CostOverride  *float64       `gorm:"type:decimal(12,2)" json:"cost_override,omitempty"`
	CurrentStock  float64        `gorm:"type:decimal(10,4);default:0" json:"current_stock"`
	Attributes    datatypes.JSON `gorm:"type:jsonb;check:jsonb_typeof(attributes) = 'object'" json:"attributes"` // {"size": "L", "color": "Red"}
	Image         *string        `gorm:"size:255" json:"image,omitempty"`
}

func (pv *ProductVariant) AfterFind(tx *gorm.DB) (err error) {
	prependCDN(pv.Image)
	return
}

// ProductComponent represents an ingredient/part of a composite product.
type ProductComponent struct {
	Base
	CompositeProductID uuid.UUID `gorm:"type:uuid;not null;index" json:"composite_product_id"`
	ComponentProductID uuid.UUID `gorm:"type:uuid;not null;index" json:"component_product_id"`
	Quantity           float64   `gorm:"type:decimal(10,4);not null" json:"quantity"`
}

// ProductHistory tracks price/cost changes.
type ProductHistory struct {
	Base
	ProductID uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	UserID    uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	Field     string    `gorm:"size:50" json:"field"`
	OldValue  string    `gorm:"type:text" json:"old_value"`
	NewValue  string    `gorm:"type:text" json:"new_value"`
}
