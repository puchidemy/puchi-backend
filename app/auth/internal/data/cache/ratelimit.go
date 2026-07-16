package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const ratelimitKeyPrefix = "ratelimit"

// Allow checks if an action identified by key is within the rate limit.
// It implements a sliding window using Redis sorted sets:
// - Each request is scored by its timestamp (UnixNano)
// - ZREMRANGEBYSCORE removes entries outside the window
// - ZCARD counts entries in the current window
//
// Returns true if the action is allowed, false if rate limited.
// Key patterns (from spec): ratelimit:login:{email}, ratelimit:magic_link:{email}, etc.
func (c *Cache) Allow(ctx context.Context, key string, maxAttempts int, window time.Duration) (bool, error) {
	now := time.Now()
	windowStart := now.Add(-window)

	redisKey := keyPrefix(ratelimitKeyPrefix, key)
	nowScore := now.UnixNano()
	minScore := windowStart.UnixNano()

	pipe := c.client.Pipeline()

	// Remove entries outside the window
	pipe.ZRemRangeByScore(ctx, redisKey, "0", formatInt64(minScore))

	// Add current entry
	pipe.ZAdd(ctx, redisKey, redis.Z{
		Score:  float64(nowScore),
		Member: nowScore,
	})

	// Count entries in window
	countCmd := pipe.ZCard(ctx, redisKey)

	// Set TTL on the key so it auto-expires
	pipe.Expire(ctx, redisKey, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmtRedisErr("Allow", err)
	}

	count, err := countCmd.Result()
	if err != nil {
		return false, fmtRedisErr("Allow.count", err)
	}

	return count <= int64(maxAttempts), nil
}

func formatInt64(v int64) string {
	buf := make([]byte, 0, 20)
	neg := v < 0
	if neg {
		v = -v
	}
	for v > 0 {
		buf = append([]byte{byte('0' + v%10)}, buf...)
		v /= 10
	}
	if len(buf) == 0 {
		buf = []byte("0")
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
