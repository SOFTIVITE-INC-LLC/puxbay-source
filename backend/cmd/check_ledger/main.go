package main

import (
	"fmt"
	"log"

	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatal(err)
	}

	var count int64
	db.Exec("SET search_path TO thinkce")
	db.Table("ledger_accounts").Count(&count)
	fmt.Printf("Ledger accounts count: %d\n", count)
}
