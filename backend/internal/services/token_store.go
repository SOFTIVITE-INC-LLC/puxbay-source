package services

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const tokenDenylistPrefix = "token:denied:"

// TokenStore defines the interface for denylisting JWT tokens.
type TokenStore interface {
	// DenyToken adds a token's JTI to the denylist for the given TTL.
	DenyToken(ctx context.Context, jti string, ttl time.Duration) error
	// DenyTokens adds multiple tokens to the denylist for the given TTL using a pipeline.
	DenyTokens(ctx context.Context, jtis []string, ttl time.Duration) error
	// IsDenied returns true if the given JTI is in the denylist.
	IsDenied(ctx context.Context, jti string) bool
	// AreDenied returns a slice of booleans indicating if each JTI is in the denylist.
	AreDenied(ctx context.Context, jtis []string) []bool
}

// RedisTokenStore implements TokenStore using Redis.
type RedisTokenStore struct {
	client *redis.Client
}

// NewRedisTokenStore creates a RedisTokenStore from a Redis URL string.
// Example URL: "redis://localhost:6379/0"
func NewRedisTokenStore(redisURL string) (*RedisTokenStore, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}
	client := redis.NewClient(opts)
	// Verify connectivity
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	return &RedisTokenStore{client: client}, nil
}

// DenyToken stores the token JTI in Redis with the given TTL.
// After TTL expires, the entry is automatically cleaned up by Redis.
func (s *RedisTokenStore) DenyToken(ctx context.Context, jti string, ttl time.Duration) error {
	key := tokenDenylistPrefix + jti
	return s.client.Set(ctx, key, "denied", ttl).Err()
}

// DenyTokens stores multiple token JTIs in Redis using pipelining.
func (s *RedisTokenStore) DenyTokens(ctx context.Context, jtis []string, ttl time.Duration) error {
	pipe := s.client.Pipeline()
	for _, jti := range jtis {
		key := tokenDenylistPrefix + jti
		pipe.Set(ctx, key, "denied", ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// IsDenied returns true if the JTI exists in the Redis denylist.
func (s *RedisTokenStore) IsDenied(ctx context.Context, jti string) bool {
	key := tokenDenylistPrefix + jti
	val, err := s.client.Exists(ctx, key).Result()
	return err == nil && val > 0
}

// AreDenied checks multiple JTIs efficiently using pipelining.
func (s *RedisTokenStore) AreDenied(ctx context.Context, jtis []string) []bool {
	pipe := s.client.Pipeline()
	var cmds []*redis.IntCmd
	for _, jti := range jtis {
		key := tokenDenylistPrefix + jti
		cmds = append(cmds, pipe.Exists(ctx, key))
	}
	_, _ = pipe.Exec(ctx)

	results := make([]bool, len(jtis))
	for i, cmd := range cmds {
		val, err := cmd.Result()
		results[i] = err == nil && val > 0
	}
	return results
}

// NoopTokenStore is a fallback that never denies tokens.
// Used when Redis is unavailable (e.g., local dev without Redis).
type NoopTokenStore struct{}

func (n *NoopTokenStore) DenyToken(_ context.Context, _ string, _ time.Duration) error {
	return nil
}

func (n *NoopTokenStore) IsDenied(_ context.Context, _ string) bool {
	return false
}

func (n *NoopTokenStore) DenyTokens(_ context.Context, _ []string, _ time.Duration) error {
	return nil
}

func (n *NoopTokenStore) AreDenied(_ context.Context, jtis []string) []bool {
	res := make([]bool, len(jtis))
	return res
}
