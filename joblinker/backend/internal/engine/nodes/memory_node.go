package nodes

import (
	"context"
	"sync"
)

// MemoryNode manages conversation memory and preference extraction.
// Stores extracted preferences per match in-memory.
type MemoryNode struct {
	mu       sync.RWMutex
	memories map[string]*MatchMemory
}

// MatchMemory holds extracted preferences for a match conversation.
type MatchMemory struct {
	MatchID      string
	JobTitle     string
	Location     string
	SalaryMin    int
	SalaryMax    int
	Skills       []string
	AgentType    string
	KeyDecisions []string
}

// NewMemoryNode creates a new MemoryNode.
func NewMemoryNode() *MemoryNode {
	return &MemoryNode{
		memories: make(map[string]*MatchMemory),
	}
}

// Process extracts and stores preferences from conversation input.
// Input:
//   - match_id: conversation identifier
//   - message: latest message text
//   - agent_type: "recruiter" or "seeker"
//
// Output:
//   - preferences: extracted preference fields
func (n *MemoryNode) Process(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	matchID, _ := input["match_id"].(string)
	if matchID == "" {
		return nil, nil
	}

	message, _ := input["message"].(string)
	agentType, _ := input["agent_type"].(string)

	n.mu.Lock()
	mem, exists := n.memories[matchID]
	if !exists {
		mem = &MatchMemory{
			MatchID:   matchID,
			AgentType: agentType,
		}
		n.memories[matchID] = mem
	}
	n.mu.Unlock()

	// Update preferences if present in input
	if title, ok := input["job_title"].(string); ok && title != "" {
		mem.JobTitle = title
	}
	if loc, ok := input["location"].(string); ok && loc != "" {
		mem.Location = loc
	}
	if skills, ok := input["skills"].([]string); ok && len(skills) > 0 {
		mem.Skills = append(mem.Skills, skills...)
	}
	if salMin, ok := input["salary_min"].(int); ok && salMin > 0 {
		mem.SalaryMin = salMin
	}
	if salMax, ok := input["salary_max"].(int); ok && salMax > 0 {
		mem.SalaryMax = salMax
	}

	if message != "" {
		mem.KeyDecisions = append(mem.KeyDecisions, message)
	}

	return map[string]interface{}{
		"job_title":  mem.JobTitle,
		"location":   mem.Location,
		"salary_min": mem.SalaryMin,
		"salary_max": mem.SalaryMax,
		"skills":     mem.Skills,
		"decisions":  len(mem.KeyDecisions),
	}, nil
}

// GetMemory retrieves stored memory for a match.
func (n *MemoryNode) GetMemory(matchID string) *MatchMemory {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.memories[matchID]
}

// ClearMemory removes memory for a match.
func (n *MemoryNode) ClearMemory(matchID string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	delete(n.memories, matchID)
}
