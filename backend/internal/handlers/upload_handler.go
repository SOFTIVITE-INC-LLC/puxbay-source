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
)

const maxUploadSize = 2 * 1024 * 1024 // 2 MB

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	_ = os.MkdirAll("uploads/images", os.ModePerm)
	_ = os.MkdirAll("uploads/logos", os.ModePerm)
	_ = os.MkdirAll("uploads/products", os.ModePerm)
	return &UploadHandler{}
}

func (h *UploadHandler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		file, err = c.FormFile("image")
		if err != nil {
			file, err = c.FormFile("logo")
		}
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded (field 'file', 'image', or 'logo' required)"})
		return
	}

	// Enforce 2 MB size limit
	if file.Size > maxUploadSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": fmt.Sprintf("File is too large. Maximum allowed size is 2 MB (uploaded: %.2f MB)", float64(file.Size)/1024/1024),
		})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" && ext != ".svg" && ext != ".gif" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file format. Allowed: JPG, PNG, WEBP, SVG, GIF"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploaded file"})
		return
	}
	defer src.Close()

	uploadType := c.Query("type")
	var folder string
	switch uploadType {
	case "logo":
		folder = "uploads/logos"
	case "product":
		folder = "uploads/products"
	default:
		// Auto-detect from filename
		nameLower := strings.ToLower(file.Filename)
		if strings.Contains(nameLower, "logo") {
			folder = "uploads/logos"
		} else {
			folder = "uploads/images"
		}
	}
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to copy file contents"})
		return
	}

	url := "/" + filepath.ToSlash(dstPath)
	c.JSON(http.StatusOK, gin.H{
		"url":      url,
		"filename": filename,
		"size":     file.Size,
	})
}
