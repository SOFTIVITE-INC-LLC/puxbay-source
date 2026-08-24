package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type ContentHandler struct {
	db *gorm.DB
}

func NewContentHandler(db *gorm.DB) *ContentHandler {
	return &ContentHandler{db: db}
}

// ListPages returns a list of CMS content pages.
// For now, it maps BlogPosts to the ContentPage frontend format.
func (h *ContentHandler) ListPages(c *gin.Context) {
	var posts []models.BlogPost

	// Fetch blog posts
	if err := h.db.Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch content pages"})
		return
	}

	// Map to frontend ContentPage format
	type ContentPage struct {
		ID         string `json:"id"`
		Title      string `json:"title"`
		Slug       string `json:"slug"`
		Status     string `json:"status"`
		LastEdited string `json:"last_edited"`
	}

	pages := make([]ContentPage, 0, len(posts))
	for _, post := range posts {
		pages = append(pages, ContentPage{
			ID:         post.ID.String(),
			Title:      post.Title,
			Slug:       post.Slug,
			Status:     post.Status,
			LastEdited: post.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	c.JSON(http.StatusOK, gin.H{"pages": pages})
}
