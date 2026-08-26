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
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/softivite/puxbay/internal/models"
)

func main() {
	fmt.Println("Puxbay Create Superuser")
	fmt.Println("-----------------------")

	envFile := flag.String("env-file", "", "Path to .env file (default: searches cwd then binary dir)")
	flag.Parse()

	// Load the .env file into os environment so we can read with os.Getenv
	resolvedEnv := resolveEnvFile(*envFile)
	if err := loadEnvFile(resolvedEnv); err != nil {
		log.Printf("Warning: %v — will use existing environment variables", err)
	} else {
		fmt.Printf("✔ Loaded config from: %s\n", resolvedEnv)
	}

	// Read DB settings, falling back to sane defaults
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "puxbay_go")
	dbSSL := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPass, dbName, dbSSL,
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

	if password != string(bytePasswordConfirm) {
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

	fmt.Println("✅ Superuser created successfully. (Django-style: is_superuser=true grants all permissions implicitly)")
}

// getEnv reads an environment variable, returning fallback if not set or empty.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// loadEnvFile parses a .env file and sets variables into the process environment.
// It handles:
//   - Blank lines and whitespace-only lines (skipped)
//   - Comment lines: # comment, \# comment, // comment (skipped)
//   - export KEY=VALUE syntax
//   - Quoted values: KEY="value" or KEY='value'
//   - Inline comments after values (stripped)
//   - Already-set env vars are NOT overwritten (os.Setenv only if not set)
func loadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open %s: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip blank lines and comment lines (all variants)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, `\#`) || strings.HasPrefix(line, "//") {
			continue
		}

		// Strip leading "export " keyword
		line = strings.TrimPrefix(line, "export ")
		line = strings.TrimSpace(line)

		idx := strings.IndexByte(line, '=')
		if idx < 0 {
			continue // not a KEY=VALUE line, skip silently
		}

		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])

		// Strip inline comments (unquoted # after value)
		val = stripInlineComment(val)

		// Strip surrounding quotes
		val = unquote(val)

		// Only set if not already in environment (don't override explicit env vars)
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}

	return scanner.Err()
}

// stripInlineComment removes trailing # comment from an unquoted value.
// e.g.  "somevalue  # this is a comment" → "somevalue"
// Quoted values are left intact.
func stripInlineComment(s string) string {
	if len(s) == 0 {
		return s
	}
	// If the value is quoted, don't strip
	if (s[0] == '"' || s[0] == '\'') {
		return s
	}
	if idx := strings.Index(s, " #"); idx >= 0 {
		s = s[:idx]
	}
	return strings.TrimSpace(s)
}

// unquote removes surrounding single or double quotes from a string.
func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// resolveEnvFile finds the .env file to use, in priority order:
//  1. Explicit --env-file flag
//  2. .env in current working directory
//  3. .env next to the binary
func resolveEnvFile(explicit string) string {
	if explicit != "" {
		return explicit
	}
	if _, err := os.Stat(".env"); err == nil {
		abs, _ := filepath.Abs(".env")
		return abs
	}
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), ".env")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ".env"
}
