package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&Product{}, &Category{}, &Tenant{}, &Branch{})
	require.NoError(t, err)
	return db
}

func TestProduct_OptimisticLocking(t *testing.T) {
	db := setupTestDB(t)

	t.Run("Optimistic Locking Prevents Lost Updates", func(t *testing.T) {
		// 1. Create a product
		_ = uuid.New()
		product := Product{
			Name:         "Test Product",
			SKU:          "SKU-123",
			CurrentStock: 10,
			SellingPrice: 50,
			Version:      1,
		}

		err := db.Create(&product).Error
		require.NoError(t, err)

		// 2. Simulate two concurrent transactions fetching the exact same record version
		var t1Product, t2Product Product
		require.NoError(t, db.First(&t1Product, "id = ?", product.ID).Error)
		require.NoError(t, db.First(&t2Product, "id = ?", product.ID).Error)

		// 3. Transaction 1 successfully sells 2 items
		t1Product.CurrentStock -= 2
		err = db.Model(&t1Product).Select("current_stock", "version").Updates(t1Product).Error
		require.NoError(t, err)

		// 4. Transaction 2 attempts to sell 5 items using the stale version
		t2Product.CurrentStock -= 5
		res := db.Model(&t2Product).Select("current_stock", "version").Updates(t2Product)

		// 5. The second transaction should effect 0 rows because the version has changed
		assert.NoError(t, res.Error)
		assert.Equal(t, int64(0), res.RowsAffected, "Second transaction should have been blocked by optimistic locking")
	})
}

func TestProduct_PrependCDN(t *testing.T) {
	tests := []struct {
		name     string
		image    string
		envCDN   string
		expected string
	}{
		{
			name:     "Already HTTP",
			image:    "https://example.com/img.jpg",
			envCDN:   "https://cdn.example.com",
			expected: "https://example.com/img.jpg",
		},
		{
			name:     "Relative path with CDN",
			image:    "images/prod1.jpg",
			envCDN:   "https://cdn.example.com/",
			expected: "https://cdn.example.com/images/prod1.jpg",
		},
		{
			name:     "Relative path without CDN env",
			image:    "images/prod1.jpg",
			envCDN:   "",
			expected: "images/prod1.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CDN_URL", tt.envCDN)
			img := tt.image
			prependCDN(&img)
			assert.Equal(t, tt.expected, img)
		})
	}
}
