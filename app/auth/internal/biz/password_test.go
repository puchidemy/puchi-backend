package biz

import (
	"testing"
)

func TestHashAndVerifyPassword(t *testing.T) {
	password := "mySecurePass123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Fatal("hash should not be empty")
	}

	// Verify correct password
	ok, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if !ok {
		t.Fatal("expected password to verify successfully")
	}

	// Verify wrong password
	ok, err = VerifyPassword("wrongPassword", hash)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if ok {
		t.Fatal("expected wrong password to fail verification")
	}
}

func TestHashPasswordUniqueSalt(t *testing.T) {
	password := "samePassword"

	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)

	if hash1 == hash2 {
		t.Fatal("same password should produce different hashes due to unique salt")
	}
}

func TestVerifyPasswordInvalidHash(t *testing.T) {
	_, err := VerifyPassword("test", "invalid-hash-format")
	if err == nil {
		t.Fatal("expected error for invalid hash format")
	}
}
