package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/softivite/puxbay/internal/models"
)

func main() {
	fmt.Println("Puxbay Create Superuser")
	fmt.Println("-----------------------")

	// Allow passing an explicit path to the .env file
	envFile := flag.String("env-file", "", "Path to .env file (default: searches cwd and binary dir)")
	flag.Parse()

	// Resolve the env file path:
	// 1. Explicit flag  2. cwd/.env  3. <binary-dir>/.env
	resolvedEnv := resolveEnvFile(*envFile)

	viper.SetConfigFile(resolvedEnv)
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: could not read %s: %v — using environment variables only", resolvedEnv, err)
	} else {
		fmt.Printf("✔ Loaded config from: %s\n", resolvedEnv)
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		viper.GetString("DB_HOST"),
		viper.GetString("DB_USER"),
		viper.GetString("DB_PASSWORD"),
		viper.GetString("DB_NAME"),
		viper.GetString("DB_PORT"),
		viper.GetString("DB_SSLMODE"),
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

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		allPerms := `["dashboard:read", "tenants:read", "tenants:write", "domains:read", "domains:write", "billing:read", "billing:write", "pricing_plans:read", "pricing_plans:write", "promo_codes:read", "promo_codes:write", "content:read", "content:write", "referrals:read", "broadcasts:read", "broadcasts:write", "apps:read", "apps:write", "webhooks:read", "webhooks:write", "backups:read", "backups:write", "api_keys:read", "api_keys:write", "security:read", "security:write", "admin_users:read", "admin_users:write", "settings:write"]`

		adminUser := models.AdminUser{
			UserID:      user.ID,
			Permissions: allPerms,
		}

		if err := tx.Create(&adminUser).Error; err != nil {
			return fmt.Errorf("failed to create admin_user: %w", err)
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Failed to create superuser: %v", err)
	}

	fmt.Println("Superuser created successfully with all permissions.")
}

// resolveEnvFile finds the .env file to use, in priority order:
//  1. Explicit --env-file flag
//  2. .env in current working directory
//  3. .env next to the binary (useful on servers where binary is in /opt/app)
func resolveEnvFile(explicit string) string {
	if explicit != "" {
		return explicit
	}

	// cwd/.env
	if _, err := os.Stat(".env"); err == nil {
		abs, _ := filepath.Abs(".env")
		return abs
	}

	// <binary-dir>/.env
	exe, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(exe), ".env")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	// fall back to cwd (viper will warn if missing)
	return ".env"
}
