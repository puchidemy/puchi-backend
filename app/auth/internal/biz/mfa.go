package biz

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"

	authCrypto "github.com/puchidemy/puchi-backend/app/auth/internal/data/crypto"
)

// MFAUsecase handles TOTP multi-factor authentication operations.
type MFAUsecase struct {
	totpRepo      TOTPRepo
	encryptionKey []byte
}

// NewMFAUsecase creates a new MFAUsecase.
func NewMFAUsecase(totpRepo TOTPRepo, encryptionKey []byte) *MFAUsecase {
	return &MFAUsecase{
		totpRepo:      totpRepo,
		encryptionKey: encryptionKey,
	}
}

// TOTPEnrollment holds the data returned after enrolling a TOTP secret.
type TOTPEnrollment struct {
	Secret        string   `json:"secret"`
	QRCodeURL     string   `json:"qr_code_url"`
	RecoveryCodes []string `json:"recovery_codes"`
}

// Enroll generates a new TOTP secret and recovery codes for a user.
func (uc *MFAUsecase) Enroll(ctx context.Context, userID uuid.UUID, email string) (*TOTPEnrollment, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Puchi",
		AccountName: email,
	})
	if err != nil {
		return nil, fmt.Errorf("generate TOTP key: %w", err)
	}

	// Generate 10 recovery codes (8 chars, uppercase alphanumeric)
	recoveryCodes := make([]string, 10)
	for i := range recoveryCodes {
		code, err := generateRandomCode(8)
		if err != nil {
			return nil, fmt.Errorf("generate recovery code: %w", err)
		}
		recoveryCodes[i] = code
	}

	codesJSON, err := json.Marshal(recoveryCodes)
	if err != nil {
		return nil, fmt.Errorf("marshal recovery codes: %w", err)
	}

	encSecret, err := authCrypto.Encrypt([]byte(key.Secret()), uc.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt TOTP secret: %w", err)
	}

	encCodes, err := authCrypto.Encrypt(codesJSON, uc.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt recovery codes: %w", err)
	}

	secret := &TOTPSecret{
		UserID:          userID,
		EncryptedSecret: encSecret,
		EncryptedCodes:  encCodes,
		IsEnabled:       false,
	}

	if err := uc.totpRepo.Upsert(ctx, secret); err != nil {
		return nil, fmt.Errorf("store TOTP secret: %w", err)
	}

	return &TOTPEnrollment{
		Secret:        key.Secret(),
		QRCodeURL:     key.URL(),
		RecoveryCodes: recoveryCodes,
	}, nil
}

// VerifyCode validates a TOTP code and enables MFA on first successful verification.
func (uc *MFAUsecase) VerifyCode(ctx context.Context, userID uuid.UUID, code string) error {
	stored, err := uc.totpRepo.GetByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("get TOTP secret: %w", err)
	}

	secretBytes, err := authCrypto.Decrypt(stored.EncryptedSecret, uc.encryptionKey)
	if err != nil {
		return fmt.Errorf("decrypt TOTP secret: %w", err)
	}

	secret := string(secretBytes)

	valid := totp.Validate(code, secret)
	if !valid {
		return fmt.Errorf("invalid TOTP code")
	}

	if !stored.IsEnabled {
		if err := uc.totpRepo.Enable(ctx, userID); err != nil {
			return fmt.Errorf("enable TOTP: %w", err)
		}
	}

	return nil
}

// ValidateCode checks a TOTP code during login flow (MFA challenge).
func (uc *MFAUsecase) ValidateCode(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	stored, err := uc.totpRepo.GetByUser(ctx, userID)
	if err != nil || !stored.IsEnabled {
		return false, nil
	}

	secretBytes, err := authCrypto.Decrypt(stored.EncryptedSecret, uc.encryptionKey)
	if err != nil {
		return false, nil
	}

	return totp.Validate(code, string(secretBytes)), nil
}

// Disable removes MFA for a user.
func (uc *MFAUsecase) Disable(ctx context.Context, userID uuid.UUID) error {
	return uc.totpRepo.Disable(ctx, userID)
}

// ValidateRecoveryCode checks a recovery code and marks it used.
func (uc *MFAUsecase) ValidateRecoveryCode(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	stored, err := uc.totpRepo.GetByUser(ctx, userID)
	if err != nil {
		return false, nil
	}

	codesBytes, err := authCrypto.Decrypt(stored.EncryptedCodes, uc.encryptionKey)
	if err != nil {
		return false, nil
	}

	var codes []string
	if err := json.Unmarshal(codesBytes, &codes); err != nil {
		return false, nil
	}

	code = strings.ToUpper(strings.TrimSpace(code))
	for i, c := range codes {
		if c == code {
			codes = append(codes[:i], codes[i+1:]...)
			newCodesJSON, _ := json.Marshal(codes)
			encCodes, err := authCrypto.Encrypt(newCodesJSON, uc.encryptionKey)
			if err != nil {
				return false, nil
			}
			if err := uc.totpRepo.Upsert(ctx, &TOTPSecret{
				UserID:          userID,
				EncryptedSecret: stored.EncryptedSecret,
				EncryptedCodes:  encCodes,
				IsEnabled:       stored.IsEnabled,
			}); err != nil {
				return false, nil
			}
			return true, nil
		}
	}

	return false, nil
}

func generateRandomCode(length int) (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	code := make([]byte, length)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code[i] = charset[n.Int64()]
	}
	return string(code), nil
}
