package data

import (
	"context"
	"time"

	"github.com/puchidemy/puchi-backend/app/auth/internal/biz"
	"github.com/puchidemy/puchi-backend/app/auth/internal/data/cache"
)

// SessionCache implements biz.SessionCache using Valkey.
type SessionCache struct {
	c *cache.Cache
}

// NewSessionCache creates a new SessionCache.
func NewSessionCache(c *cache.Cache) *SessionCache {
	return &SessionCache{c: c}
}

func (s *SessionCache) Set(ctx context.Context, sessionID string, data []byte, ttl time.Duration) error {
	return s.c.SetSession(ctx, sessionID, data, ttl)
}

func (s *SessionCache) Get(ctx context.Context, sessionID string) ([]byte, error) {
	return s.c.GetSession(ctx, sessionID)
}

func (s *SessionCache) Del(ctx context.Context, sessionID string) error {
	return s.c.DelSession(ctx, sessionID)
}

var _ biz.SessionCache = (*SessionCache)(nil)

// TokenBlacklist implements biz.TokenBlacklist using Valkey.
type TokenBlacklist struct {
	c *cache.Cache
}

// NewTokenBlacklist creates a new TokenBlacklist.
func NewTokenBlacklist(c *cache.Cache) *TokenBlacklist {
	return &TokenBlacklist{c: c}
}

func (b *TokenBlacklist) BlacklistJTI(ctx context.Context, jti string, ttl time.Duration) error {
	return b.c.BlacklistJTI(ctx, jti, ttl)
}

func (b *TokenBlacklist) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	return b.c.IsBlacklisted(ctx, jti)
}

var _ biz.TokenBlacklist = (*TokenBlacklist)(nil)

// RateLimiter implements biz.RateLimiter using Valkey.
type RateLimiter struct {
	c *cache.Cache
}

// NewRateLimiter creates a new RateLimiter.
func NewRateLimiter(c *cache.Cache) *RateLimiter {
	return &RateLimiter{c: c}
}

func (r *RateLimiter) Allow(ctx context.Context, key string, maxAttempts int, window time.Duration) (bool, error) {
	return r.c.Allow(ctx, key, maxAttempts, window)
}

var _ biz.RateLimiter = (*RateLimiter)(nil)
