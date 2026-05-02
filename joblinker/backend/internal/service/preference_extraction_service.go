package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"joblinker/internal/model"
	"joblinker/internal/repository"

	"github.com/google/uuid"
)

type PreferenceExtractionService struct {
	summaryRepo *repository.ConversationSummaryRepository
}

func NewPreferenceExtractionService(
	summaryRepo *repository.ConversationSummaryRepository,
) *PreferenceExtractionService {
	return &PreferenceExtractionService{
		summaryRepo: summaryRepo,
	}
}

// ExtractedPreference represents an auto-extracted preference
type ExtractedPreference struct {
	Type       string  `json:"type"`
	Value      string  `json:"value"`
	Confidence float64 `json:"confidence"`
}

// ExtractAndStorePreferences automatically extracts preferences from conversation
func (s *PreferenceExtractionService) ExtractAndStorePreferences(ctx context.Context, userID uuid.UUID, messages []*model.Message) error {
	log.Printf("Auto-extracting preferences for user %s", userID)

	// Extract preferences from conversation
	preferences := s.extractPreferencesFromMessages(messages)

	log.Printf("Extracted %d preferences for user %s: %+v", len(preferences), userID, preferences)
	return nil
}

func (s *PreferenceExtractionService) extractPreferencesFromMessages(messages []*model.Message) []ExtractedPreference {
	var preferences []ExtractedPreference

	for _, msg := range messages {
		content := msg.ContentXML

		// Extract salary preferences
		if s.containsSalaryContext(content) {
			preferences = append(preferences, ExtractedPreference{
				Type:       "salary_expectation",
				Value:      extractValueFromContent(content, "salary"),
				Confidence: 0.85,
			})
		}

		// Extract location preferences
		if s.containsLocationContext(content) {
			preferences = append(preferences, ExtractedPreference{
				Type:       "preferred_location",
				Value:      extractPrefLocationValue(content),
				Confidence: 0.9,
			})
		}

		// Extract job type preferences
		if s.containsJobTypeContext(content) {
			preferences = append(preferences, ExtractedPreference{
				Type:       "job_type",
				Value:      extractPrefJobTypeValue(content),
				Confidence: 0.85,
			})
		}
	}

	return preferences
}

func (s *PreferenceExtractionService) containsSalaryContext(content string) bool {
	lower := strings.ToLower(content)
	return strings.Contains(lower, "salary") ||
		strings.Contains(lower, "compensation") ||
		strings.Contains(lower, "$") ||
		strings.Contains(lower, "pay")
}

func (s *PreferenceExtractionService) containsLocationContext(content string) bool {
	lower := strings.ToLower(content)
	return strings.Contains(lower, "remote") ||
		strings.Contains(lower, "onsite") ||
		strings.Contains(lower, "hybrid") ||
		strings.Contains(lower, "location") ||
		strings.Contains(lower, "地点")
}

func (s *PreferenceExtractionService) containsJobTypeContext(content string) bool {
	lower := strings.ToLower(content)
	return strings.Contains(lower, "job type") ||
		strings.Contains(lower, "work mode") ||
		strings.Contains(lower, "full-time") ||
		strings.Contains(lower, "part-time")
}

func extractValueFromContent(content string, preferenceType string) string {
	// Simple extraction - in production would parse structured data
	return fmt.Sprintf("extracted_%s_value", preferenceType)
}

func extractPrefLocationValue(content string) string {
	lower := strings.ToLower(content)
	if strings.Contains(lower, "remote") {
		return "remote"
	}
	if strings.Contains(lower, "hybrid") {
		return "hybrid"
	}
	if strings.Contains(lower, "onsite") {
		return "onsite"
	}
	return "unspecified"
}

func extractPrefJobTypeValue(content string) string {
	lower := strings.ToLower(content)
	if strings.Contains(lower, "full-time") {
		return "full-time"
	}
	if strings.Contains(lower, "part-time") {
		return "part-time"
	}
	return "full-time"
}

// StorePreference stores an extracted preference (placeholder for vector storage)
func (s *PreferenceExtractionService) StorePreference(ctx context.Context, userID uuid.UUID, pref ExtractedPreference) error {
	log.Printf("Storing preference for user %s: type=%s, value=%s, confidence=%.2f",
		userID, pref.Type, pref.Value, pref.Confidence)
	// In production, would store to vector database (pgvector)
	return nil
}

// RecallPreferences retrieves stored preferences for a user
func (s *PreferenceExtractionService) RecallPreferences(ctx context.Context, userID uuid.UUID) ([]ExtractedPreference, error) {
	// In production, would query vector database
	log.Printf("Recalling preferences for user %s", userID)
	return []ExtractedPreference{}, nil
}