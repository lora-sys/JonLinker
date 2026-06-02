package memory

import (
	"context"
	"sync"

	"github.com/cloudwego/eino/schema"
)

type InMemoryStore struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{data: make(map[string][]byte)}
}

func (s *InMemoryStore) Write(_ context.Context, sessionID string, msgs []*schema.Message) error {
	b, err := EncodeMessages(msgs)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.data[sessionID] = b
	s.mu.Unlock()
	return nil
}

func (s *InMemoryStore) Read(_ context.Context, sessionID string) ([]*schema.Message, error) {
	s.mu.RLock()
	b := s.data[sessionID]
	s.mu.RUnlock()
	return DecodeMessages(b)
}
