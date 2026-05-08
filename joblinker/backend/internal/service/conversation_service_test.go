package service

import (
	"strings"
	"testing"
)

func TestExtractLocationValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantNot string // must NOT equal this
	}{
		{
			name:    "remote work in San Francisco",
			input:   "I'm looking for remote work in San Francisco",
			wantNot: "extracted_from_conversation",
		},
		{
			name:    "onsite in NYC",
			input:   "I prefer an onsite position in NYC",
			wantNot: "extracted_from_conversation",
		},
		{
			name:    "hybrid position",
			input:   "I'm interested in a hybrid position",
			wantNot: "extracted_from_conversation",
		},
		{
			name:    "random gibberish",
			input:   "xyzzy plugh not a real location at all",
			wantNot: "extracted_from_conversation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractLocationValue(tt.input)
			if result == tt.wantNot {
				t.Errorf("extractLocationValue(%q) = %q, must not be placeholder", tt.input, result)
			}
			if result == "" {
				t.Errorf("extractLocationValue(%q) returned empty string", tt.input)
			}
		})
	}
}

func TestExtractLocationValue_ContainsKnownKeywords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string // at least one of these should appear in result
	}{
		{
			name:     "remote keyword",
			input:    "Looking for remote work",
			expected: []string{"remote", "Remote"},
		},
		{
			name:     "onsite keyword",
			input:    "Prefer onsite in NYC",
			expected: []string{"onsite", "Onsite", "NYC", "New York"},
		},
		{
			name:     "hybrid keyword",
			input:    "Hybrid position preferred",
			expected: []string{"hybrid", "Hybrid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractLocationValue(tt.input)
			found := false
			for _, keyword := range tt.expected {
				if strings.Contains(result, keyword) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("extractLocationValue(%q) = %q, expected one of %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractSkillsValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantNot string
	}{
		{
			name:    "Go and Python",
			input:   "I have 5 years of Go and Python experience",
			wantNot: "extracted_from_conversation",
		},
		{
			name:    "React Node Docker",
			input:   "My skills include React, Node.js, and Docker",
			wantNot: "extracted_from_conversation",
		},
		{
			name:    "random text no skills",
			input:   "xyzzy plugh not a real skill at all",
			wantNot: "extracted_from_conversation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSkillsValue(tt.input)
			if result == tt.wantNot {
				t.Errorf("extractSkillsValue(%q) = %q, must not be placeholder", tt.input, result)
			}
			if result == "" {
				t.Errorf("extractSkillsValue(%q) returned empty string", tt.input)
			}
		})
	}
}

func TestExtractSkillsValue_ContainsKnownSkills(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Go and Python",
			input:    "5 years of Go and Python",
			expected: []string{"Go", "Python"},
		},
		{
			name:     "React Node Docker",
			input:    "React, Node.js, Docker experience",
			expected: []string{"React", "Node", "Docker"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSkillsValue(tt.input)
			for _, skill := range tt.expected {
				if !strings.Contains(result, skill) {
					t.Errorf("extractSkillsValue(%q) = %q, missing %q", tt.input, result, skill)
				}
			}
		})
	}
}

func TestExtractSalaryValue_Fallback(t *testing.T) {
	// When no salary info is present, should NOT return placeholder
	input := "Hello, I'm interested in this position"
	result := extractSalaryValue(input)
	if result == "extracted_from_conversation" {
		t.Errorf("extractSalaryValue(%q) returned placeholder, must return substring of input", input)
	}
	if result == "" {
		t.Errorf("extractSalaryValue(%q) returned empty string", input)
	}
}
