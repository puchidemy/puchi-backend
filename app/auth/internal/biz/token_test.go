package biz

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestIssueAndVerifyAccessToken(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	uc := NewTokenUsecase(TokenConfig{
		PrivateKey:      privateKey,
		PublicKey:       &privateKey.PublicKey,
		KeyID:           "test-key",
		Issuer:          "http://localhost:8080",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 30 * 24 * time.Hour,
	})

	claims := AccessTokenClaims{
		UserID:        uuid.New(),
		Email:         "test@example.com",
		EmailVerified: true,
		Roles:         []string{"student"},
		PermVersion:   1,
		SessionID:     uuid.New(),
	}

	tokenStr, err := uc.IssueAccessToken(claims)
	if err != nil {
		t.Fatalf("IssueAccessToken failed: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("token should not be empty")
	}

	parsed, err := uc.VerifyAccessToken(tokenStr)
	if err != nil {
		t.Fatalf("VerifyAccessToken failed: %v", err)
	}

	if parsed.UserID != claims.UserID {
		t.Errorf("UserID mismatch: got %v, want %v", parsed.UserID, claims.UserID)
	}
	if parsed.Email != claims.Email {
		t.Errorf("Email mismatch: got %s, want %s", parsed.Email, claims.Email)
	}
	if parsed.EmailVerified != claims.EmailVerified {
		t.Errorf("EmailVerified mismatch: got %v, want %v", parsed.EmailVerified, claims.EmailVerified)
	}
	if len(parsed.Roles) != 1 || parsed.Roles[0] != "student" {
		t.Errorf("Roles mismatch: got %v", parsed.Roles)
	}
	if parsed.PermVersion != 1 {
		t.Errorf("PermVersion mismatch: got %d", parsed.PermVersion)
	}
	if parsed.SessionID != claims.SessionID {
		t.Errorf("SessionID mismatch: got %v, want %v", parsed.SessionID, claims.SessionID)
	}
}

func TestExpiredTokenRejected(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	uc := NewTokenUsecase(TokenConfig{
		PrivateKey:      privateKey,
		PublicKey:       &privateKey.PublicKey,
		KeyID:           "test-key",
		Issuer:          "http://localhost:8080",
		AccessTokenTTL:  -45 * time.Second, // past leeway (30s)
		RefreshTokenTTL: 30 * 24 * time.Hour,
	})

	tokenStr, err := uc.IssueAccessToken(AccessTokenClaims{
		UserID: uuid.New(),
		Email:  "test@example.com",
		Roles:  []string{"student"},
	})
	if err != nil {
		t.Fatalf("IssueAccessToken failed: %v", err)
	}

	_, err = uc.VerifyAccessToken(tokenStr)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestWrongIssuerRejected(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	uc := NewTokenUsecase(TokenConfig{
		PrivateKey:      privateKey,
		PublicKey:       &privateKey.PublicKey,
		KeyID:          "test-key",
		Issuer:          "http://correct-issuer",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 30 * 24 * time.Hour,
	})

	tokenStr, err := uc.IssueAccessToken(AccessTokenClaims{
		UserID: uuid.New(),
		Email:  "test@example.com",
		Roles:  []string{"student"},
	})
	if err != nil {
		t.Fatalf("IssueAccessToken failed: %v", err)
	}

	// Token from correct issuer should verify fine
	_, err = uc.VerifyAccessToken(tokenStr)
	if err != nil {
		t.Fatalf("expected token from same issuer to verify: %v", err)
	}

	// Create a different TokenUsecase with a different issuer
	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate other key: %v", err)
	}
	ucWrong := NewTokenUsecase(TokenConfig{
		PrivateKey:      otherKey,
		PublicKey:       &otherKey.PublicKey,
		KeyID:           "test-key",
		Issuer:          "http://wrong-issuer",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 30 * 24 * time.Hour,
	})

	// Verify a token from one issuer against a different issuer should fail
	_, err = ucWrong.VerifyAccessToken(tokenStr)
	if err == nil {
		t.Fatal("expected error when verifying token from wrong issuer")
	}
}
