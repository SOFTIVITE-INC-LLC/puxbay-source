package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"github.com/softivite/puxbay/internal/config"
)

// EncryptionService handles field-level encryption using AES-GCM.
// This replaces Django's EncryptedTextField backed by Fernet.
type EncryptionService struct {
	key []byte
}

// NewEncryptionService creates a new encryption service from the config.
func NewEncryptionService(cfg *config.FernetConfig) (*EncryptionService, error) {
	if cfg.Key == "" {
		return &EncryptionService{key: nil}, nil // No-op if no key configured
	}

	// Decode the base64-encoded key
	key, err := base64.URLEncoding.DecodeString(cfg.Key)
	if err != nil {
		// Try standard base64
		key, err = base64.StdEncoding.DecodeString(cfg.Key)
		if err != nil {
			return nil, errors.New("invalid encryption key: must be base64-encoded")
		}
	}

	// AES-256 requires a 32-byte key; if key is longer (Fernet keys are 32 bytes),
	// use the first 32 bytes
	if len(key) < 16 {
		return nil, errors.New("encryption key too short: must be at least 16 bytes")
	}
	if len(key) > 32 {
		key = key[:32]
	}

	return &EncryptionService{key: key}, nil
}

// Encrypt encrypts plaintext and returns a base64-encoded ciphertext.
func (e *EncryptionService) Encrypt(plaintext string) (string, error) {
	if e.key == nil || plaintext == "" {
		return plaintext, nil
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64-encoded ciphertext back to plaintext.
func (e *EncryptionService) Decrypt(encoded string) (string, error) {
	if e.key == nil || encoded == "" {
		return encoded, nil
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		// If it can't be decoded, return as-is (might be plaintext from old data)
		return encoded, nil
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return encoded, nil // Too short to be encrypted, return as-is
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		// Decryption failed — might be plaintext or encrypted with a different key
		return encoded, nil
	}

	return string(plaintext), nil
}
