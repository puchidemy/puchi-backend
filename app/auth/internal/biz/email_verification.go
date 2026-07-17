package biz

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
)

// Sentinel errors for email verification operations.
var (
	ErrInvalidVerificationCode = errors.New("invalid or expired verification code")
	ErrEmailAlreadyVerified    = errors.New("email already verified")
)

// EmailVerification represents an email verification record.
type EmailVerification struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	CodeHash   string
	Email      string
	ExpiresAt  time.Time
	VerifiedAt *time.Time
	CreatedAt  time.Time
}

// EmailVerificationRepo defines the data access interface for email verifications.
type EmailVerificationRepo interface {
	Create(ctx context.Context, ev *EmailVerification) error
	GetByUserAndHash(ctx context.Context, userID uuid.UUID, codeHash string) (*EmailVerification, error)
	MarkVerified(ctx context.Context, id uuid.UUID) error
}

// EmailVerificationConfig holds configuration for the email verification usecase.
type EmailVerificationConfig struct {
	FrontendURL string
	EmailFrom   string
}

// EmailVerificationUsecase handles email verification business logic.
type EmailVerificationUsecase struct {
	evRepo      EmailVerificationRepo
	userRepo    UserRepo
	tokenUC     *TokenUsecase
	publisher   EventPublisher
	frontendURL string
	emailFrom   string
}

// NewEmailVerificationUsecase creates a new EmailVerificationUsecase.
func NewEmailVerificationUsecase(evRepo EmailVerificationRepo, userRepo UserRepo, tokenUC *TokenUsecase, publisher EventPublisher, cfg EmailVerificationConfig) *EmailVerificationUsecase {
	return &EmailVerificationUsecase{
		evRepo:      evRepo,
		userRepo:    userRepo,
		tokenUC:     tokenUC,
		publisher:   publisher,
		frontendURL: cfg.FrontendURL,
		emailFrom:   cfg.EmailFrom,
	}
}

// Send generates a verification code, stores its hash, and publishes an email
// event via NATS outbox. Returns without error on success (email sending is async).
func (uc *EmailVerificationUsecase) Send(ctx context.Context, userID uuid.UUID, email string) (string, error) {
	// Generate 6-digit code
	code, err := generateVerificationCode()
	if err != nil {
		return "", fmt.Errorf("generate code: %w", err)
	}

	// Hash the code and store in DB
	codeHash := hashVerificationCode(code)
	ev := &EmailVerification{
		UserID:   userID,
		CodeHash: codeHash,
		Email:    email,
	}
	if err := uc.evRepo.Create(ctx, ev); err != nil {
		return "", fmt.Errorf("store verification: %w", err)
	}

	// Publish email.send NATS event with the raw code
	if err := uc.publisher.Publish(ctx, "email.send", map[string]any{
		"to":       email,
		"from":     uc.emailFrom,
		"template": "email-verify",
		"data": map[string]any{
			"code":      code,
			"user_id":   userID.String(),
			"email":     email,
			"frontend_url": uc.frontendURL,
		},
	}); err != nil {
		// Log but don't fail — the verification is stored and can be retried
		return code, nil
	}

	return code, nil
}

// Verify checks the code and marks the user's email as verified.
func (uc *EmailVerificationUsecase) Verify(ctx context.Context, userID uuid.UUID, code string) error {
	codeHash := hashVerificationCode(code)

	ev, err := uc.evRepo.GetByUserAndHash(ctx, userID, codeHash)
	if err != nil {
		return ErrInvalidVerificationCode
	}

	if err := uc.userRepo.SetEmailVerified(ctx, userID); err != nil {
		return fmt.Errorf("set email verified: %w", err)
	}

	if err := uc.evRepo.MarkVerified(ctx, ev.ID); err != nil {
		return fmt.Errorf("mark verification used: %w", err)
	}

	return nil
}

// generateVerificationCode generates a random 6-digit code.
func generateVerificationCode() (string, error) {
	max := big.NewInt(900000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("rand int: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()+100000), nil
}

// hashVerificationCode returns the SHA-256 hash of a code as a base64url-encoded string.
func hashVerificationCode(code string) string {
	h := sha256.Sum256([]byte(code))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
