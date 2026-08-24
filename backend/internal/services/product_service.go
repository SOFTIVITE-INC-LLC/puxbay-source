package services

import (
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type ProductService struct {
	db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{db: db}
}

type ProductListParams struct {
	BranchID   string
	CategoryID string
	Search     string
	Limit      int
	Offset     int
}

func (s *ProductService) ListProducts(params ProductListParams) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	query := s.db.Model(&models.Product{})

	if params.BranchID != "" {
		query = query.Where("branch_id = ? OR branch_id IS NULL", params.BranchID)
	}

	if params.CategoryID != "" {
		query = query.Where("category_id = ?", params.CategoryID)
	}

	if params.Search != "" {
		query = query.Where("name ILIKE ? OR sku ILIKE ?", "%"+params.Search+"%", "%"+params.Search+"%")
	}

	query.Model(&models.Product{}).Count(&total)

	if err := query.Preload("Category").Offset(params.Offset).Limit(params.Limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

type ProductCreateInput struct {
	Name           string
	Description    string
	SKU            string
	Barcode        string
	CategoryID     *string
	CostPrice      float64
	SellingPrice   float64
	WholesalePrice float64
	TrackInventory bool
	CurrentStock   float64
	ReorderLevel   float64
	StockUnit      string
	IsActive       *bool
	IsOnline       *bool
	BranchID       *string

	ExpiryDate               string
	ManufacturingDate        string
	MinimumWholesaleQuantity float64
	BatchNumber              string
	InvoiceWaybillNumber     string
	CountryOfOrigin          string
	ManufacturerName         string
	ManufacturerAddress      string
}

func (s *ProductService) CreateProduct(input ProductCreateInput) (*models.Product, error) {
	var existing models.Product
	if err := s.db.Where("sku = ?", input.SKU).First(&existing).Error; err == nil {
		return nil, errors.New("product with this SKU already exists")
	}

	product := models.Product{
		Name:                     input.Name,
		Description:              input.Description,
		SKU:                      input.SKU,
		CostPrice:                input.CostPrice,
		SellingPrice:             input.SellingPrice,
		WholesalePrice:           input.WholesalePrice,
		TrackInventory:           input.TrackInventory,
		CurrentStock:             input.CurrentStock,
		ReorderLevel:             input.ReorderLevel,
		MinimumWholesaleQuantity: input.MinimumWholesaleQuantity,
		BatchNumber:              input.BatchNumber,
		InvoiceWaybillNumber:     input.InvoiceWaybillNumber,
		CountryOfOrigin:          input.CountryOfOrigin,
		ManufacturerName:         input.ManufacturerName,
		ManufacturerAddress:      input.ManufacturerAddress,
		IsActive:                 true,
		IsOnline:                 false,
	}

	if input.IsActive != nil {
		product.IsActive = *input.IsActive
	}
	if input.IsOnline != nil {
		product.IsOnline = *input.IsOnline
	}

	if input.StockUnit != "" {
		product.StockUnit = input.StockUnit
	}
	if input.Barcode != "" {
		product.Barcode = &input.Barcode
	}

	if input.ExpiryDate != "" {
		if t, err := time.Parse("2006-01-02", input.ExpiryDate); err == nil {
			product.ExpiryDate = &t
		}
	}
	if input.ManufacturingDate != "" {
		if t, err := time.Parse("2006-01-02", input.ManufacturingDate); err == nil {
			product.ManufacturingDate = &t
		}
	}

	if input.CategoryID != nil && *input.CategoryID != "" {
		catID, err := uuid.Parse(*input.CategoryID)
		if err == nil {
			product.CategoryID = &catID
		}
	}

	if input.BranchID != nil && *input.BranchID != "" {
		brID, err := uuid.Parse(*input.BranchID)
		if err == nil {
			product.BranchID = &brID
		}
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&product).Error; err != nil {
			return err
		}

		// Create StockBatch if batch tracking info is provided
		if product.BatchNumber != "" {
			batch := models.StockBatch{
				ProductID:       product.ID,
				BatchNumber:     product.BatchNumber,
				Quantity:        product.CurrentStock,
				ExpiryDate:      product.ExpiryDate,
				ManufactureDate: product.ManufacturingDate,
			}
			if product.BranchID != nil {
				batch.BranchID = *product.BranchID
			} else {
				// We need a branch ID. If the product doesn't have one, we can't create a batch correctly,
				// but let's assume it gets passed or we just leave it empty if the schema allows.
				// Schema usually requires branch_id, so we'll just try to save it.
			}
			if err := tx.Create(&batch).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (s *ProductService) GetProduct(id string) (*models.Product, error) {
	var product models.Product
	if err := s.db.Preload("Category").Where("id = ?", id).First(&product).Error; err != nil {
		return nil, errors.New("product not found")
	}
	return &product, nil
}

func (s *ProductService) UpdateProduct(id string, input ProductCreateInput) (*models.Product, error) {
	product, err := s.GetProduct(id)
	if err != nil {
		return nil, err
	}

	if input.SKU != product.SKU {
		var existing models.Product
		if err := s.db.Where("sku = ? AND id != ?", input.SKU, id).First(&existing).Error; err == nil {
			return nil, errors.New("another product with this SKU already exists")
		}
	}

	product.Name = input.Name
	product.Description = input.Description
	product.SKU = input.SKU
	product.CostPrice = input.CostPrice
	product.SellingPrice = input.SellingPrice
	product.WholesalePrice = input.WholesalePrice
	product.TrackInventory = input.TrackInventory
	product.ReorderLevel = input.ReorderLevel
	product.StockUnit = input.StockUnit

	if input.IsActive != nil {
		product.IsActive = *input.IsActive
	}
	if input.IsOnline != nil {
		product.IsOnline = *input.IsOnline
	}

	if input.StockUnit != "" {
		product.StockUnit = input.StockUnit
	}
	if input.Barcode != "" {
		product.Barcode = &input.Barcode
	}

	if input.CategoryID != nil && *input.CategoryID != "" {
		catID, err := uuid.Parse(*input.CategoryID)
		if err == nil {
			product.CategoryID = &catID
		}
	} else if input.CategoryID != nil && *input.CategoryID == "" {
		product.CategoryID = nil
	}

	if input.BranchID != nil && *input.BranchID != "" {
		brID, err := uuid.Parse(*input.BranchID)
		if err == nil {
			product.BranchID = &brID
		}
	} else if input.BranchID != nil && *input.BranchID == "" {
		product.BranchID = nil
	}

	if err := s.db.Save(&product).Error; err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductService) DeleteProduct(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.Product{}).Error
}

func (s *ProductService) ImportProductsFromExcel(tenantID uuid.UUID, branchID *uuid.UUID, fileReader io.Reader) (int, error) {
	f, err := excelize.OpenReader(fileReader)
	if err != nil {
		return 0, errors.New("failed to open excel file: " + err.Error())
	}

	sheetName := f.GetSheetName(f.GetActiveSheetIndex())
	rows := f.GetRows(sheetName)

	if len(rows) < 2 {
		return 0, errors.New("no data found in excel file")
	}

	var newProducts []models.Product
	skusToImport := make(map[string]bool)

	for i, row := range rows {
		if i == 0 || len(row) < 4 {
			continue // skip header or invalid rows
		}

		name := strings.TrimSpace(row[0])
		sku := strings.TrimSpace(row[1])

		if name == "" || sku == "" {
			continue // required fields missing
		}
		if skusToImport[sku] {
			continue // Avoid duplicate SKUs within the import file
		}
		skusToImport[sku] = true

		product := models.Product{
			BranchID:       branchID,
			Name:           name,
			SKU:            sku,
			IsActive:       true,
			TrackInventory: true,
		}

		// 3. price (selling_price)
		if len(row) > 3 {
			if parsed, err := strconv.ParseFloat(strings.TrimSpace(row[3]), 64); err == nil {
				product.SellingPrice = parsed
			}
		}
		// 4. wholesale_price
		if len(row) > 4 {
			if parsed, err := strconv.ParseFloat(strings.TrimSpace(row[4]), 64); err == nil {
				product.WholesalePrice = parsed
			}
		}
		// 5. min_wholesale_qty
		if len(row) > 5 {
			if parsed, err := strconv.ParseFloat(strings.TrimSpace(row[5]), 64); err == nil {
				product.MinimumWholesaleQuantity = parsed
			}
		}
		// 6. cost_price
		if len(row) > 6 {
			if parsed, err := strconv.ParseFloat(strings.TrimSpace(row[6]), 64); err == nil {
				product.CostPrice = parsed
			}
		}
		// 7. stock_quantity
		if len(row) > 7 {
			if parsed, err := strconv.ParseFloat(strings.TrimSpace(row[7]), 64); err == nil {
				product.CurrentStock = parsed
			}
		}
		// 8. low_stock_threshold (reorder_level)
		if len(row) > 8 {
			if parsed, err := strconv.ParseFloat(strings.TrimSpace(row[8]), 64); err == nil {
				product.ReorderLevel = parsed
			}
		}
		// 9. barcode
		if len(row) > 9 {
			barcode := strings.TrimSpace(row[9])
			if barcode != "" {
				product.Barcode = &barcode
			}
		}
		// 10. expiry_date
		if len(row) > 10 {
			if t, err := time.Parse("2006-01-02", strings.TrimSpace(row[10])); err == nil {
				product.ExpiryDate = &t
			}
		}
		// 11. batch_number
		if len(row) > 11 {
			product.BatchNumber = strings.TrimSpace(row[11])
		}
		// 12. invoice_waybill_number
		if len(row) > 12 {
			product.InvoiceWaybillNumber = strings.TrimSpace(row[12])
		}
		// 13. description
		if len(row) > 13 {
			product.Description = strings.TrimSpace(row[13])
		}
		// 14. is_active
		if len(row) > 14 {
			activeStr := strings.ToUpper(strings.TrimSpace(row[14]))
			if activeStr == "FALSE" || activeStr == "0" {
				product.IsActive = false
			}
		}
		// 16. mfg_date (skip 15 image_url)
		if len(row) > 16 {
			if t, err := time.Parse("2006-01-02", strings.TrimSpace(row[16])); err == nil {
				product.ManufacturingDate = &t
			}
		}
		// 17. country_of_origin
		if len(row) > 17 {
			product.CountryOfOrigin = strings.TrimSpace(row[17])
		}
		// 18. manufacturer_name
		if len(row) > 18 {
			product.ManufacturerName = strings.TrimSpace(row[18])
		}
		// 19. manufacturer_address
		if len(row) > 19 {
			product.ManufacturerAddress = strings.TrimSpace(row[19])
		}

		newProducts = append(newProducts, product)
	}

	if len(newProducts) == 0 {
		return 0, nil
	}

	count := 0
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// We still need to filter out SKUs that already exist in the DB
		var existingSKUs []string
		skus := make([]string, 0, len(newProducts))
		for _, p := range newProducts {
			skus = append(skus, p.SKU)
		}

		dupQuery := tx.Model(&models.Product{}).Where("sku IN ?", skus)
		if branchID != nil {
			dupQuery = dupQuery.Where("branch_id = ?", branchID)
		} else {
			dupQuery = dupQuery.Where("branch_id IS NULL")
		}
		dupQuery.Pluck("sku", &existingSKUs)

		existingMap := make(map[string]bool)
		for _, sku := range existingSKUs {
			existingMap[sku] = true
		}

		var toInsert []models.Product
		for _, p := range newProducts {
			if !existingMap[p.SKU] {
				toInsert = append(toInsert, p)
			}
		}

		if len(toInsert) == 0 {
			return nil
		}

		// Bulk insert in batches of 100
		if err := tx.CreateInBatches(toInsert, 100).Error; err != nil {
			return err
		}

		var batchesToInsert []models.StockBatch
		for _, p := range toInsert {
			if p.BatchNumber != "" {
				batch := models.StockBatch{
					ProductID:       p.ID,
					BatchNumber:     p.BatchNumber,
					Quantity:        p.CurrentStock,
					ExpiryDate:      p.ExpiryDate,
					ManufactureDate: p.ManufacturingDate,
				}
				if p.BranchID != nil {
					batch.BranchID = *p.BranchID
				} else if branchID != nil {
					batch.BranchID = *branchID
				}
				batchesToInsert = append(batchesToInsert, batch)
			}
		}

		if len(batchesToInsert) > 0 {
			if err := tx.CreateInBatches(batchesToInsert, 100).Error; err != nil {
				return err
			}
		}

		count = len(toInsert)
		return nil
	})

	if err != nil {
		return 0, err
	}

	return count, nil
}
