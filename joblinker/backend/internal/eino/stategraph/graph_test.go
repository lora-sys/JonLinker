package stategraph

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"joblinker/internal/eino/agent"
	"joblinker/pkg/ai"
)

// mockAIServer returns predefined responses for each call. It cycles through
// canned responses so the graph can run through all phases.
func mockAIServer(t *testing.T) *httptest.Server {
	t.Helper()
	callCount := 0

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var reqBody ai.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Logf("[Mock AI] Failed to decode request: %v", err)
		}
		_ = reqBody // we ignore input and return canned response

		// Cycle through varied responses that the graph's detect functions
		// will interpret as progress
		responses := []string{
			// Ask the graph for introduction
			"My name is Alex Chen, a software engineer with 5 years of experience in backend development. I'm looking for a challenging role in distributed systems.",
			"We're hiring a Senior Backend Engineer for our Platform team. You'll work on microservices, API design, and cloud infrastructure.",

			// Negotiation round 1: seeker expects, recruiter counters
			"I'm looking for a salary in the range of $180,000 to $220,000 based on my experience.",
			"We can offer $190,000 base salary plus equity and benefits. I agree this is a fair deal.",

			// Interview round 1: recruiter asks, seeker answers, recruiter evaluates (PASS)
			"Tell me about your experience with distributed systems and how you've handled data consistency in a microservices architecture.",
			"I designed and implemented an event-driven system using Kafka and PostgreSQL, ensuring eventual consistency through idempotent consumers and exactly-once processing.",
			"That's an excellent answer. I rate it 8/10. PASS.",

			// Offer round 1: recruiter presents, seeker accepts
			"We'd like to offer you $190,000 base salary, equity, and benefits. Please let us know if you accept.",
			"I accept the offer! Thank you for this opportunity.",
		}

		idx := (callCount - 1) % len(responses)
		respText := responses[idx]

		resp := ai.ChatResponse{
			ID: fmt.Sprintf("mock-%d", callCount),
			Choices: []ai.Choice{
				{
					Message: ai.Message{
						Role:    "assistant",
						Content: respText,
					},
					Finish: "stop",
				},
			},
			Usage: ai.Usage{
				PromptTokens:     50,
				CompletionTokens: 100,
				TotalTokens:      150,
			},
		}

		// Support streaming (SSE) format — EinoChatModel.Generate uses DoChatStream
		if reqBody.Stream {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")
			flusher, _ := w.(http.Flusher)

			chunk := struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					Finish string `json:"finish_reason"`
				} `json:"choices"`
			}{
				Choices: []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					Finish string `json:"finish_reason"`
				}{
					{
						Delta: struct {
							Content string `json:"content"`
						}{Content: respText},
						Finish: "stop",
					},
				},
			}
			chunkData, _ := json.Marshal(chunk)
			fmt.Fprintf(w, "data: %s\n\n", chunkData)
			if flusher != nil {
				flusher.Flush()
			}
			fmt.Fprintf(w, "data: [DONE]\n\n")
			if flusher != nil {
				flusher.Flush()
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func newAIClientWithServer(server *httptest.Server) *ai.Client {
	return &ai.Client{
		BaseURL:     server.URL,
		APIKey:      "mock-key",
		Model:       "mock-model",
		MaxTokens:   1024,
		Temperature: 0.2,
		HTTPClient:  http.DefaultClient,
	}
}

func TestRecruitmentGraphE2E(t *testing.T) {
	mockServer := mockAIServer(t)
	defer mockServer.Close()

	aiClient := newAIClientWithServer(mockServer)

	seeker := agent.NewSeekerAgent(aiClient)
	recruiter := agent.NewRecruiterAgent(aiClient)

	graph := NewRecruitmentGraph(seeker, recruiter)

	collectedMsgs := make([]string, 0)
	phaseChanges := make([]Phase, 0)

	onMsg := func(sender, message string, phase Phase) {
		collectedMsgs = append(collectedMsgs, fmt.Sprintf("[%s] %s", sender, message))
	}
	onPhase := func(phase Phase) {
		phaseChanges = append(phaseChanges, phase)
	}

	matchID := uuid.New()
	state, err := graph.Run(context.Background(), matchID,
		WithCallbacks(onMsg, onPhase),
		WithMaxNegotiationRounds(3),
		WithMaxInterviewRounds(2),
		WithMaxOfferRounds(2),
	)
	if err != nil {
		t.Fatalf("Graph.Run failed: %v", err)
	}

	// Verify final phase is COMPLETED
	if state.Phase != PhaseCompleted {
		t.Errorf("expected PhaseCompleted, got %s", state.Phase)
	}

	// Verify phase progression
	expectedPhases := []Phase{PhaseIntroduction, PhaseNegotiation, PhaseInterview, PhaseOffer}
	if len(phaseChanges) < 4 {
		t.Errorf("expected at least 4 phase changes, got %d: %v", len(phaseChanges), phaseChanges)
	} else {
		for i, exp := range expectedPhases {
			if phaseChanges[i] != exp {
				t.Errorf("phase change %d: expected %s, got %s", i, exp, phaseChanges[i])
			}
		}
	}

	// Verify messages were collected via callbacks
	if len(collectedMsgs) == 0 {
		t.Error("no messages collected from callbacks")
	} else {
		t.Logf("Total messages from callbacks: %d", len(collectedMsgs))
	}

	// Verify state fields
	if state.CandidateName == "" {
		t.Error("CandidateName is empty")
	}
	if state.JobTitle == "" {
		t.Error("JobTitle is empty")
	}
	if len(state.NegotiationRounds) == 0 {
		t.Error("No negotiation rounds")
	}
	if len(state.InterviewRounds) == 0 {
		t.Error("No interview rounds")
	}
	if !state.OfferAccepted {
		t.Log("Offer was not accepted (depends on mock response content)")
	}

	t.Logf("Final state: phase=%s candidate=%q job=%q negotiated=%v interview_passed=%v offer_accepted=%v",
		state.Phase, state.CandidateName, state.JobTitle,
		len(state.NegotiationRounds), state.InterviewPassed, state.OfferAccepted)

	// Print all messages for visual inspection
	t.Logf("--- Full conversation (%d messages) ---", len(state.Messages))
	for i, m := range state.Messages {
		short := m
		if len(short) > 200 {
			short = short[:200] + "..."
		}
		t.Logf("  [%d] %s", i, short)
	}
}

func TestRecruitmentGraphNegotiationOnly(t *testing.T) {
	mockServer := mockAIServer(t)
	defer mockServer.Close()

	aiClient := newAIClientWithServer(mockServer)

	seeker := agent.NewSeekerAgent(aiClient)
	recruiter := agent.NewRecruiterAgent(aiClient)

	graph := NewRecruitmentGraph(seeker, recruiter)

	// Create state that's already past INTRODUCTION
	state := NewRecruitmentState(uuid.New())
	state.CandidateName = "Test Candidate"
	state.JobTitle = "Software Engineer"
	state.JobDesc = "A test position"
	state.Phase = PhaseNegotiation

	// Run one negotiation round
	state, err := graph.negotiationNode(context.Background(), state)
	if err != nil {
		t.Fatalf("negotiationNode failed: %v", err)
	}

	if len(state.NegotiationRounds) != 1 {
		t.Errorf("expected 1 negotiation round, got %d", len(state.NegotiationRounds))
	}
	if len(state.Messages) < 2 {
		t.Errorf("expected at least 2 messages, got %d", len(state.Messages))
	}

	t.Logf("Negotiation round: seeker=%s", truncate(state.NegotiationRounds[0].SeekerExpect, 100))
	t.Logf("  recruiter=%s", truncate(state.NegotiationRounds[0].RecruiterOffer, 100))
}

func TestParseMessage(t *testing.T) {
	tests := []struct {
		raw        string
		wantSender string
		wantText   string
	}{
		{"[Seeker] Hello, I'm John.", "Seeker", "Hello, I'm John."},
		{"[Recruiter] We have a position.", "Recruiter", "We have a position."},
		{"[System] Match complete.", "System", "Match complete."},
		{"plain message", "system", "plain message"},
	}
	for _, tt := range tests {
		gotSender, gotText := parseMessage(tt.raw)
		gotSender = strings.TrimSpace(gotSender)
		if gotSender != tt.wantSender {
			t.Errorf("parseMessage(%q) sender = %q, want %q", tt.raw, gotSender, tt.wantSender)
		}
		if gotText != tt.wantText {
			t.Errorf("parseMessage(%q) text = %q, want %q", tt.raw, gotText, tt.wantText)
		}
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
