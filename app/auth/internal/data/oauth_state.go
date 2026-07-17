package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/puchidemy/puchi-backend/app/auth/internal/biz"
)

const oauthStateKeyPrefix = "oauth_state"

// OAuthStateRepo manages OAuth2 state storage in Valkey.
// States are short-lived (10 min TTL) and used for PKCE + CSRF protection.
type OAuthStateRepo struct {
	cache biz.SessionCache
}

// NewOAuthStateRepo creates a new OAuthStateRepo.
func NewOAuthStateRepo(cache biz.SessionCache) *OAuthStateRepo {
	return &OAuthStateRepo{cache: cache}
}

// Save stores OAuth state data in Valkey with the given TTL.
func (r *OAuthStateRepo) Save(ctx context.Context, state string, data *biz.OAuthStateData, ttl time.Duration) error {
	key := fmt.Sprintf("%s:%s", oauthStateKeyPrefix, state)
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal oauth state: %w", err)
	}
	return r.cache.Set(ctx, key, payload, ttl)
}

// GetAndDelete atomically retrieves and removes OAuth state from Valkey.
// Returns nil if the state does not exist or has expired.
func (r *OAuthStateRepo) GetAndDelete(ctx context.Context, state string) (*biz.OAuthStateData, error) {
	key := fmt.Sprintf("%s:%s", oauthStateKeyPrefix, state)
	data, err := r.cache.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("get oauth state: %w", err)
	}
	if data == nil {
		return nil, nil
	}

	// Delete the key after retrieval (best-effort)
	_ = r.cache.Del(ctx, key)

	var result biz.OAuthStateData
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal oauth state: %w", err)
	}
	return &result, nil
}
