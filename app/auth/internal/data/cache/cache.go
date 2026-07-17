package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Cache wraps a Valkey (Redis-compatible) client for caching operations.
type Cache struct {
	client *redis.Client
}

// New creates a new Cache with the given Valkey configuration.
func New(addr string, password string, db int, poolSize int) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
		PoolSize: poolSize,
	})
	return &Cache{client: client}
}

// Close closes the underlying Valkey connection.
func (c *Cache) Close() error {
	return c.client.Close()
}

// Ping checks connectivity to Valkey.
func (c *Cache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Client returns the underlying redis client for advanced operations.
func (c *Cache) Client() *redis.Client {
	return c.client
}

// keyPrefix returns a namespaced key with the given prefix and parts.
func keyPrefix(parts ...string) string {
	s := ""
	for i, p := range parts {
		if i > 0 {
			s += ":"
		}
		s += p
	}
	return s
}

// fmtRedisErr wraps an error with context.
func fmtRedisErr(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("cache.%s: %w", op, err)
}
