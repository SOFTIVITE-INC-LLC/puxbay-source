package main

import (
	"fmt"
	"log"
	"os"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=puxbay port=5432 sslmode=disable TimeZone=UTC"
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		dsn = dbURL
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Just count how many products exist in any schema
	type Product struct {
		ID string
		Name string
		BranchID *string
	}
	var products []Product
    // Try to query softivite.products
	err = db.Raw("SELECT id, name, branch_id FROM softivite.products LIMIT 5").Scan(&products).Error
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	for _, p := range products {
        bid := "NULL"
        if p.BranchID != nil {
            bid = *p.BranchID
        }
		fmt.Printf("Product: %s, Branch: %s\n", p.Name, bid)
	}
}
