package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"joblinker/internal/eino/a2ui"
	"joblinker/internal/eino/hooks"
	"joblinker/internal/eino/runner"
	"joblinker/internal/repository"
)

type A2ASSEHandler struct {
	runner    *runner.ADKRunner
	matchRepo *repository.MatchRepository
	jobRepo   *repository.JobRepository
}

func NewA2ASSEHandler(runner *runner.ADKRunner) *A2ASSEHandler {
	return &A2ASSEHandler{runner: runner}
}

// WithMatchContext provides repos for loading match context data from the database.
// If set, the SSE handler will load candidate/job info and inject it into the ADK session.
func (h *A2ASSEHandler) WithMatchContext(matchRepo *repository.MatchRepository, jobRepo *repository.JobRepository) *A2ASSEHandler {
	h.matchRepo = matchRepo
	h.jobRepo = jobRepo
	return h
}

type a2aChatRequest struct {
	Message string `json:"message"`
}

func (h *A2ASSEHandler) Chat(c *gin.Context) {
	var req a2aChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message required"})
		return
	}

	sessionID := c.DefaultQuery("session_id", uuid.New().String())

	// ── Layer 2: Match context injection ──
	ctx := c.Request.Context()
	matchIDStr := c.Query("match_id")
	agentType := c.DefaultQuery("agent_type", "seeker")

	if matchIDStr != "" {
		ctx = hooks.WithMatchID(ctx, matchIDStr)
		ctx = hooks.WithAgentType(ctx, agentType)

		if matchID, err := uuid.Parse(matchIDStr); err == nil && h.matchRepo != nil {
			// Try to load match with full context
			match, err := h.matchRepo.GetByID(matchID)
			if err == nil {
				mc := &hooks.MatchContext{
					Phase:      c.Query("phase"),
					MatchScore: match.Score,
				}
				// Populate from match data (Agent has no Name field — use ID)
				if match.SeekerAgent != nil {
					mc.CandidateName = "Candidate #" + match.SeekerAgentID.String()[:8]
				}
				if match.RecruiterAgent != nil {
					mc.JobTitle = "Job #" + match.JobID.String()[:8]
				}
				// Try to extract job info from StructuredJSON
				if match.Job != nil && len(match.Job.StructuredJSON) > 0 {
					jobInfo := parseJobStructured(match.Job.StructuredJSON)
					if jobInfo.title != "" {
						mc.JobTitle = jobInfo.title
					}
					if jobInfo.desc != "" {
						mc.JobDesc = jobInfo.desc
					}
				}

				ctx = hooks.WithMatchContext(ctx, mc)
				log.Printf("[A2A SSE] Injected match context: match=%s title=%q phase=%s score=%.2f",
					matchIDStr, mc.JobTitle, mc.Phase, mc.MatchScore)
			} else {
				log.Printf("[A2A SSE] Match not found (id=%s): %v — continuing without context", matchIDStr, err)
			}
		}
	}

	// ── SSE setup ──
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.WriteHeader(http.StatusOK)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return
	}

	events := h.runner.Query(ctx, req.Message)

	if err := a2ui.StreamToWriter(ctx, c.Writer, events, sessionID); err != nil && err != io.EOF {
		log.Printf("A2A SSE stream error: %v", err)
		errData, _ := json.Marshal(map[string]string{"error": err.Error()})
		fmt.Fprintf(c.Writer, "data: %s\n\n", errData)
	}

	flusher.Flush()
}

// parseJobStructured extracts title and desc from Job StructuredJSON (json.RawMessage).
// The structured JSON typically contains fields like "title", "description", "skills", etc.
// Returns empty strings if parsing fails or the fields are absent.
type jobStructuredInfo struct {
	title string
	desc  string
}

func parseJobStructured(data []byte) jobStructuredInfo {
	var info struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(data, &info); err != nil {
		return jobStructuredInfo{}
	}
	return jobStructuredInfo{
		title: info.Title,
		desc:  info.Description,
	}
}
