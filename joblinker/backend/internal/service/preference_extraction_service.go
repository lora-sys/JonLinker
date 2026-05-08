package service

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"joblinker/internal/model"
	"joblinker/internal/repository"
	"joblinker/pkg/chroma"

	"github.com/google/uuid"
)

type PreferenceExtractionService struct {
	chromaClient *chroma.Client
	summaryRepo  *repository.ConversationSummaryRepository
}

func NewPreferenceExtractionService(
	chromaClient *chroma.Client,
	summaryRepo *repository.ConversationSummaryRepository,
) *PreferenceExtractionService {
	return &PreferenceExtractionService{
		chromaClient: chromaClient,
		summaryRepo:  summaryRepo,
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

	preferences := s.extractPreferencesFromMessages(messages)

	for _, pref := range preferences {
		if err := s.StorePreference(ctx, userID, pref); err != nil {
			log.Printf("Failed to store preference: %v", err)
		}
	}

	log.Printf("Extracted %d preferences for user %s", len(preferences), userID)
	return nil
}

func (s *PreferenceExtractionService) extractPreferencesFromMessages(messages []*model.Message) []ExtractedPreference {
	var preferences []ExtractedPreference

	for _, msg := range messages {
		content := msg.ContentXML

		if s.containsSalaryContext(content) {
			preferences = append(preferences, ExtractedPreference{
				Type:       "salary_expectation",
				Value:      extractValueFromContent(content, "salary"),
				Confidence: 0.85,
			})
		}

		if s.containsLocationContext(content) {
			preferences = append(preferences, ExtractedPreference{
				Type:       "preferred_location",
				Value:      extractPrefLocationValue(content),
				Confidence: 0.9,
			})
		}

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

var salaryPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\$[\d,]+(?:\.\d+)?`),      // $120,000 or $120000.00
	regexp.MustCompile(`(\d+)\s*k\b`),              // 120k
	regexp.MustCompile(`salary[:\s]*(\d[\d,]*)`),   // salary: 120,000
}

func extractValueFromContent(content string, preferenceType string) string {
	if preferenceType == "salary" {
		for _, re := range salaryPatterns {
			if match := re.FindString(content); match != "" {
				return strings.TrimSpace(match)
			}
		}
	}
	// Fallback: return first 100 chars of content
	if len(content) > 100 {
		return content[:100]
	}
	return content
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

// StorePreference stores an extracted preference to Chroma
func (s *PreferenceExtractionService) StorePreference(ctx context.Context, userID uuid.UUID, pref ExtractedPreference) error {
	meta := map[string]interface{}{
		"user_id":   userID.String(),
		"pref_type": pref.Type,
		"confidence": pref.Confidence,
	}
	id := uuid.New().String()
	err := s.chromaClient.Add("user_preferences", []string{id}, []string{pref.Value}, []map[string]interface{}{meta})
	if err != nil {
		return fmt.Errorf("failed to store preference in Chroma: %w", err)
	}
	log.Printf("Stored preference for user %s: type=%s, value=%s, confidence=%.2f",
		userID, pref.Type, pref.Value, pref.Confidence)
	return nil
}

// RecallPreferences retrieves stored preferences for a user from Chroma
func (s *PreferenceExtractionService) RecallPreferences(ctx context.Context, userID uuid.UUID) ([]ExtractedPreference, error) {
	where := map[string]interface{}{
		"user_id": userID.String(),
	}
	result, err := s.chromaClient.Query("user_preferences", []string{"preference"}, 10, where)
	if err != nil {
		return nil, fmt.Errorf("failed to recall preferences from Chroma: %w", err)
	}

	var prefs []ExtractedPreference
	if len(result.IDs) > 0 && len(result.IDs[0]) > 0 {
		for i := 0; i < len(result.IDs[0]); i++ {
			var prefType, confidence string
			if len(result.Metadatas) > 0 && len(result.Metadatas[0]) > i {
				if t, ok := result.Metadatas[0][i]["pref_type"].(string); ok {
					prefType = t
				}
				if c, ok := result.Metadatas[0][i]["confidence"].(string); ok {
					confidence = c
				}
			}
			pref := ExtractedPreference{
				Type:  prefType,
				Value: result.Documents[0][i],
			}
			if confidence != "" {
				fmt.Sscanf(confidence, "%f", &pref.Confidence)
			}
			prefs = append(prefs, pref)
		}
	}
	log.Printf("Recalled %d preferences for user %s", len(prefs), userID)
	return prefs, nil
}