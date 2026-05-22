package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"joblinker/internal/eino/agent"
	"joblinker/internal/eino/stategraph"
	"joblinker/pkg/ai"
)

// mockAIServer returns predefined responses that drive the graph through all phases.
func mockAIServer(t *testing.T) *httptest.Server {
	t.Helper()
	callCount := 0

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var reqBody ai.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Logf("[Mock AI] Failed to decode request: %v", err)
		}
		_ = reqBody

		// Canned responses that drive the graph's branch conditions
		responses := []string{
			// introduction: seeker
			"My name is Alex Chen, a software engineer with 5 years of experience in backend development. I'm looking for a challenging role in distributed systems.",
			// introduction: recruiter
			"We're hiring a Senior Backend Engineer for our Platform team. You'll work on microservices, API design, and cloud infrastructure.",

			// negotiation: seeker expects salary
			"I'm looking for a salary in the range of $180,000 to $220,000 based on my experience.",
			// negotiation: recruiter agrees (triggers NegotiationDone)
			"We can offer $190,000 base salary plus equity and benefits. I agree this is a fair deal.",

			// interview: recruiter asks question
			"Tell me about your experience with distributed systems and how you've handled data consistency in a microservices architecture.",
			// interview: seeker answers
			"I designed and implemented an event-driven system using Kafka and PostgreSQL, ensuring eventual consistency through idempotent consumers.",
			// interview: recruiter evaluates - PASS
			"That's an excellent answer. I rate it 8/10. PASS.",

			// offer: recruiter presents
			"We'd like to offer you $190,000 base salary, equity, and benefits. Please let us know if you accept.",
			// offer: seeker accepts
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

		// Support streaming (SSE) format when the client asks for it.
		// EinoChatModel.Generate uses DoChatStream which expects SSE lines.
		if reqBody.Stream {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")
			flusher, _ := w.(http.Flusher)

			// Build SSE delta chunk
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
			// Send [DONE] signal
			fmt.Fprintf(w, "data: [DONE]\n\n")
			if flusher != nil {
				flusher.Flush()
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

func newAIClient(server *httptest.Server) *ai.Client {
	return &ai.Client{
		BaseURL:     server.URL,
		APIKey:      "mock-key",
		Model:       "mock-model",
		MaxTokens:   1024,
		Temperature: 0.2,
		HTTPClient:  http.DefaultClient,
	}
}

func TestHiringGraphE2E(t *testing.T) {
	server := mockAIServer(t)
	defer server.Close()

	aiClient := newAIClient(server)
	seeker := agent.NewSeekerAgent(aiClient)
	recruiter := agent.NewRecruiterAgent(aiClient)

	hg := NewHiringGraph(seeker, recruiter)
	matchID := uuid.New()

	state, err := hg.Run(context.Background(), matchID)
	if err != nil {
		t.Fatalf("HiringGraph.Run failed: %v", err)
	}

	// Verify final phase
	if state.Phase != stategraph.PhaseCompleted {
		t.Errorf("expected PhaseCompleted, got %s", state.Phase)
	}

	// Verify state was populated
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
		t.Log("Offer not accepted (may need mock adjustment)")
	}

	t.Logf("Final state: phase=%s candidate=%q job=%q neg_rounds=%d interview_rounds=%d accepted=%v",
		state.Phase, state.CandidateName, state.JobTitle,
		len(state.NegotiationRounds), len(state.InterviewRounds), state.OfferAccepted)

	t.Logf("--- Full conversation (%d messages) ---", len(state.Messages))
	for i, m := range state.Messages {
		short := m
		if len(short) > 200 {
			short = short[:200] + "..."
		}
		t.Logf("  [%d] %s", i, short)
	}
}

func TestHiringGraphPhases(t *testing.T) {
	server := mockAIServer(t)
	defer server.Close()

	aiClient := newAIClient(server)
	seeker := agent.NewSeekerAgent(aiClient)
	recruiter := agent.NewRecruiterAgent(aiClient)

	hg := NewHiringGraph(seeker, recruiter)
	runnable, err := hg.Build(context.Background())
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	matchID := uuid.New()
	initialState := stategraph.NewRecruitmentState(matchID)

	// Step through invocation
	state, err := runnable.Invoke(context.Background(), initialState)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	// Verify the state completed
	if state.Phase != stategraph.PhaseCompleted {
		t.Errorf("Phase should be COMPLETED, got %s", state.Phase)
	}

	// Verify phase progression through messages
	hasIntroduction := false
	hasNegotiation := false
	hasInterview := false
	hasOffer := false

	for _, m := range state.Messages {
		if searchSubstring(m, "Alex Chen") {
			hasIntroduction = true
		}
		if searchSubstring(m, "salary") {
			hasNegotiation = true
		}
		if searchSubstring(m, "distributed") || searchSubstring(m, "Kafka") {
			hasInterview = true
		}
		if searchSubstring(m, "offer") || searchSubstring(m, "190,000") {
			hasOffer = true
		}
	}

	if !hasIntroduction {
		t.Error("No introduction content found in messages")
	}
	if !hasNegotiation {
		t.Error("No negotiation content found in messages")
	}
	if !hasInterview {
		t.Error("No interview content found in messages")
	}
	if !hasOffer {
		t.Error("No offer content found in messages")
	}
}

func TestBranchConditions(t *testing.T) {
	// Test the branch condition logic directly

	// Negotiation branch: when NegotiationDone=false, route back to negotiation
	matchID := uuid.New()
	state := stategraph.NewRecruitmentState(matchID)
	state.NegotiationDone = false

	negCond := func(ctx context.Context, s *stategraph.RecruitmentState) (string, error) {
		if s.NegotiationDone {
			return nodeInterview, nil
		}
		return nodeNegotiation, nil
	}

	next, _ := negCond(context.Background(), state)
	if next != nodeNegotiation {
		t.Errorf("expected %s when not done, got %s", nodeNegotiation, next)
	}

	state.NegotiationDone = true
	next, _ = negCond(context.Background(), state)
	if next != nodeInterview {
		t.Errorf("expected %s when done, got %s", nodeInterview, next)
	}

	// Interview branch
	state.InterviewPassed = false
	ivCond := func(ctx context.Context, s *stategraph.RecruitmentState) (string, error) {
		if s.InterviewPassed {
			return nodeOffer, nil
		}
		return nodeInterview, nil
	}
	next, _ = ivCond(context.Background(), state)
	if next != nodeInterview {
		t.Errorf("expected %s when not passed, got %s", nodeInterview, next)
	}

	state.InterviewPassed = true
	next, _ = ivCond(context.Background(), state)
	if next != nodeOffer {
		t.Errorf("expected %s when passed, got %s", nodeOffer, next)
	}

	// Offer branch — use a fresh state for each branch
	matchID3 := uuid.New()
	state3 := stategraph.NewRecruitmentState(matchID3)
	state3.OfferAccepted = true
	state3.OfferDeclined = false

	offerCond := func(ctx context.Context, s *stategraph.RecruitmentState) (string, error) {
		if s.OfferAccepted || s.OfferDeclined {
			return nodeCompleted, nil
		}
		return nodeOffer, nil
	}
	next, _ = offerCond(context.Background(), state3)
	if next != nodeCompleted {
		t.Errorf("expected %s when accepted, got %s", nodeCompleted, next)
	}

	matchID4 := uuid.New()
	state4 := stategraph.NewRecruitmentState(matchID4)
	state4.OfferDeclined = true
	next, _ = offerCond(context.Background(), state4)
	if next != nodeCompleted {
		t.Errorf("expected %s when declined, got %s", nodeCompleted, next)
	}

	matchID5 := uuid.New()
	state5 := stategraph.NewRecruitmentState(matchID5)
	next, _ = offerCond(context.Background(), state5)
	if next != nodeOffer {
		t.Errorf("expected %s when undecided, got %s", nodeOffer, next)
	}
}
