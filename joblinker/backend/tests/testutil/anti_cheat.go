// Package testutil contains anti-cheat utilities to detect hardcoded fake data.
package testutil

import (
	"strings"
	"testing"
)

// KnownHardcodedValues are hardcoded strings that must NEVER appear in test outputs.
// Detection = the code being tested is returning fake/hardcoded data.
var KnownHardcodedValues = []string{
	"Sample Job",
	"Sample Job Title",
	"Mock Job",
	"Sample Candidate",
	"Mock Candidate",
	"Thank you for your message",
	"Thank you for your introduction",
	"I am a professional AI recruitment agent",
	"I am a professional recruiter",
	"2026-06-01T10:00:00Z", // hardcoded interview time
	"2026-07-01",            // hardcoded offer start_date
	"2026-04-23T00:00:00Z", // hardcoded timestamp
	"extracted_from_conversation",
	"Seeker proposal in round",
	"Based on my analysis:",
}

// KnownHardcodedSalaries are salary values that indicate hardcoded fake data.
var KnownHardcodedSalaries = []int{150000, 120000}

// KnownHardcodedScore is the hardcoded match score (0.9).
const KnownHardcodedScore = 0.9

// AssertNotHardcoded checks that a string field does not contain known hardcoded values.
// If detected, the test FAILS with a clear error message.
func AssertNotHardcoded(t *testing.T, fieldName, value string) {
	for _, fake := range KnownHardcodedValues {
		if strings.Contains(value, fake) {
			t.Fatalf("FAKE DATA DETECTED: %q field contains known hardcoded value %q", fieldName, fake)
		}
	}
}

// AssertSalaryNotHardcoded checks that a salary value is not a known hardcoded value.
func AssertSalaryNotHardcoded(t *testing.T, fieldName string, salary int) {
	for _, fake := range KnownHardcodedSalaries {
		if salary == fake {
			t.Fatalf("FAKE DATA DETECTED: %s is hardcoded value %d, not from Job.salary_min/max", fieldName, fake)
		}
	}
}

// AssertScoreNotHardcoded checks that a match score is not the hardcoded 0.9.
func AssertScoreNotHardcoded(t *testing.T, fieldName string, score float64) {
	if score == KnownHardcodedScore {
		t.Fatalf("FAKE DATA DETECTED: %s is hardcoded %.1f, not from AI evaluation", fieldName, score)
	}
}

// AssertTimestampNotHardcoded checks that a timestamp is not the known hardcoded value.
func AssertTimestampNotHardcoded(t *testing.T, fieldName, timestamp string) {
	if timestamp == "2026-04-23T00:00:00Z" {
		t.Fatalf("FAKE DATA DETECTED: %s is hardcoded %q, not time.Now()", fieldName, timestamp)
	}
}

// AssertNotFallback checks that an AI response is not a generic fallback message.
func AssertNotFallback(t *testing.T, fieldName, response string) {
	// These phrases indicate generic fallback responses, not real AI generation
	fallbackPhrases := []string{
		"Thank you for your message",
		"I am a professional AI recruitment agent",
		"I am a professional recruiter",
		"This is an automated response",
		"I am an AI assistant",
	}
	for _, phrase := range fallbackPhrases {
		if strings.Contains(response, phrase) {
			t.Fatalf("FAKE AI RESPONSE: %q contains fallback phrase %q", fieldName, phrase)
		}
	}
}

// AssertNotHardcodedJSON is a convenience wrapper that checks a map field.
func AssertNotHardcodedJSON(t *testing.T, data map[string]interface{}, field string) {
	if val, ok := data[field].(string); ok && val != "" {
		AssertNotHardcoded(t, field, val)
	}
}
