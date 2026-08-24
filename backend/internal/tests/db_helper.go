package tests

import (
	"log"
	"os"

	"github.com/softivite/puxbay/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB initializes an in-memory SQLite database for testing and runs migrations.
func SetupTestDB() (*gorm.DB, func()) {
	// Use an in-memory sqlite db. The cache=shared is sometimes useful but a simple in-memory is fine.
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				LogLevel: logger.Silent, // drop logs during tests
			},
		),
	})

	if err != nil {
		panic("failed to connect database")
	}

	// Migrate schemas
	err = db.AutoMigrate(
		&models.Tenant{},
		&models.User{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
		&models.Customer{},
		&models.Category{},
		&models.Branch{},
		// Add other models as needed
	)
	if err != nil {
		panic("failed to migrate database: " + err.Error())
	}

	teardown := func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	return db, teardown
}
