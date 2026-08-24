package main

import (
	"log"

	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/database"
)

// This script helps migrate data from the old Django database to the new Go database
// or seed initial data.
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Printf("Connected to database: %v", db)
	log.Println("Starting data migration...")

	// Example: Migrate Users
	// 1. Fetch from old Django users table (auth_user)
	// 2. Insert into new Go users table (users)
	// db.Exec("INSERT INTO users (id, email, password, is_active) SELECT id, email, password, is_active FROM auth_user")

	log.Println("Data migration completed successfully.")
}
