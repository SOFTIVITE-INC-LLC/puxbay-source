package services

import (
	"bytes"
	"errors"
	"fmt"
	"image/png"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/qr"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type BarcodeService struct {
	db *gorm.DB
}

func NewBarcodeService(db *gorm.DB) *BarcodeService {
	return &BarcodeService{db: db}
}

func (s *BarcodeService) GenerateProductBarcode(productID string) ([]byte, string, error) {
	var product models.Product
	if err := s.db.Where("id = ?", productID).First(&product).Error; err != nil {
		return nil, "", errors.New("product not found")
	}

	barcodeData := product.SKU
	if product.Barcode != nil && *product.Barcode != "" {
		barcodeData = *product.Barcode
	}

	if barcodeData == "" {
		return nil, "", errors.New("product has no barcode or SKU")
	}

	bcode, err := code128.Encode(barcodeData)
	if err != nil {
		return nil, "", errors.New("failed to generate barcode")
	}

	scaled, err := barcode.Scale(bcode, 300, 50)
	if err != nil {
		return nil, "", errors.New("failed to scale barcode")
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, scaled); err != nil {
		return nil, "", errors.New("failed to encode barcode to png")
	}

	return buf.Bytes(), barcodeData, nil
}

func (s *BarcodeService) GenerateProductQR(productID string) ([]byte, error) {
	var product models.Product
	if err := s.db.Where("id = ?", productID).First(&product).Error; err != nil {
		return nil, errors.New("product not found")
	}

	qrData := fmt.Sprintf("https://puxbay.com/p/%s", product.ID)

	qrCode, err := qr.Encode(qrData, qr.M, qr.Auto)
	if err != nil {
		return nil, errors.New("failed to generate QR code")
	}

	scaled, err := barcode.Scale(qrCode, 200, 200)
	if err != nil {
		return nil, errors.New("failed to scale QR code")
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, scaled); err != nil {
		return nil, errors.New("failed to encode QR code to png")
	}

	return buf.Bytes(), nil
}

func (s *BarcodeService) BulkGenerateBarcodes(productIDs []string) (map[string][]byte, error) {
	results := make(map[string][]byte)
	for _, id := range productIDs {
		imgBytes, _, err := s.GenerateProductBarcode(id)
		if err == nil {
			results[id] = imgBytes
		}
	}
	return results, nil
}
