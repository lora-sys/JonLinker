package service

import (
	"context"
	"fmt"
	"log"

	"joblinker/internal/model"
	"joblinker/internal/repository"
	"joblinker/pkg/ai"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// AgentMemoryService implements eino/memory.MemoryService for persistent storage
type AgentMemoryService struct {
	db       *gorm.DB
	vecRepo  *repository.VectorRepository
	prefRepo *repository.PreferenceVectorRepository
	aiClient *ai.Client
}

// NewAgentMemoryService creates a new memory service with database and vector repositories
func NewAgentMemoryService(
	db *gorm.DB,
	vecRepo *repository.VectorRepository,
	prefRepo *repository.PreferenceVectorRepository,
	aiClient *ai.Client,
) *AgentMemoryService {
	return &AgentMemoryService{
		db:       db,
		vecRepo:  vecRepo,
		prefRepo: prefRepo,
		aiClient: aiClient,
	}
}

// StoreMemory stores a memory entry with optional embedding for long-term recall
func (s *AgentMemoryService) StoreMemory(ctx context.Context, matchID uuid.UUID, memoryType model.MemoryType, content string) error {
	mem := &model.AgentMemory{
		ID:         uuid.New(),
		MatchID:    matchID,
		MemoryType: memoryType,
		Content:    content,
	}

	if memoryType == model.MemoryTypeLongTerm {
		embedding, err := s.aiClient.GenerateEmbedding(content)
		if err != nil {
			log.Printf("WARNING: failed to generate embedding for memory: %v", err)
		} else {
			mem.Embedding = pq.Float64Array(embedding)
			if s.vecRepo != nil {
				if vecErr := s.vecRepo.StoreVector(ctx, "agent_memories", mem.ID.String(), embedding, content, map[string]interface{}{
					"match_id":    matchID.String(),
					"memory_type": string(memoryType),
				}); vecErr != nil {
					log.Printf("WARNING: failed to store vector: %v", vecErr)
				}
			}
		}
	}

	return s.db.WithContext(ctx).Create(mem).Error
}

// GetRecentMemories retrieves recent memories for a match, ordered by creation time
func (s *AgentMemoryService) GetRecentMemories(ctx context.Context, matchID uuid.UUID, memoryType model.MemoryType, limit int) ([]*model.AgentMemory, error) {
	var memories []*model.AgentMemory
	query := s.db.WithContext(ctx).Where("match_id = ?", matchID)
	if memoryType != "" {
		query = query.Where("memory_type = ?", memoryType)
	}
	if err := query.Order("created_at DESC").Limit(limit).Find(&memories).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent memories: %w", err)
	}
	return memories, nil
}

// GetConversationSummary retrieves the latest conversation summary for a match
func (s *AgentMemoryService) GetConversationSummary(ctx context.Context, matchID uuid.UUID) (*model.ConversationSummary, error) {
	var summary model.ConversationSummary
	if err := s.db.WithContext(ctx).Where("match_id = ?", matchID).Order("created_at DESC").First(&summary).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get conversation summary: %w", err)
	}
	return &summary, nil
}

// StoreConversationSummary stores or updates a conversation summary
func (s *AgentMemoryService) StoreConversationSummary(ctx context.Context, summary *model.ConversationSummary) error {
	if summary.ID == uuid.Nil {
		summary.ID = uuid.New()
	}
	return s.db.WithContext(ctx).Save(summary).Error
}

// StorePreference stores a user preference with vector embedding for similarity search
func (s *AgentMemoryService) StorePreference(ctx context.Context, userID uuid.UUID, prefType model.PreferenceType, rawValue string) error {
	pref := &model.UserPreferenceVector{
		UserID:         userID,
		PreferenceType: prefType,
		RawValue:       rawValue,
	}

	embedding, err := s.aiClient.GenerateEmbedding(rawValue)
	if err != nil {
		log.Printf("WARNING: failed to generate preference embedding: %v", err)
	} else {
		pref.Embedding = pq.Float64Array(embedding)
	}

	return s.db.WithContext(ctx).Create(pref).Error
}

// SearchSimilarPreferences finds preferences similar to those of the given user
func (s *AgentMemoryService) SearchSimilarPreferences(ctx context.Context, userID uuid.UUID, prefType model.PreferenceType, limit int) ([]*model.UserPreferenceVector, error) {
	if s.prefRepo == nil {
		return nil, nil
	}
	return s.prefRepo.SearchSimilar(userID, prefType, nil, limit)
}
