package agent

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Memory handles agent persistent state and context
type Memory struct {
	AgentID       uuid.UUID      `json:"agent_id"`
	History       []MemoryEntry  `json:"history"`
	Preferences   map[string]any `json:"preferences"`
	LearnedFacts  map[string]bool `json:"learned_facts"`
	MaxEntries    int            `json:"-"`
	ContextBudget int            `json:"-"`
}

// MemoryEntry is a single memory unit
type MemoryEntry struct {
	Type       string    `json:"type"` // "interaction", "preference", "outcome", "fact"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Importance float64  `json:"importance"` // 0-1
	Compressed bool     `json:"compressed"`
}

// NewMemory creates a new agent memory
func NewMemory(agentID uuid.UUID) *Memory {
	return &Memory{
		AgentID:       agentID,
		History:       []MemoryEntry{},
		Preferences:   make(map[string]any),
		LearnedFacts:  make(map[string]bool),
		MaxEntries:    1000,
		ContextBudget: 100, // max entries in active context
	}
}

// AddEntry adds a new memory entry
func (m *Memory) AddEntry(entry MemoryEntry) {
	entry.Timestamp = time.Now()
	m.History = append(m.History, entry)

	// Evict if over budget
	if len(m.History) > m.MaxEntries {
		m.Evict()
	}
}

// Evict removes low-importance entries when memory is full
func (m *Memory) Evict() {
	if len(m.History) <= m.ContextBudget {
		return
	}

	// Compress older entries
	compressed := m.compressToEntry()
	var newHistory []MemoryEntry
	newHistory = append(newHistory, *compressed)

	// Keep most recent entries up to context budget
	remaining := m.ContextBudget - 1
	if remaining > 0 && len(m.History) > remaining {
		start := len(m.History) - remaining
		newHistory = append(newHistory, m.History[start:]...)
	}

	m.History = newHistory
}

func (m *Memory) compressToEntry() *MemoryEntry {
	count := len(m.History)
	if count == 0 {
		return &MemoryEntry{Type: "summary", Content: "Empty memory"}
	}

	// Count interaction types
	interactions, preferences, outcomes := 0, 0, 0
	for _, e := range m.History {
		switch e.Type {
		case "interaction":
			interactions++
		case "preference":
			preferences++
		case "outcome":
			outcomes++
		}
	}

	return &MemoryEntry{
		Type:       "summary",
		Timestamp:  time.Now(),
		Importance: 0.5,
		Compressed: true,
		Content:    fmt.Sprintf(`{"interactions":%d,"preferences":%d,"outcomes":%d,"span":%d}`, interactions, preferences, outcomes, count),
	}
}

// GetActiveContext returns recent entries within context budget
func (m *Memory) GetActiveContext() []MemoryEntry {
	if len(m.History) <= m.ContextBudget {
		return m.History
	}
	return m.History[len(m.History)-m.ContextBudget:]
}

// LearnFact stores a discovered fact
func (m *Memory) LearnFact(fact string) {
	m.LearnedFacts[fact] = true
	m.AddEntry(MemoryEntry{
		Type:       "fact",
		Content:    fact,
		Importance: 0.8,
	})
}

// StorePreference stores a learned preference
func (m *Memory) StorePreference(key string, value any) {
	m.Preferences[key] = value
	m.AddEntry(MemoryEntry{
		Type:       "preference",
		Content:    fmt.Sprintf("%s: %v", key, value),
		Importance: 0.7,
	})
}
