package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"github.com/softivite/puxbay/internal/utils"
	"gorm.io/gorm"
)

type ProductHandler struct {
	db *gorm.DB
}

func NewProductHandler(db *gorm.DB) *ProductHandler {
	return &ProductHandler{db: db}
}

func (h *ProductHandler) service(c *gin.Context) *services.ProductService {
	return services.NewProductService(getDB(c, h.db))
}

func (h *ProductHandler) List(c *gin.Context) {
	p := utils.GetPagination(c)

	params := services.ProductListParams{
		BranchID:   middleware.ResolveBranchID(c, c.Query("branch_id")),
		CategoryID: c.Query("category_id"),
		Search:     c.Query("q"),
		Limit:      p.Limit,
		Offset:     p.Offset,
	}

	products, total, err := h.service(c).ListProducts(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	role, _ := c.Get(middleware.ContextKeyRole)
	permissions := c.GetStringSlice(middleware.ContextKeyPermissions)
	maskedProducts := utils.MaskCollection(products, role.(string), permissions)

	c.JSON(http.StatusOK, gin.H{
		"data":  maskedProducts,
		"total": total,
		"page":  p.Page,
		"limit": p.Limit,
	})
}

type ProductCreateRequest struct {
	Name           string  `json:"name" binding:"required"`
	Description    string  `json:"description"`
	SKU            string  `json:"sku" binding:"required"`
	Barcode        string  `json:"barcode"`
	Image          *string `json:"image"`
	ImageURL       *string `json:"image_url"`
	CategoryID     *string `json:"category_id"`
	CostPrice      float64 `json:"cost_price"`
	SellingPrice   float64 `json:"selling_price" binding:"required"`
	WholesalePrice float64 `json:"wholesale_price"`
	TrackInventory bool    `json:"track_inventory"`
	CurrentStock   float64 `json:"current_stock"`
	ReorderLevel   float64 `json:"reorder_level"`
	StockUnit      string  `json:"stock_unit"`
	IsActive       *bool   `json:"is_active"`
	IsOnline       *bool   `json:"is_online"`

	ExpiryDate               string  `json:"expiry_date"`
	ManufacturingDate        string  `json:"manufacturing_date"`
	MinimumWholesaleQuantity float64 `json:"minimum_wholesale_quantity"`
	BatchNumber              string  `json:"batch_number"`
	InvoiceWaybillNumber     string  `json:"invoice_waybill_number"`
	CountryOfOrigin          string  `json:"country_of_origin"`
	ManufacturerName         string  `json:"manufacturer_name"`
	ManufacturerAddress      string  `json:"manufacturer_address"`
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req ProductCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	var branchID *string
	if ctxBranchID, ok := middleware.GetBranchID(c); ok {
		strID := ctxBranchID.String()
		branchID = &strID
	}

	img := req.Image
	if img == nil || *img == "" {
		img = req.ImageURL
	}

	input := services.ProductCreateInput{
		Name:                     req.Name,
		Description:              req.Description,
		SKU:                      req.SKU,
		Barcode:                  req.Barcode,
		Image:                    img,
		ImageURL:                 img,
		CategoryID:               req.CategoryID,
		CostPrice:                req.CostPrice,
		SellingPrice:             req.SellingPrice,
		WholesalePrice:           req.WholesalePrice,
		TrackInventory:           req.TrackInventory,
		CurrentStock:             req.CurrentStock,
		ReorderLevel:             req.ReorderLevel,
		StockUnit:                req.StockUnit,
		IsActive:                 req.IsActive,
		IsOnline:                 req.IsOnline,
		BranchID:                 branchID,
		ExpiryDate:               req.ExpiryDate,
		ManufacturingDate:        req.ManufacturingDate,
		MinimumWholesaleQuantity: req.MinimumWholesaleQuantity,
		BatchNumber:              req.BatchNumber,
		InvoiceWaybillNumber:     req.InvoiceWaybillNumber,
		CountryOfOrigin:          req.CountryOfOrigin,
		ManufacturerName:         req.ManufacturerName,
		ManufacturerAddress:      req.ManufacturerAddress,
	}

	product, err := h.service(c).CreateProduct(input)
	if err != nil {
		if err.Error() == "product with this SKU already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, product)
}

func (h *ProductHandler) Get(c *gin.Context) {
	id := c.Param("id")

	product, err := h.service(c).GetProduct(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	role, _ := c.Get(middleware.ContextKeyRole)
	permissions := c.GetStringSlice(middleware.ContextKeyPermissions)
	maskedProduct := utils.MaskCollection(product, role.(string), permissions)

	c.JSON(http.StatusOK, maskedProduct)
}

func (h *ProductHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req ProductCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}
	var branchID *string
	if ctxBranchID, ok := middleware.GetBranchID(c); ok {
		strID := ctxBranchID.String()
		branchID = &strID
	}

	img := req.Image
	if img == nil || *img == "" {
		img = req.ImageURL
	}

	input := services.ProductCreateInput{
		Name:                     req.Name,
		Description:              req.Description,
		SKU:                      req.SKU,
		Barcode:                  req.Barcode,
		Image:                    img,
		ImageURL:                 img,
		CategoryID:               req.CategoryID,
		CostPrice:                req.CostPrice,
		SellingPrice:             req.SellingPrice,
		WholesalePrice:           req.WholesalePrice,
		TrackInventory:           req.TrackInventory,
		CurrentStock:             req.CurrentStock,
		ReorderLevel:             req.ReorderLevel,
		StockUnit:                req.StockUnit,
		IsActive:                 req.IsActive,
		IsOnline:                 req.IsOnline,
		BranchID:                 branchID,
		ExpiryDate:               req.ExpiryDate,
		ManufacturingDate:        req.ManufacturingDate,
		MinimumWholesaleQuantity: req.MinimumWholesaleQuantity,
		BatchNumber:              req.BatchNumber,
		InvoiceWaybillNumber:     req.InvoiceWaybillNumber,
		CountryOfOrigin:          req.CountryOfOrigin,
		ManufacturerName:         req.ManufacturerName,
		ManufacturerAddress:      req.ManufacturerAddress,
	}

	product, err := h.service(c).UpdateProduct(id, input)
	if err != nil {
		if err.Error() == "another product with this SKU already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service(c).DeleteProduct(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *ProductHandler) ImportExcel(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer f.Close()

	rawTenantID, _ := c.Get(middleware.ContextKeyTenantID)
	tenantID, ok := rawTenantID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid tenant ID"})
		return
	}

	bIDStr := middleware.ResolveBranchID(c, "")
	var branchID *uuid.UUID
	if bIDStr != "" {
		if b, err := uuid.Parse(bIDStr); err == nil {
			branchID = &b
		}
	}

	count, err := h.service(c).ImportProductsFromExcel(tenantID, branchID, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import products: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Products imported successfully", "count": count})
}

// ─── Gallery Images ────────────────────────────────────────────────────────

const maxGalleryImages = 5

// GetImages returns the gallery images for a product.
func (h *ProductHandler) GetImages(c *gin.Context) {
	id := c.Param("id")
	productID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}
	db := getDB(c, h.db)
	var images []models.ProductImageGallery
	if err := db.Where("product_id = ?", productID).Order("\"order\" asc, created_at asc").Find(&images).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch images"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": images})
}

// AddImage uploads a new gallery image for a product (max 5, max 2 MB each).
func (h *ProductHandler) AddImage(c *gin.Context) {
	id := c.Param("id")
	productID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}
	db := getDB(c, h.db)

	// Enforce gallery limit
	var count int64
	db.Model(&models.ProductImageGallery{}).Where("product_id = ?", productID).Count(&count)
	if count >= int64(maxGalleryImages) {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Maximum of %d gallery images allowed per product", maxGalleryImages)})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded (field 'file' required)"})
		return
	}

	// Enforce 2 MB limit
	if file.Size > maxUploadSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": fmt.Sprintf("File too large (%.2f MB). Maximum is 2 MB", float64(file.Size)/1024/1024),
		})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" && ext != ".svg" && ext != ".gif" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid format. Allowed: JPG, PNG, WEBP, SVG, GIF"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}
	defer src.Close()

	folder := "uploads/products"
	_ = os.MkdirAll(folder, os.ModePerm)
	filename := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), uuid.New().String()[:8], ext)
	dstPath := filepath.Join(folder, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write file"})
		return
	}

	imageURL := "/" + filepath.ToSlash(dstPath)
	gallery := models.ProductImageGallery{
		ProductID: productID,
		ImageURL:  imageURL,
		Order:     uint(count),
	}
	if err := db.Create(&gallery).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image record"})
		return
	}

	c.JSON(http.StatusCreated, gallery)
}

// DeleteImage removes a gallery image from a product.
func (h *ProductHandler) DeleteImage(c *gin.Context) {
	id := c.Param("id")
	imgID := c.Param("imgId")

	productID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}
	imageID, err := uuid.Parse(imgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image ID"})
		return
	}

	db := getDB(c, h.db)
	var img models.ProductImageGallery
	if err := db.Where("id = ? AND product_id = ?", imageID, productID).First(&img).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	// Remove file from disk
	if img.ImageURL != "" {
		diskPath := strings.TrimPrefix(img.ImageURL, "/")
		_ = os.Remove(diskPath)
	}

	if err := db.Delete(&img).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
