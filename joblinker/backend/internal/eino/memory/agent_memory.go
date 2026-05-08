package memory

import (
	"context"
	"log"

	"joblinker/internal/model"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

// MemoryService defines the interface for memory operations
// This breaks the import cycle between eino/memory and service
type MemoryService interface {
	StoreMemory(ctx context.Context, matchID uuid.UUID, memoryType model.MemoryType, content string) error
	GetRecentMemories(ctx context.Context, matchID uuid.UUID, memoryType model.MemoryType, limit int) ([]*model.AgentMemory, error)
	GetConversationSummary(ctx context.Context, matchID uuid.UUID) (*model.ConversationSummary, error)
	StoreConversationSummary(ctx context.Context, summary *model.ConversationSummary) error
	StorePreference(ctx context.Context, userID uuid.UUID, prefType model.PreferenceType, rawValue string) error
	SearchSimilarPreferences(ctx context.Context, userID uuid.UUID, prefType model.PreferenceType, limit int) ([]*model.UserPreferenceVector, error)
}

// AgentMemory wraps optional memory service with Eino Memory interface
// Provides conversation context management and summary capabilities
type AgentMemory struct {
	svc      MemoryService // Use interface to avoid import cycle
	matchID  uuid.UUID
	maxMsgs  int
	messages []*schema.Message
}

// NewAgentMemory creates a new Eino-compatible agent memory
// Pass nil for svc if persistent memory is not needed
func NewAgentMemory(svc MemoryService, matchID uuid.UUID) *AgentMemory {
	return &AgentMemory{
		svc:     svc,
		matchID: matchID,
		maxMsgs: 50, // Default window size
	}
}

// NewAgentMemoryWithWindow creates with custom window size
func NewAgentMemoryWithWindow(svc MemoryService, matchID uuid.UUID, maxMsgs int) *AgentMemory {
	return &AgentMemory{
		svc:     svc,
		matchID: matchID,
		maxMsgs: maxMsgs,
	}
}

// AddMessage adds a message to the conversation memory
func (m *AgentMemory) AddMessage(ctx context.Context, role schema.RoleType, content string) error {
	msg := &schema.Message{
		Role:    role,
		Content: content,
	}
	m.messages = append(m.messages, msg)

	// Trim if exceeds window
	if len(m.messages) > m.maxMsgs {
		m.messages = m.messages[len(m.messages)-m.maxMsgs:]
	}

	// Store in persistent memory if service is available
	if m.svc != nil {
		if err := m.svc.StoreMemory(ctx, m.matchID, model.MemoryTypeShortTerm, content); err != nil {
			log.Printf("Warning: failed to store memory: %v", err)
		}
	}

	return nil
}

// GetMessages returns the current conversation history
func (m *AgentMemory) GetMessages() []*schema.Message {
	return m.messages
}

// GetRecentMemories retrieves recent memories from persistent storage
func (m *AgentMemory) GetRecentMemories(ctx context.Context, memoryType model.MemoryType, limit int) ([]*model.AgentMemory, error) {
	if m.svc == nil {
		return nil, nil
	}
	return m.svc.GetRecentMemories(ctx, m.matchID, memoryType, limit)
}

// GetConversationSummary retrieves or creates a conversation summary
func (m *AgentMemory) GetConversationSummary(ctx context.Context) (*model.ConversationSummary, error) {
	if m.svc == nil {
		return nil, nil
	}
	return m.svc.GetConversationSummary(ctx, m.matchID)
}

// StoreConversationSummary stores a summary for future context
func (m *AgentMemory) StoreConversationSummary(ctx context.Context, summary *model.ConversationSummary) error {
	if m.svc == nil {
		return nil
	}
	return m.svc.StoreConversationSummary(ctx, summary)
}

// StorePreference stores a user preference with vector embedding
func (m *AgentMemory) StorePreference(ctx context.Context, prefType model.PreferenceType, rawValue string) error {
	if m.svc == nil {
		return nil
	}
	return m.svc.StorePreference(ctx, m.matchID, prefType, rawValue)
}

// SearchSimilarPreferences finds similar stored preferences
func (m *AgentMemory) SearchSimilarPreferences(ctx context.Context, prefType model.PreferenceType, limit int) ([]*model.UserPreferenceVector, error) {
	if m.svc == nil {
		return nil, nil
	}
	return m.svc.SearchSimilarPreferences(ctx, m.matchID, prefType, limit)
}

// Clear resets the in-memory conversation history
func (m *AgentMemory) Clear() {
	m.messages = nil
}

// WindowSize returns the configured max messages
func (m *AgentMemory) WindowSize() int {
	return m.maxMsgs
}

// MessageCount returns current number of messages in memory
func (m *AgentMemory) MessageCount() int {
	return len(m.messages)
}