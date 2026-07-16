package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Set stores session metadata in Valkey with the given TTL.
// Alias for SetSession to satisfy biz.SessionCache interface.
func (c *Cache) Set(ctx context.Context, sessionID string, dataJSON []byte, ttl time.Duration) error {
	return c.SetSession(ctx, sessionID, dataJSON, ttl)
}

// Get retrieves session metadata from Valkey.
// Alias for GetSession to satisfy biz.SessionCache interface.
func (c *Cache) Get(ctx context.Context, sessionID string) ([]byte, error) {
	return c.GetSession(ctx, sessionID)
}

// Del deletes a session from Valkey.
// Alias for DelSession to satisfy biz.SessionCache interface.
func (c *Cache) Del(ctx context.Context, sessionID string) error {
	return c.DelSession(ctx, sessionID)
}

const sessionKeyPrefix = "session"

// SetSession stores session metadata in Valkey with the given TTL.
// The data is serialized as JSON bytes by the caller.
func (c *Cache) SetSession(ctx context.Context, sessionID string, dataJSON []byte, ttl time.Duration) error {
	key := keyPrefix(sessionKeyPrefix, sessionID)
	err := c.client.Set(ctx, key, dataJSON, ttl).Err()
	return fmtRedisErr("SetSession", err)
}

// GetSession retrieves session metadata from Valkey.
// Returns nil, nil if the key does not exist.
func (c *Cache) GetSession(ctx context.Context, sessionID string) ([]byte, error) {
	key := keyPrefix(sessionKeyPrefix, sessionID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmtRedisErr("GetSession", err)
	}
	return data, nil
}

// DelSession deletes a session from Valkey.
func (c *Cache) DelSession(ctx context.Context, sessionID string) error {
	key := keyPrefix(sessionKeyPrefix, sessionID)
	err := c.client.Del(ctx, key).Err()
	return fmtRedisErr("DelSession", err)
}
