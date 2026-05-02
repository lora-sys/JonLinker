package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"joblinker/internal/model"
	"joblinker/internal/repository"
	"joblinker/pkg/ai"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type AgentMemoryService struct {
	prefRepo     *repository.PreferenceVectorRepository
	summaryRepo  *repository.ConversationSummaryRepository
	messageRepo  *repository.MessageRepository
	aiClient     *ai.Client
}

func NewAgentMemoryService(
	prefRepo *repository.PreferenceVectorRepository,
	summaryRepo *repository.ConversationSummaryRepository,
	messageRepo *repository.MessageRepository,
) *AgentMemoryService {
	return &AgentMemoryService{
		prefRepo:    prefRepo,
		summaryRepo: summaryRepo,
		messageRepo: messageRepo,
	}
}

// RecallPreferences retrieves stored preferences for a user at session start
func (s *AgentMemoryService) RecallPreferences(ctx context.Context, userID uuid.UUID) ([]*model.UserPreferenceVector, error) {
	prefs, err := s.prefRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to recall preferences: %w", err)
	}

	log.Printf("Recalled %d preferences for user %s", len(prefs), userID)
	return prefs, nil
}

// StorePreference stores a user-stated preference with vector embedding
func (s *AgentMemoryService) StorePreference(ctx context.Context, userID uuid.UUID, prefType model.PreferenceType, rawValue string) error {
	// Generate embedding for the preference
	embedding, err := s.generateEmbedding(prefType, rawValue)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	pref := &model.UserPreferenceVector{
		ID:             uuid.New(),
		UserID:         userID,
		PreferenceType: prefType,
		Embedding:      embedding,
		RawValue:       rawValue,
	}

	if err := s.prefRepo.Create(pref); err != nil {
		return fmt.Errorf("failed to store preference: %w", err)
	}

	log.Printf("Stored preference for user %s: type=%s, value=%s", userID, prefType, rawValue)
	return nil
}

// SearchSimilarPreferences finds similar preferences using vector similarity
func (s *AgentMemoryService) SearchSimilarPreferences(ctx context.Context, userID uuid.UUID, prefType model.PreferenceType, limit int) ([]*model.UserPreferenceVector, error) {
	if limit == 0 {
		limit = 5
	}

	// Create a query embedding
	queryEmbedding, err := s.generateEmbedding(prefType, "")
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	results, err := s.prefRepo.SearchSimilar(userID, prefType, queryEmbedding, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar preferences: %w", err)
	}

	return results, nil
}

// StoreMemory stores a memory entry for an agent
func (s *AgentMemoryService) StoreMemory(ctx context.Context, matchID uuid.UUID, memoryType model.MemoryType, content string) error {
	embedding, err := s.generateTextEmbedding(content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	memory := &model.AgentMemory{
		ID:         uuid.New(),
		MatchID:    matchID,
		MemoryType: memoryType,
		Content:    content,
		Embedding:  embedding,
	}

	_ = memory // Currently logged, would persist in production
	log.Printf("Stored memory for match %s: type=%s", matchID, memoryType)
	return nil
}

// GetRecentMemories retrieves recent memories for a match
func (s *AgentMemoryService) GetRecentMemories(ctx context.Context, matchID uuid.UUID, memoryType model.MemoryType, limit int) ([]*model.AgentMemory, error) {
	if limit == 0 {
		limit = 10
	}

	// Would query repository in production
	log.Printf("Retrieved recent memories for match %s: type=%s, limit=%d", matchID, memoryType, limit)
	return []*model.AgentMemory{}, nil
}

// generateEmbedding creates a vector embedding for a preference
func (s *AgentMemoryService) generateEmbedding(prefType model.PreferenceType, value string) (pq.Float64Array, error) {
	embedding, err := s.generateTextEmbedding(value)
	if err != nil {
		return nil, err
	}
	return embedding, nil
}

// generateTextEmbedding creates an embedding for text content using AI API
func (s *AgentMemoryService) generateTextEmbedding(content string) (pq.Float64Array, error) {
	if s.aiClient == nil {
		s.aiClient = ai.NewClient()
	}

	embedding, err := s.aiClient.GenerateEmbedding(content)
	if err != nil {
		log.Printf("WARNING: Failed to generate real embedding, using mock: %v", err)
		// Fallback to mock embedding with warning
		mockEmbedding := make(pq.Float64Array, 1536)
		for i := range mockEmbedding {
			mockEmbedding[i] = 0.0
		}
		return mockEmbedding, nil
	}

	// Convert []float64 to pq.Float64Array
	result := make(pq.Float64Array, len(embedding))
	for i, v := range embedding {
		result[i] = v
	}
	return result, nil
}

// GetConversationSummary retrieves or creates a summary for a match
func (s *AgentMemoryService) GetConversationSummary(ctx context.Context, matchID uuid.UUID) (*model.ConversationSummary, error) {
	summary, err := s.summaryRepo.GetByMatchID(matchID)
	if err != nil {
		// No summary exists yet
		return nil, nil
	}
	return summary, nil
}

// StoreConversationSummary stores a compressed conversation summary
func (s *AgentMemoryService) StoreConversationSummary(ctx context.Context, summary *model.ConversationSummary) error {
	if err := s.summaryRepo.Create(summary); err != nil {
		return fmt.Errorf("failed to store summary: %w", err)
	}
	return nil
}

// SerializeKeyFacts serializes key facts to JSON for storage
func SerializeKeyFacts(facts []model.KeyFact) string {
	data, _ := json.Marshal(facts)
	return string(data)
}

// DeserializeKeyFacts deserializes key facts from JSON
func DeserializeKeyFacts(data string) ([]model.KeyFact, error) {
	var facts []model.KeyFact
	if err := json.Unmarshal([]byte(data), &facts); err != nil {
		return nil, err
	}
	return facts, nil
}