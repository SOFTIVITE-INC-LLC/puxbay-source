package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/database"
	"github.com/softivite/puxbay/internal/models"
)

func main() {
	cfg, _ := config.Load()
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatal(err)
	}

	var logs []models.TelemetryLog
	err = db.Model(&models.TelemetryLog{}).Order("created_at desc").Limit(5).Find(&logs).Error
	if err != nil {
		log.Fatal(err)
	}

	b, err := json.Marshal(logs)
	if err != nil {
		log.Fatalf("JSON marshal error: %v", err)
	}
	fmt.Printf("JSON: %s\n", b)
}
