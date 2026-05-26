package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"joblinker/internal/adapters"
	"joblinker/internal/core"

	"github.com/google/uuid"
)

// ToolNode executes AI tool calls against core domain services.
type ToolNode struct {
	jobRepo       core.JobRepository
	agentRepo     core.AgentRepository
	matchRepo     core.MatchRepository
	offerRepo     core.OfferRepository
	interviewRepo core.InterviewRepository
	allowedTools  []string
}

// NewToolNode creates a new ToolNode.
func NewToolNode(
	jobRepo core.JobRepository,
	agentRepo core.AgentRepository,
	matchRepo core.MatchRepository,
	offerRepo core.OfferRepository,
	interviewRepo core.InterviewRepository,
) *ToolNode {
	return &ToolNode{
		jobRepo:       jobRepo,
		agentRepo:     agentRepo,
		matchRepo:     matchRepo,
		offerRepo:     offerRepo,
		interviewRepo: interviewRepo,
	}
}

// SetAllowedTools sets which tools are permitted (empty = all allowed).
func (n *ToolNode) SetAllowedTools(tools []string) {
	n.allowedTools = tools
}

// Process executes the requested tool and returns the result.
func (n *ToolNode) Process(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	toolName, _ := input["tool"].(string)
	if toolName == "" {
		return nil, fmt.Errorf("tool name required")
	}

	args, _ := input["arguments"].(map[string]interface{})
	if args == nil {
		args = make(map[string]interface{})
	}

	matchIDStr, _ := input["match_id"].(string)

	if !n.canUseTool(toolName) {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("tool '%s' is not permitted", toolName),
		}, nil
	}

	var result *ToolExecutionResult
	var err error

	switch toolName {
	case "query_jobs":
		result, err = n.executeQueryJobs(ctx, args)
	case "get_candidate":
		result, err = n.executeGetCandidate(ctx, args)
	case "create_offer":
		result, err = n.executeCreateOffer(ctx, args)
	case "schedule_interview":
		result, err = n.executeScheduleInterview(ctx, args)
	case "search_candidates":
		result, err = n.executeSearchCandidates(ctx, args)
	default:
		err = fmt.Errorf("unknown tool: %s", toolName)
	}

	if matchIDStr != "" {
		matchID, parseErr := uuid.Parse(matchIDStr)
		if parseErr == nil {
			n.recordToolCall(matchID, toolName, args, result, err)
		}
	}

	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"success": result.Success,
		"data":    result.Data,
		"error":   result.Error,
	}, nil
}

// GetTools returns all tool definitions for AI function calling.
func (n *ToolNode) GetTools() []adapters.Tool {
	return []adapters.Tool{
		{
			Type: "function",
			Function: adapters.ToolFunction{
				Name:        "query_jobs",
				Description: "Search for jobs based on location, skills, salary range, and job type",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location":  map[string]interface{}{"type": "string", "description": "Desired work location"},
						"skills":    map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
						"salary_min": map[string]interface{}{"type": "number"},
						"job_type":  map[string]interface{}{"type": "string"},
						"limit":     map[string]interface{}{"type": "number"},
					},
				},
			},
		},
		{
			Type: "function",
			Function: adapters.ToolFunction{
				Name:        "get_candidate",
				Description: "Get detailed candidate profile by ID",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"candidate_id": map[string]interface{}{"type": "string"},
					},
					"required": []string{"candidate_id"},
				},
			},
		},
		{
			Type: "function",
			Function: adapters.ToolFunction{
				Name:        "create_offer",
				Description: "Create a job offer for a candidate",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"match_id":   map[string]interface{}{"type": "string"},
						"salary":     map[string]interface{}{"type": "number"},
						"start_date": map[string]interface{}{"type": "string"},
						"notes":      map[string]interface{}{"type": "string"},
					},
					"required": []string{"match_id", "salary"},
				},
			},
		},
		{
			Type: "function",
			Function: adapters.ToolFunction{
				Name:        "schedule_interview",
				Description: "Schedule an interview between candidate and recruiter",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"match_id":       map[string]interface{}{"type": "string"},
						"datetime":       map[string]interface{}{"type": "string"},
						"interview_type": map[string]interface{}{"type": "string"},
					},
					"required": []string{"match_id", "datetime"},
				},
			},
		},
		{
			Type: "function",
			Function: adapters.ToolFunction{
				Name:        "search_candidates",
				Description: "Search for candidates matching job requirements",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"skills":         map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
						"location":       map[string]interface{}{"type": "string"},
						"experience_min": map[string]interface{}{"type": "number"},
						"limit":          map[string]interface{}{"type": "number"},
					},
					"required": []string{"skills"},
				},
			},
		},
	}
}

func (n *ToolNode) canUseTool(name string) bool {
	if len(n.allowedTools) == 0 {
		return true
	}
	for _, t := range n.allowedTools {
		if t == name {
			return true
		}
	}
	return false
}

func (n *ToolNode) executeQueryJobs(ctx context.Context, args map[string]interface{}) (*ToolExecutionResult, error) {
	location, _ := args["location"].(string)
	skillsRaw, _ := args["skills"].([]interface{})
	salaryMin, _ := args["salary_min"].(float64)
	jobType, _ := args["job_type"].(string)
	limit, _ := args["limit"].(float64)

	var skills []string
	for _, s := range skillsRaw {
		if sk, ok := s.(string); ok {
			skills = append(skills, sk)
		}
	}
	if limit == 0 {
		limit = 10
	}

	var jobs []*core.Job
	var err error
	if n.jobRepo != nil {
		jobs, err = n.jobRepo.Search(ctx, "", skills, location, int(salaryMin), jobType, int(limit))
		if err != nil {
			log.Printf("JobRepository.Search error: %v", err)
			jobs = []*core.Job{}
		}
	} else {
		jobs = []*core.Job{}
	}

	jobList := make([]map[string]interface{}, len(jobs))
	for i, job := range jobs {
		var structuredData map[string]interface{}
		if len(job.StructuredJSON) > 0 {
			json.Unmarshal(job.StructuredJSON, &structuredData)
		}
		title := ""
		jobLocation := ""
		salary := 0
		jobTypeStr := ""
		if structuredData != nil {
			if t, ok := structuredData["title"].(string); ok {
				title = t
			}
			if loc, ok := structuredData["location"].(string); ok {
				jobLocation = loc
			}
			if sal, ok := structuredData["salary"].(float64); ok {
				salary = int(sal)
			}
			if jt, ok := structuredData["job_type"].(string); ok {
				jobTypeStr = jt
			}
		}
		jobList[i] = map[string]interface{}{
			"id":       job.ID.String(),
			"title":    title,
			"location": jobLocation,
			"salary":   salary,
			"skills":   skills,
			"job_type": jobTypeStr,
		}
	}

	return &ToolExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"jobs":  jobList,
			"count": len(jobList),
		},
	}, nil
}

func (n *ToolNode) executeGetCandidate(ctx context.Context, args map[string]interface{}) (*ToolExecutionResult, error) {
	candidateIDStr, _ := args["candidate_id"].(string)
	if candidateIDStr == "" {
		return nil, fmt.Errorf("candidate_id required")
	}

	candidateID, err := uuid.Parse(candidateIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid candidate_id: %w", err)
	}

	if n.agentRepo == nil {
		return nil, fmt.Errorf("agent repository not configured")
	}

	agent, err := n.agentRepo.GetByID(candidateID)
	if err != nil {
		return &ToolExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("failed to get candidate: %v", err),
		}, nil
	}

	var configData map[string]interface{}
	if len(agent.ConfigJSON) > 0 {
		json.Unmarshal(agent.ConfigJSON, &configData)
	}
	name := ""
	if agent.User != nil {
		name = agent.User.Email
	}
	skills := []string{}
	experience := 0
	if configData != nil {
		if s, ok := configData["skills"].([]interface{}); ok {
			for _, sk := range s {
				if skStr, ok := sk.(string); ok {
					skills = append(skills, skStr)
				}
			}
		}
		if exp, ok := configData["experience_years"].(float64); ok {
			experience = int(exp)
		}
	}

	return &ToolExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"id":         agent.ID.String(),
			"name":       name,
			"skills":     skills,
			"experience": experience,
		},
	}, nil
}

func (n *ToolNode) executeCreateOffer(ctx context.Context, args map[string]interface{}) (*ToolExecutionResult, error) {
	matchIDStr, _ := args["match_id"].(string)
	salary, _ := args["salary"].(float64)
	currency, _ := args["currency"].(string)
	startDateStr, _ := args["start_date"].(string)

	if matchIDStr == "" {
		return nil, fmt.Errorf("match_id required")
	}

	matchID, err := uuid.Parse(matchIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid match_id: %w", err)
	}

	var match *core.Match
	if n.matchRepo != nil {
		match, _ = n.matchRepo.GetByID(matchID)
	}

	if salary == 0 && match != nil && match.Job != nil {
		var jobData map[string]interface{}
		if jsonErr := json.Unmarshal(match.Job.StructuredJSON, &jobData); jsonErr == nil {
			if min, ok := jobData["salary_min"].(float64); ok && min > 0 {
				if max, ok := jobData["salary_max"].(float64); ok && max > min {
					salary = (min + max) / 2
				} else {
					salary = min
				}
			}
			if currency == "" {
				if c, ok := jobData["currency"].(string); ok {
					currency = c
				}
			}
		}
	}
	if salary == 0 {
		salary = 150000
	}
	if currency == "" {
		currency = "USD"
	}

	var startDate time.Time
	if startDateStr != "" {
		startDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			startDate, err = time.Parse("2006-01-02", startDateStr)
		}
		if err != nil {
			startDate = time.Now().AddDate(0, 1, 0)
		}
	} else {
		startDate = time.Now().AddDate(0, 1, 0)
	}

	compensation := map[string]interface{}{
		"base_salary": salary,
		"currency":    currency,
		"bonus": map[string]interface{}{
			"amount":      15000,
			"description": "Annual performance bonus",
		},
		"equity": map[string]interface{}{
			"shares":        5000,
			"vesting_period": "4 years",
		},
		"benefits": []string{"Health insurance", "401k matching", "Unlimited PTO"},
	}
	compensationJSON, _ := json.Marshal(compensation)

	offer := &core.Offer{
		ID:               uuid.New(),
		MatchID:          matchID,
		CompensationJSON: json.RawMessage(compensationJSON),
		StartDate:        startDate,
		Status:           core.OfferStatusPending,
	}
	if match != nil {
		offer.TenantID = match.TenantID
	}

	if n.offerRepo != nil {
		if err := n.offerRepo.Create(offer); err != nil {
			return &ToolExecutionResult{
				Success: false,
				Error:   fmt.Sprintf("failed to create offer: %v", err),
			}, nil
		}
	}

	return &ToolExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"offer_id": offer.ID.String(),
			"status":   "created",
		},
	}, nil
}

func (n *ToolNode) executeScheduleInterview(ctx context.Context, args map[string]interface{}) (*ToolExecutionResult, error) {
	matchIDStr, _ := args["match_id"].(string)
	datetimeStr, _ := args["datetime"].(string)
	interviewType, _ := args["interview_type"].(string)

	if matchIDStr == "" {
		return nil, fmt.Errorf("match_id required")
	}
	if datetimeStr == "" {
		return nil, fmt.Errorf("datetime required")
	}

	datetime, err := time.Parse(time.RFC3339, datetimeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid datetime format: %w", err)
	}

	matchID, err := uuid.Parse(matchIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid match_id: %w", err)
	}

	var format core.InterviewFormat
	switch interviewType {
	case "video":
		format = core.InterviewFormatVideo
	case "phone":
		format = core.InterviewFormatPhone
	case "onsite":
		format = core.InterviewFormatOnsite
	default:
		format = core.InterviewFormatVideo
	}

	interview := &core.Interview{
		ID:          uuid.New(),
		MatchID:     matchID,
		ScheduledAt: datetime,
		Format:      format,
		Status:      core.InterviewStatusScheduled,
	}

	if n.matchRepo != nil {
		match, err := n.matchRepo.GetByID(matchID)
		if err == nil {
			interview.TenantID = match.TenantID
		}
	}

	if n.interviewRepo != nil {
		if err := n.interviewRepo.Create(interview); err != nil {
			return &ToolExecutionResult{
				Success: false,
				Error:   fmt.Sprintf("failed to create interview: %v", err),
			}, nil
		}
	}

	return &ToolExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"interview_id": interview.ID.String(),
			"status":       "scheduled",
			"datetime":     datetimeStr,
		},
	}, nil
}

func (n *ToolNode) executeSearchCandidates(ctx context.Context, args map[string]interface{}) (*ToolExecutionResult, error) {
	skillsRaw, _ := args["skills"].([]interface{})
	location, _ := args["location"].(string)
	experienceMin, _ := args["experience_min"].(float64)
	limit, _ := args["limit"].(float64)

	var skills []string
	for _, s := range skillsRaw {
		if sk, ok := s.(string); ok {
			skills = append(skills, sk)
		}
	}
	if limit == 0 {
		limit = 10
	}

	var agentList []*core.Agent
	var err error
	if n.agentRepo != nil {
		agentList, err = n.agentRepo.SearchBySkills(ctx, skills, location, int(experienceMin), int(limit))
		if err != nil {
			log.Printf("AgentRepository.SearchBySkills error: %v", err)
			agentList = []*core.Agent{}
		}
	} else {
		agentList = []*core.Agent{}
	}

	candidates := make([]map[string]interface{}, len(agentList))
	for i, agent := range agentList {
		var configData map[string]interface{}
		if len(agent.ConfigJSON) > 0 {
			json.Unmarshal(agent.ConfigJSON, &configData)
		}
		name := ""
		if agent.User != nil {
			name = agent.User.Email
		}
		agentSkills := []string{}
		if configData != nil {
			if s, ok := configData["skills"].([]interface{}); ok {
				for _, sk := range s {
					if skStr, ok := sk.(string); ok {
						agentSkills = append(agentSkills, skStr)
					}
				}
			}
		}
		candidates[i] = map[string]interface{}{
			"id":         agent.ID.String(),
			"name":       name,
			"skills":     agentSkills,
			"location":   location,
			"experience": int(experienceMin),
		}
	}

	return &ToolExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"candidates": candidates,
			"count":      len(candidates),
		},
	}, nil
}

func (n *ToolNode) recordToolCall(matchID uuid.UUID, toolName string, args map[string]interface{}, result *ToolExecutionResult, err error) {
	argumentsJSON, _ := json.Marshal(args)

	status := core.ToolStatusSuccess
	if err != nil || (result != nil && !result.Success) {
		status = core.ToolStatusFailed
	}

	toolCall := &core.AgentToolCall{
		ID:        uuid.New(),
		MatchID:   matchID,
		ToolName:  toolName,
		Arguments: string(argumentsJSON),
		Status:    status,
	}
	if result != nil && result.Data != nil {
		resultJSON, _ := json.Marshal(result.Data)
		toolCall.Result = string(resultJSON)
	}

	if n.matchRepo != nil {
		if dbErr := n.matchRepo.CreateToolCall(context.Background(), toolCall); dbErr != nil {
			log.Printf("Failed to persist tool call: %v", dbErr)
		} else {
			log.Printf("Tool call recorded: %s.%s - %s", matchID, toolName, status)
		}
	}
}

// ToolExecutionResult holds the result of a tool execution.
type ToolExecutionResult struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// ExecuteToolForAI executes a tool and returns JSON string for AI function calling.
func (n *ToolNode) ExecuteToolForAI(toolName string, args map[string]interface{}) (string, error) {
	result, err := n.Process(context.Background(), map[string]interface{}{
		"tool":      toolName,
		"arguments": args,
	})
	if err != nil {
		return "", err
	}
	data, _ := json.Marshal(result["data"])
	return string(data), nil
}
