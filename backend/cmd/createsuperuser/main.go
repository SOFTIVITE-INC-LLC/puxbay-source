package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/models"
)

func main() {
	fmt.Println("Puxbay Create Superuser")
	fmt.Println("-----------------------")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		log.Fatal("Username cannot be empty")
	}

	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)
	if email == "" {
		log.Fatal("Email cannot be empty")
	}

	fmt.Print("Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		log.Fatalf("Error reading password: %v", err)
	}
	password := string(bytePassword)

	fmt.Print("Password (again): ")
	bytePasswordConfirm, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		log.Fatalf("Error reading password confirmation: %v", err)
	}
	passwordConfirm := string(bytePasswordConfirm)

	if password != passwordConfirm {
		log.Fatal("Passwords do not match.")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	user := models.User{
		ID:          uuid.New(),
		Username:    username,
		Email:       email,
		Password:    string(hashedPassword),
		IsActive:    true,
		IsSuperuser: true,
		IsStaff:     true,
	}

	if err := db.Create(&user).Error; err != nil {
		log.Fatalf("Failed to create superuser: %v", err)
	}

	fmt.Println("Superuser created successfully.")
}
