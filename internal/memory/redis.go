package memory

import (
	"context"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/redis/go-redis/v9"
)

const defaultTTL = 1 * time.Hour

type RedisStore struct {
	cli *redis.Client
}

func NewRedisStore(cli *redis.Client) *RedisStore {
	return &RedisStore{cli: cli}
}

func (s *RedisStore) Write(ctx context.Context, sessionID string, msgs []*schema.Message) error {
	b, err := EncodeMessages(msgs)
	if err != nil {
		return err
	}
	return s.cli.Set(ctx, sessionID, b, defaultTTL).Err()
}

func (s *RedisStore) Read(ctx context.Context, sessionID string) ([]*schema.Message, error) {
	res, err := s.cli.Get(ctx, sessionID).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return DecodeMessages(res)
}
