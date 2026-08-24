package handlers

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/softivite/puxbay/internal/services"
	"gorm.io/gorm"
)

type BarcodeHandler struct {
	db *gorm.DB
}

func NewBarcodeHandler(db *gorm.DB) *BarcodeHandler {
	return &BarcodeHandler{}
}

func (h *BarcodeHandler) service(c *gin.Context) *services.BarcodeService {
	return services.NewBarcodeService(getDB(c, h.db))
}

// GenerateProductBarcode generates a Code128 barcode image for a given product ID.
func (h *BarcodeHandler) GenerateProductBarcode(c *gin.Context) {
	productID := c.Param("id")

	imageBytes, barcodeData, err := h.service(c).GenerateProductBarcode(productID)
	if err != nil {
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "product has no barcode or SKU" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	format := c.Query("format")
	if format == "base64" {
		encoded := base64.StdEncoding.EncodeToString(imageBytes)
		c.JSON(http.StatusOK, gin.H{"image": "data:image/png;base64," + encoded, "data": barcodeData})
		return
	}

	c.Data(http.StatusOK, "image/png", imageBytes)
}

// GenerateProductQR generates a QR code image for a given product ID.
func (h *BarcodeHandler) GenerateProductQR(c *gin.Context) {
	productID := c.Param("id")

	imageBytes, err := h.service(c).GenerateProductQR(productID)
	if err != nil {
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	format := c.Query("format")
	if format == "base64" {
		encoded := base64.StdEncoding.EncodeToString(imageBytes)
		c.JSON(http.StatusOK, gin.H{"image": "data:image/png;base64," + encoded})
		return
	}

	c.Data(http.StatusOK, "image/png", imageBytes)
}

type BulkGenerateRequest struct {
	ProductIDs []string `json:"product_ids" binding:"required"`
}

func (h *BarcodeHandler) BulkGenerateBarcodes(c *gin.Context) {
	var req BulkGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results, err := h.service(c).BulkGenerateBarcodes(req.ProductIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate barcodes"})
		return
	}

	// Convert to base64
	base64Results := make(map[string]string)
	for id, imgBytes := range results {
		encoded := base64.StdEncoding.EncodeToString(imgBytes)
		base64Results[id] = "data:image/png;base64," + encoded
	}

	c.JSON(http.StatusOK, gin.H{"images": base64Results})
}
