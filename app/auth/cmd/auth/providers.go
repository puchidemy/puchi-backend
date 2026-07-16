package main

import (
	"crypto/rsa"
	"crypto/x509"
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
