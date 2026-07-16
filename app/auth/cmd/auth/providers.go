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
// Falls back to a deterministic dev key if not set.
func NewEncryptionKey() []byte {
	key := os.Getenv("TOTP_ENCRYPTION_KEY")
	if key == "" {
		k, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
		return k
	}
	k, err := hex.DecodeString(key)
	if err != nil {
		panic(fmt.Sprintf("invalid TOTP_ENCRYPTION_KEY: %v", err))
	}
	return k
}
