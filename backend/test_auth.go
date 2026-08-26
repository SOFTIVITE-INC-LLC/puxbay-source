package main

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/database"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	dbConn, err := database.Initialize(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	
	authService := services.NewAuthService(&cfg.JWT, dbConn, nil, cfg.Server.RootDomain)
	
	var user models.User
	if err := dbConn.Where("username = ?", "afari").First(&user).Error; err != nil {
		log.Fatalf("User not found: %v", err)
	}
	
	profile, err := authService.CurrentUser(user.ID, uuid.Nil)
	if err != nil {
		log.Fatalf("CurrentUser error: %v", err)
	}
	
	fmt.Printf("Success! Subdomain is '%s'\n", profile.Tenant.Subdomain)
}
