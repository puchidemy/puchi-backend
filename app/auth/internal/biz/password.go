package biz

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2Params defines the parameters for argon2id password hashing.
type Argon2Params struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
	SaltLen uint32
}

// DefaultArgon2Params are the recommended parameters for password hashing.
// These values balance security and performance:
//   - Time=3: 3 iterations (minimum recommended)
//   - Memory=64*1024: 64 MB memory usage
//   - Threads=4: 4 parallelism threads
//   - KeyLen=32: 256-bit hash output
//   - SaltLen=16: 128-bit random salt
var DefaultArgon2Params = Argon2Params{
	Time:    3,
	Memory:  64 * 1024,
	Threads: 4,
	KeyLen:  32,
	SaltLen: 16,
}

// HashPassword hashes a password using argon2id with default parameters and
// returns the encoded hash string.
//
// Format: $argon2id$v=19$m=65536,t=3,p=4$<base64-salt>$<base64-hash>
func HashPassword(password string) (string, error) {
	return HashPasswordWithParams(password, DefaultArgon2Params)
}

// HashPasswordWithParams hashes a password using argon2id with the given parameters.
func HashPasswordWithParams(password string, params Argon2Params) (string, error) {
	salt := make([]byte, params.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, params.Time, params.Memory, params.Threads, params.KeyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		params.Memory, params.Time, params.Threads, b64Salt, b64Hash)

	return encoded, nil
}

// VerifyPassword checks a password against an encoded argon2id hash.
// It uses constant-time comparison to prevent timing attacks.
func VerifyPassword(password, encodedHash string) (bool, error) {
	params, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	computed := argon2.IDKey([]byte(password), salt, params.Time, params.Memory, params.Threads, params.KeyLen)

	return subtle.ConstantTimeCompare(hash, computed) == 1, nil
}

// decodeHash parses an encoded argon2id hash string into its components.
// Expected format: $argon2id$v=19$m=65536,t=3,p=4$<base64-salt>$<base64-hash>
func decodeHash(encoded string) (Argon2Params, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return Argon2Params{}, nil, nil, fmt.Errorf("invalid hash format: expected 6 parts, got %d", len(parts))
	}

	if parts[1] != "argon2id" {
		return Argon2Params{}, nil, nil, fmt.Errorf("unsupported algorithm: %s", parts[1])
	}

	var params Argon2Params
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.Memory, &params.Time, &params.Threads)
	if err != nil {
		return Argon2Params{}, nil, nil, fmt.Errorf("failed to parse params: %w", err)
	}
	params.KeyLen = 32 // fixed for our implementation

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return Argon2Params{}, nil, nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return Argon2Params{}, nil, nil, fmt.Errorf("failed to decode hash: %w", err)
	}

	return params, salt, hash, nil
}
