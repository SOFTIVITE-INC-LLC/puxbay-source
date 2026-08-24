package services

import (
	"encoding/base32"
	"testing"
	"time"
)

func TestSecurityService_GenerateTOTPSecret(t *testing.T) {
	secret1, err := generateTOTPSecret()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	secret2, err := generateTOTPSecret()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if secret1 == secret2 {
		t.Errorf("Expected dynamically generated secrets to be unique, got duplicate: %s", secret1)
	}

	if len(secret1) < 16 {
		t.Errorf("Expected secret to be at least 16 chars, got length %d", len(secret1))
	}
}

func TestSecurityService_HMACBasedOTP(t *testing.T) {
	// Fixed key and counter should always produce the same HOTP code
	key := []byte("test-secret-key")
	counter := uint64(123456789)

	code1 := hmacBasedOTP(key, counter)
	code2 := hmacBasedOTP(key, counter)

	if code1 != code2 {
		t.Errorf("Expected deterministic HOTP generation, got %s and %s", code1, code2)
	}

	if len(code1) != 6 {
		t.Errorf("Expected 6 digit code, got %d", len(code1))
	}
}

func TestSecurityService_ValidateTOTPCode(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP" // base32 encoding of some bytes

	// Get current time window counter
	counter := uint64(time.Now().Unix()) / 30

	// We must pass the decoded secret bytes to hmacBasedOTP
	decodedSecret, _ := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	code := hmacBasedOTP(decodedSecret, counter)

	if !validateTOTPCode(secret, code) {
		t.Errorf("Expected valid code %s to be accepted for secret %s", code, secret)
	}

	if validateTOTPCode(secret, "000000") {
		t.Errorf("Expected invalid code 000000 to be rejected")
	}
}
