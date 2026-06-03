package memory

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"

	"github.com/cloudwego/eino/schema"
)

// FileStore persists session memory to a JSON file.
// Thread-safe: sync.RWMutex guards both in-memory cache and file I/O.
// Each session's messages are Gob-encoded and stored as base64 in JSON.
type FileStore struct {
	mu       sync.RWMutex
	data     map[string][]byte
	filePath string
}

func NewFileStore(filePath string) *FileStore {
	s := &FileStore{data: make(map[string][]byte), filePath: filePath}
	if filePath != "" {
		s.load()
	}
	return s
}

func (s *FileStore) Read(_ context.Context, sessionID string) ([]*schema.Message, error) {
	s.mu.RLock()
	b := s.data[sessionID]
	s.mu.RUnlock()
	return DecodeMessages(b)
}

func (s *FileStore) Write(_ context.Context, sessionID string, msgs []*schema.Message) error {
	b, err := EncodeMessages(msgs)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.data[sessionID] = b
	s.mu.Unlock()

	if s.filePath != "" {
		if err := s.persist(); err != nil {
			log.Printf("file store persist: %v", err)
		}
	}
	return nil
}

func (s *FileStore) load() {
	b, err := os.ReadFile(s.filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("file store load read: %v", err)
		}
		return
	}
	var data map[string][]byte
	if err := json.Unmarshal(b, &data); err != nil {
		log.Printf("file store load unmarshal: %v", err)
		return
	}
	s.data = data
}

func (s *FileStore) persist() error {
	s.mu.RLock()
	b, err := json.MarshalIndent(s.data, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return err
	}
	tmpPath := s.filePath + ".tmp"
	if err := os.WriteFile(tmpPath, b, 0600); err != nil {
		return err
	}
	return os.Rename(tmpPath, s.filePath)
}
