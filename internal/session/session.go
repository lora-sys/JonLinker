package session

import (
	"context"
	"encoding/json"
	"sync"
)

type Store struct {
	mu   sync.RWMutex
	data map[string]any
}

func NewStore() *Store {
	return &Store{data: make(map[string]any)}
}

func (s *Store) Get(ctx context.Context, key string, dest any) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	if !ok {
		return nil
	}
	b, _ := json.Marshal(v)
	return json.Unmarshal(b, dest)
}

func (s *Store) Set(ctx context.Context, key string, val any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = val
	return nil
}
