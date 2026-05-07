package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"joblinker/internal/model"
	"joblinker/internal/repository"
	"joblinker/pkg/chroma"

	"github.com/google/uuid"
)

type AgentMemoryService struct {
	chromaClient *chroma.Client
	prefRepo     *repository.PreferenceVectorRepository
	summaryRepo  *repository.ConversationSummaryRepository
	messageRepo  *repository.MessageRepository
}

const (
	CollectionAgentMemories  = "agent_memories"
	CollectionUserPreferences = "user_preferences"
)

func NewAgentMemoryService(
	chromaClient *chroma.Client,
	prefRepo *repository.PreferenceVectorRepository,
	summaryRepo *repository.ConversationSummaryRepository,
	messageRepo *repository.MessageRepository,
) *AgentMemoryService {
	return &AgentMemoryService{
		chromaClient: chromaClient,
		prefRepo:     prefRepo,
		summaryRepo:  summaryRepo,
		messageRepo:  messageRepo,
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
	meta := map[string]interface{}{
		"user_id": userID.String(),
		"type":    string(prefType),
	}
	id := uuid.New().String()
	err := s.chromaClient.Add(CollectionUserPreferences, []string{id}, []string{rawValue}, []map[string]interface{}{meta})
	if err != nil {
		return fmt.Errorf("failed to store preference in Chroma: %w", err)
	}

	log.Printf("Stored preference for user %s: type=%s, value=%s", userID, prefType, rawValue)
	return nil
}

// SearchSimilarPreferences finds similar preferences using vector similarity
func (s *AgentMemoryService) SearchSimilarPreferences(ctx context.Context, userID uuid.UUID, prefType model.PreferenceType, limit int) ([]*model.UserPreferenceVector, error) {
	if limit == 0 {
		limit = 5
	}

	where := map[string]interface{}{
		"user_id": userID.String(),
		"type":    string(prefType),
	}
	result, err := s.chromaClient.Query(CollectionUserPreferences, []string{string(prefType)}, limit, where)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar preferences: %w", err)
	}

	var prefs []*model.UserPreferenceVector
	if len(result.IDs) > 0 && len(result.IDs[0]) > 0 {
		for i := 0; i < len(result.IDs[0]); i++ {
			pref := &model.UserPreferenceVector{
				ID:       uuid.MustParse(result.IDs[0][i]),
				UserID:   userID,
				RawValue: result.Documents[0][i],
			}
			prefs = append(prefs, pref)
		}
	}
	return prefs, nil
}

// StoreMemory stores a memory entry for an agent
func (s *AgentMemoryService) StoreMemory(ctx context.Context, matchID uuid.UUID, memoryType model.MemoryType, content string) error {
	meta := map[string]interface{}{
		"match_id":    matchID.String(),
		"memory_type": string(memoryType),
	}
	id := uuid.New().String()
	err := s.chromaClient.Add(CollectionAgentMemories, []string{id}, []string{content}, []map[string]interface{}{meta})
	if err != nil {
		return fmt.Errorf("failed to store memory in Chroma: %w", err)
	}

	log.Printf("Stored memory for match %s: type=%s", matchID, memoryType)
	return nil
}

// GetRecentMemories retrieves recent memories for a match
func (s *AgentMemoryService) GetRecentMemories(ctx context.Context, matchID uuid.UUID, memoryType model.MemoryType, limit int) ([]*model.AgentMemory, error) {
	if limit == 0 {
		limit = 10
	}

	where := map[string]interface{}{
		"match_id":    matchID.String(),
		"memory_type": string(memoryType),
	}
	result, err := s.chromaClient.Query(CollectionAgentMemories, []string{matchID.String()}, limit, where)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent memories: %w", err)
	}

	var memories []*model.AgentMemory
	if len(result.IDs) > 0 && len(result.IDs[0]) > 0 {
		for i := 0; i < len(result.IDs[0]); i++ {
			mem := &model.AgentMemory{
				ID:      uuid.MustParse(result.IDs[0][i]),
				MatchID: matchID,
				Content: result.Documents[0][i],
			}
			memories = append(memories, mem)
		}
	}
	return memories, nil
}

// GetConversationSummary retrieves or creates a summary for a match
func (s *AgentMemoryService) GetConversationSummary(ctx context.Context, matchID uuid.UUID) (*model.ConversationSummary, error) {
	summary, err := s.summaryRepo.GetByMatchID(matchID)
	if err != nil {
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