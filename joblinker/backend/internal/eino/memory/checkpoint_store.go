package memory

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCheckPointStore struct {
	client     *redis.Client
	prefix     string
	defaultTTL time.Duration
}

func NewRedisCheckPointStore(client *redis.Client, prefix string) *RedisCheckPointStore {
	return &RedisCheckPointStore{
		client:     client,
		prefix:     prefix,
		defaultTTL: 24 * time.Hour,
	}
}

func (s *RedisCheckPointStore) Get(ctx context.Context, id string) ([]byte, bool, error) {
	key := s.prefix + id
	data, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return data, true, nil
}

func (s *RedisCheckPointStore) Set(ctx context.Context, id string, data []byte) error {
	key := s.prefix + id
	return s.client.Set(ctx, key, data, s.defaultTTL).Err()
}
