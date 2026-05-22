package config

import (
	"encoding/json"
	"fmt"

	"joblinker/internal/model"
)

type ToolPermission struct {
	Enabled    bool   `json:"enabled"`
	RateLimit  int    `json:"rate_limit,omitempty"`
	MaxRetries int    `json:"max_retries,omitempty"`
}

type AgentToolConfig struct {
	Tools map[string]ToolPermission `json:"tools"`
	CacheBudget int                 `json:"cache_budget,omitempty"`
}

var DefaultSeekerTools = map[string]ToolPermission{
	"query_jobs":          {Enabled: true, RateLimit: 10, MaxRetries: 3},
	"get_candidate":       {Enabled: true, RateLimit: 20, MaxRetries: 3},
	"search_candidates":   {Enabled: false},
	"create_offer":        {Enabled: false},
	"schedule_interview":  {Enabled: false},
	"get_match_progress":  {Enabled: true, RateLimit: 10, MaxRetries: 3},
	"get_interview_details": {Enabled: true, RateLimit: 10, MaxRetries: 3},
	"get_offer_details":   {Enabled: true, RateLimit: 10, MaxRetries: 3},
	"get_user_profile":    {Enabled: true, RateLimit: 10, MaxRetries: 3},
}

var DefaultRecruiterTools = map[string]ToolPermission{
	"search_candidates":   {Enabled: true, RateLimit: 15, MaxRetries: 3},
	"get_candidate":       {Enabled: true, RateLimit: 20, MaxRetries: 3},
	"query_jobs":          {Enabled: false},
	"create_offer":        {Enabled: true, RateLimit: 5, MaxRetries: 2},
	"schedule_interview":  {Enabled: true, RateLimit: 5, MaxRetries: 2},
	"get_match_progress":  {Enabled: true, RateLimit: 10, MaxRetries: 3},
	"get_interview_details": {Enabled: true, RateLimit: 10, MaxRetries: 3},
	"get_offer_details":   {Enabled: true, RateLimit: 10, MaxRetries: 3},
	"get_user_profile":    {Enabled: true, RateLimit: 10, MaxRetries: 3},
}

func GetDefaultToolsForAgentType(agentType model.AgentType) map[string]ToolPermission {
	switch agentType {
	case model.AgentTypeSeeker:
		return DefaultSeekerTools
	case model.AgentTypeRecruiter:
		return DefaultRecruiterTools
	default:
		return DefaultSeekerTools
	}
}

func ParseAgentToolConfig(configJSON string) (*AgentToolConfig, error) {
	if configJSON == "" || configJSON == "null" {
		return &AgentToolConfig{
			Tools: GetDefaultToolsForAgentType(model.AgentTypeSeeker),
		}, nil
	}

	var cfg AgentToolConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse agent tool config: %w", err)
	}

	if cfg.Tools == nil {
		cfg.Tools = GetDefaultToolsForAgentType(model.AgentTypeSeeker)
	}

	return &cfg, nil
}

func (c *AgentToolConfig) CanUseTool(toolName string) bool {
	if perm, ok := c.Tools[toolName]; ok {
		return perm.Enabled
	}
	return false
}

func (c *AgentToolConfig) GetRateLimit(toolName string) int {
	if perm, ok := c.Tools[toolName]; ok {
		return perm.RateLimit
	}
	return 0
}

func (c *AgentToolConfig) GetMaxRetries(toolName string) int {
	if perm, ok := c.Tools[toolName]; ok {
		return perm.MaxRetries
	}
	return 0
}

func (c *AgentToolConfig) ToJSON() string {
	bytes, _ := json.Marshal(c)
	return string(bytes)
}

// IsToolAllowed checks if a tool is permitted for a given agent type.
func IsToolAllowed(toolName string, agentType model.AgentType) bool {
	tools := GetDefaultToolsForAgentType(agentType)
	perm, ok := tools[toolName]
	return ok && perm.Enabled
}

type ToolDefinition struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Parameters  []string `json:"parameters"`
}

var AllTools = []ToolDefinition{
	{Name: "query_jobs", Description: "Search for jobs based on location and skills", Parameters: []string{"location", "skills", "salary_min"}},
	{Name: "search_candidates", Description: "Search for candidates based on skills", Parameters: []string{"skills", "location"}},
	{Name: "get_candidate", Description: "Get candidate details by ID", Parameters: []string{"candidate_id"}},
	{Name: "create_offer", Description: "Create a job offer for a candidate", Parameters: []string{"match_id", "salary", "start_date"}},
	{Name: "schedule_interview", Description: "Schedule an interview", Parameters: []string{"match_id", "datetime", "interview_type"}},
}

func GetToolDefinition(toolName string) *ToolDefinition {
	for _, tool := range AllTools {
		if tool.Name == toolName {
			return &tool
		}
	}
	return nil
}