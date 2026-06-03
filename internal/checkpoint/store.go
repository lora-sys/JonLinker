package checkpoint

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync"

	"github.com/cloudwego/eino/compose"
)

var _ compose.CheckPointStore = (*Store)(nil)

type Store struct {
	mu       sync.RWMutex
	data     map[string][]byte
	filePath string
}

func NewStore() *Store {
	return &Store{data: make(map[string][]byte)}
}

func NewPersistentStore(filePath string) *Store {
	s := NewStore()
	if filePath == "" {
		return s
	}
	s.filePath = filePath
	s.load()
	return s
}

func (s *Store) Get(_ context.Context, id string) ([]byte, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[id]
	if !ok {
		return nil, false, nil
	}
	return v, true, nil
}

func (s *Store) Set(_ context.Context, id string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = data
	if s.filePath != "" {
		return s.persist()
	}
	return nil
}

func (s *Store) load() {
	b, err := os.ReadFile(s.filePath)
	if err != nil {
		return
	}
	var data map[string][]byte
	if err := json.Unmarshal(b, &data); err != nil {
		return
	}
	s.data = data
}

func (s *Store) DeletePrefix(_ context.Context, prefix string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k := range s.data {
		if strings.HasPrefix(k, prefix) {
			delete(s.data, k)
		}
	}
	if s.filePath != "" {
		return s.persist()
	}
	return nil
}

func (s *Store) persist() error {
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, b, 0644)
}
