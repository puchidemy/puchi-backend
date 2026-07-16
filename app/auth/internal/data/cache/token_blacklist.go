package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const blacklistKeyPrefix = "blacklist"

// BlacklistJTI adds a JTI (JWT ID) to the blacklist with the given TTL.
// The TTL should match the remaining lifetime of the access token (at most 15 min).
func (c *Cache) BlacklistJTI(ctx context.Context, jti string, ttl time.Duration) error {
	key := keyPrefix(blacklistKeyPrefix, jti)
	err := c.client.Set(ctx, key, "1", ttl).Err()
	return fmtRedisErr("BlacklistJTI", err)
}

// IsBlacklisted checks whether a JTI is in the blacklist.
func (c *Cache) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	key := keyPrefix(blacklistKeyPrefix, jti)
	_, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmtRedisErr("IsBlacklisted", err)
	}
	return true, nil
}
