package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/puchidemy/puchi-backend/app/auth/internal/biz"
	"github.com/puchidemy/puchi-backend/app/auth/internal/conf"
)

// NewEmailVerificationConfig creates the email verification configuration
// from the auth config. The frontend URL is taken from the first allowed CORS origin.
func NewEmailVerificationConfig(authCfg *conf.Auth) biz.EmailVerificationConfig {
	frontendURL := "https://puchi.io.vn"
	if len(authCfg.CorsAllowedOrigins) > 0 {
		frontendURL = authCfg.CorsAllowedOrigins[0]
	}
	return biz.EmailVerificationConfig{
		FrontendURL: frontendURL,
		EmailFrom:   "noreply@puchi.io.vn",
	}
}

// NewTokenConfig creates a TokenConfig from the auth configuration.
func NewTokenConfig(authCfg *conf.Auth) (biz.TokenConfig, error) {
	keyData, err := os.ReadFile(authCfg.PrivateKeyPath)
	if err != nil {
		return biz.TokenConfig{}, fmt.Errorf("read private key: %w", err)
	}
	block, _ := pem.Decode(keyData)
	if block == nil {
		return biz.TokenConfig{}, fmt.Errorf("failed to decode PEM private key block")
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return biz.TokenConfig{}, fmt.Errorf("parse PKCS8 private key: %w", err)
	}
	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return biz.TokenConfig{}, fmt.Errorf("private key is not an RSA key")
	}
	return biz.TokenConfig{
		PrivateKey:      rsaKey,
		PublicKey:       &rsaKey.PublicKey,
		KeyID:           authCfg.KeyId,
		Issuer:          authCfg.Issuer,
		AccessTokenTTL:  time.Duration(authCfg.AccessTokenTtl) * time.Second,
		RefreshTokenTTL: time.Duration(authCfg.RefreshTokenTtl) * time.Second,
	}, nil
}

// NewEncryptionKey provides the AES-256-GCM encryption key for TOTP secrets.
// Reads from TOTP_ENCRYPTION_KEY env var (hex-encoded, 64 hex chars = 32 bytes).
// Returns an error if the env var is not set or is invalid.
func NewEncryptionKey() ([]byte, error) {
	keyHex := os.Getenv("TOTP_ENCRYPTION_KEY")
	if keyHex == "" {
		return nil, fmt.Errorf("TOTP_ENCRYPTION_KEY environment variable is required")
	}
	k, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid TOTP_ENCRYPTION_KEY: %w", err)
	}
	return k, nil
}

// NewFrontendURL returns the frontend URL from the first CORS allowed origin.
// Falls back to a sensible default if not configured.
func NewFrontendURL(authCfg *conf.Auth) string {
	if len(authCfg.CorsAllowedOrigins) > 0 {
		return authCfg.CorsAllowedOrigins[0]
	}
	return "https://puchi.io.vn"
}
