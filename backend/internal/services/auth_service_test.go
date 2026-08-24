package services

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/config"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_CheckPassword(t *testing.T) {
	service := NewAuthService(&config.JWTConfig{Secret: "test-secret"}, nil, nil, "localhost")

	password := "my-secure-password"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to generate test hash: %v", err)
	}

	if !service.CheckPassword(password, string(hash)) {
		t.Error("Expected CheckPassword to return true for valid password")
	}

	if service.CheckPassword("wrong-password", string(hash)) {
		t.Error("Expected CheckPassword to return false for invalid password")
	}
}

// Test Token Expiry (Gap #4 Fix Verification)
func TestAuthService_GenerateToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:        "test-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 24 * time.Hour,
	}
	service := NewAuthService(cfg, nil, nil, "localhost")

	userID := uuid.New()
	tenantID := uuid.New()

	expiryDur := 45 * time.Minute
	tokenStr, err := service.generateToken(userID, tenantID, nil, "admin", "access", 1, expiryDur)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		t.Fatalf("Invalid token claims")
	}

	expectedExpiry := time.Now().Add(expiryDur)
	actualExpiry := claims.ExpiresAt.Time

	diff := expectedExpiry.Sub(actualExpiry)
	if diff < -2*time.Second || diff > 2*time.Second {
		t.Errorf("Expected token to expire near %v, got %v (diff: %v)", expectedExpiry, actualExpiry, diff)
	}
}
