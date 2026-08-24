package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheService struct {
	client *redis.Client
}

func NewCacheService(redisURL string) *CacheService {
	if redisURL == "" {
		return &CacheService{}
	}
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return &CacheService{}
	}
	client := redis.NewClient(opts)
	return &CacheService{client: client}
}

func (s *CacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if s.client == nil {
		return nil
	}
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, key, bytes, ttl).Err()
}

func (s *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
	if s.client == nil {
		return redis.Nil
	}
	bytes, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, dest)
}

func (s *CacheService) Delete(ctx context.Context, key string) error {
	if s.client == nil {
		return nil
	}
	return s.client.Del(ctx, key).Err()
}
