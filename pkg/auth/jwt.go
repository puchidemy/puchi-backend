package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

var (
	ErrNoSession      = errors.New("no session found")
	ErrSessionExpired = errors.New("session expired")
)

var defaultHTTPClient = &http.Client{Timeout: 5 * time.Second}

// SessionInfo is the authenticated identity resolved from a Limen session.
type SessionInfo struct {
	UserID   string
	Email    string
	Username string
	Roles    []string
}

// SessionValidator validates opaque session tokens via auth-service introspect.
type SessionValidator struct {
	authServiceURL string
	httpClient     *http.Client
	cacheTTL       time.Duration

	mu    sync.RWMutex
	cache map[string]cacheEntry
}

type cacheEntry struct {
	info      SessionInfo
	expiresAt time.Time
}

// Claims is kept for compatibility with older call sites; prefer SessionInfo.
type Claims = SessionInfo

// JWTValidator is an alias for SessionValidator (legacy name used by Wire).
type JWTValidator = SessionValidator

// ParseAndValidate validates a Bearer session token and returns identity claims.
func (v *SessionValidator) ParseAndValidate(ctx context.Context, tokenStr string) (*SessionInfo, error) {
	if tokenStr == "" {
		return nil, ErrNoSession
	}

	key := hashToken(tokenStr)
	if info, ok := v.getCached(key); ok {
		return &info, nil
	}

	info, err := v.introspect(ctx, tokenStr)
	if err != nil {
		return nil, err
	}
	v.setCached(key, *info)
	return info, nil
}

func (v *SessionValidator) introspect(ctx context.Context, tokenStr string) (*SessionInfo, error) {
	url := v.authServiceURL + "/internal/session"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	client := v.httpClient
	if client == nil {
		client = defaultHTTPClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("introspect session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrNoSession
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("introspect session: status %d", resp.StatusCode)
	}

	var body struct {
		UserID   string `json:"user_id"`
		Email    string `json:"email"`
		Username string `json:"username"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if body.UserID == "" {
		return nil, ErrNoSession
	}
	return &SessionInfo{
		UserID:   body.UserID,
		Email:    body.Email,
		Username: body.Username,
	}, nil
}

func (v *SessionValidator) getCached(key string) (SessionInfo, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.cache == nil {
		return SessionInfo{}, false
	}
	e, ok := v.cache[key]
	if !ok || time.Now().After(e.expiresAt) {
		return SessionInfo{}, false
	}
	return e.info, true
}

func (v *SessionValidator) setCached(key string, info SessionInfo) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.cache == nil {
		v.cache = make(map[string]cacheEntry)
	}
	ttl := v.cacheTTL
	if ttl <= 0 {
		ttl = 60 * time.Second
	}
	v.cache[key] = cacheEntry{info: info, expiresAt: time.Now().Add(ttl)}
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
