package services

import (
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type ExportService struct {
	db *gorm.DB
}

func NewExportService(db *gorm.DB) *ExportService {
	return &ExportService{db: db}
}

func (s *ExportService) ExportOrdersCSV(writer io.Writer, branchID, startDateStr, endDateStr string) error {
	query := s.db.Model(&models.Order{})
	if branchID != "" {
		query = query.Where("branch_id = ?", branchID)
	}
	if startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			query = query.Where("created_at <= ?", t.Add(24*time.Hour))
		}
	}

	var orders []models.Order
	if err := query.Find(&orders).Error; err != nil {
		return err
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	csvWriter.Write([]string{"Order ID", "Order Number", "Date", "Status", "Subtotal", "Tax", "Discount", "Total", "Payment Method"})

	for _, order := range orders {
		csvWriter.Write([]string{
			order.ID.String(),
			order.OrderNumber,
			order.CreatedAt.Format(time.RFC3339),
			order.Status,
			fmt.Sprintf("%.2f", order.Subtotal),
			fmt.Sprintf("%.2f", order.Tax),
			fmt.Sprintf("%.2f", order.Discount),
			fmt.Sprintf("%.2f", order.Total),
			order.PaymentMethod,
		})
	}
	return nil
}

func (s *ExportService) ExportProductsCSV(writer io.Writer, branchID string) error {
	query := s.db.Model(&models.Product{})
	if branchID != "" {
		query = query.Where("branch_id = ?", branchID)
	}

	var products []models.Product
	if err := query.Find(&products).Error; err != nil {
		return err
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	csvWriter.Write([]string{"Product ID", "Name", "SKU", "Barcode", "Selling Price", "Cost Price", "Stock", "Is Active", "Is Online"})

	for _, product := range products {
		barcode := ""
		if product.Barcode != nil {
			barcode = *product.Barcode
		}
		csvWriter.Write([]string{
			product.ID.String(),
			product.Name,
			product.SKU,
			barcode,
			fmt.Sprintf("%.2f", product.SellingPrice),
			fmt.Sprintf("%.2f", product.CostPrice),
			fmt.Sprintf("%.2f", product.CurrentStock),
			fmt.Sprintf("%t", product.IsActive),
			fmt.Sprintf("%t", product.IsOnline),
		})
	}
	return nil
}

func (s *ExportService) ExportInventoryCSV(writer io.Writer, branchID string) error {
	query := s.db.Model(&models.Product{}).Where("track_inventory = ?", true)
	if branchID != "" {
		query = query.Where("branch_id = ?", branchID)
	}

	var products []models.Product
	if err := query.Order("name ASC").Find(&products).Error; err != nil {
		return err
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	csvWriter.Write([]string{"Product ID", "Name", "SKU", "Current Stock", "Reorder Level", "Stock Unit", "Cost Price", "Selling Price"})

	for _, product := range products {
		csvWriter.Write([]string{
			product.ID.String(),
			product.Name,
			product.SKU,
			fmt.Sprintf("%.2f", product.CurrentStock),
			fmt.Sprintf("%.2f", product.ReorderLevel),
			product.StockUnit,
			fmt.Sprintf("%.2f", product.CostPrice),
			fmt.Sprintf("%.2f", product.SellingPrice),
		})
	}
	return nil
}

func (s *ExportService) ExportCustomersCSV(writer io.Writer) error {
	var customers []models.Customer
	if err := s.db.Order("name ASC").Find(&customers).Error; err != nil {
		return err
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	csvWriter.Write([]string{"Customer ID", "Name", "Email", "Phone", "Total Spend", "Order Count", "Loyalty Points", "Debt Balance"})

	for _, customer := range customers {
		email := ""
		if customer.Email != nil {
			email = *customer.Email
		}
		phone := ""
		if customer.Phone != nil {
			phone = *customer.Phone
		}
		csvWriter.Write([]string{
			customer.ID.String(),
			customer.Name,
			email,
			phone,
			fmt.Sprintf("%.2f", customer.TotalSpend),
			fmt.Sprintf("%d", customer.OrderCount),
			fmt.Sprintf("%.2f", customer.LoyaltyPts),
			fmt.Sprintf("%.2f", customer.DebtBalance),
		})
	}
	return nil
}
