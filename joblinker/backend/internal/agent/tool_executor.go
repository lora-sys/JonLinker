package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"joblinker/internal/cache"
	"joblinker/internal/config"
	"joblinker/internal/model"
	"joblinker/internal/repository"
	"joblinker/pkg/ai"

	"github.com/google/uuid"
)

// ToolExecutor handles agent tool/function execution
type ToolExecutor struct {
	jobRepo       *repository.JobRepository
	agentRepo     *repository.AgentRepository
	matchRepo     *repository.MatchRepository
	offerRepo     *repository.OfferRepository
	interviewRepo *repository.InterviewRepository
	toolCache     *cache.ToolCache
	allowedTools  *config.AgentToolConfig
	// Observability
	onError func(errorType, msg string, ctx map[string]interface{})
}

func NewToolExecutor(
	jobRepo *repository.JobRepository,
	agentRepo *repository.AgentRepository,
	matchRepo *repository.MatchRepository,
	offerRepo *repository.OfferRepository,
	interviewRepo *repository.InterviewRepository,
) *ToolExecutor {
	return &ToolExecutor{
		jobRepo:       jobRepo,
		agentRepo:     agentRepo,
		matchRepo:     matchRepo,
		offerRepo:     offerRepo,
		interviewRepo: interviewRepo,
	}
}

// SetCache sets the tool cache for result caching
func (e *ToolExecutor) SetCache(cache *cache.ToolCache) {
	e.toolCache = cache
}

// SetAllowedTools sets the allowed tools configuration
func (e *ToolExecutor) SetAllowedTools(cfg *config.AgentToolConfig) {
	e.allowedTools = cfg
}

// SetAllowedToolsFromJSON sets allowed tools from JSON config
func (e *ToolExecutor) SetAllowedToolsFromJSON(configJSON string) error {
	cfg, err := config.ParseAgentToolConfig(configJSON)
	if err != nil {
		return err
	}
	e.allowedTools = cfg
	return nil
}

// SetErrorHandler sets a callback for error tracking
func (e *ToolExecutor) SetErrorHandler(onError func(errorType, msg string, ctx map[string]interface{})) {
	e.onError = onError
}

// canUseTool checks if the tool is allowed for this executor
func (e *ToolExecutor) canUseTool(toolName string) bool {
	if e.allowedTools == nil {
		return true
	}
	return e.allowedTools.CanUseTool(toolName)
}

// ExecuteTool executes a tool call and records it
func (e *ToolExecutor) ExecuteTool(ctx context.Context, matchID uuid.UUID, toolName string, arguments map[string]interface{}) (*ToolExecutionResult, error) {
	log.Printf("Executing tool: %s for match %s", toolName, matchID)

	// Check tool permission
	if !e.canUseTool(toolName) {
		log.Printf("Tool %s is not allowed for this agent", toolName)
		return &ToolExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("tool '%s' is not permitted", toolName),
		}, fmt.Errorf("tool '%s' is not permitted", toolName)
	}

	var result *ToolExecutionResult
	var err error

	switch toolName {
	case "query_jobs":
		result, err = e.executeQueryJobs(ctx, arguments)
	case "get_candidate":
		result, err = e.executeGetCandidate(ctx, arguments)
	case "create_offer":
		result, err = e.executeCreateOffer(ctx, arguments)
	case "schedule_interview":
		result, err = e.executeScheduleInterview(ctx, arguments)
	case "search_candidates":
		result, err = e.executeSearchCandidates(ctx, arguments)
	default:
		err = fmt.Errorf("unknown tool: %s", toolName)
	}

	// Cache result if caching is enabled and execution was successful
	if e.toolCache != nil && result != nil && result.Success && result.Data != nil {
		cacheKey := e.toolCache.GenerateCacheKey(toolName, arguments)
		cached := e.toolCache.Set(cacheKey, result.Data)
		log.Printf("Cached tool result: key=%s, summary=%s", cacheKey, cached.Summary)
	}

	// Record tool call
	e.recordToolCall(matchID, toolName, arguments, result, err)

	return result, err
}

// ExecuteToolWithCache executes tool and returns cache key + summary instead of full result
func (e *ToolExecutor) ExecuteToolWithCache(ctx context.Context, matchID uuid.UUID, toolName string, arguments map[string]interface{}) (*CachedToolResult, error) {
	log.Printf("Executing tool with cache: %s for match %s", toolName, matchID)

	// Check tool permission
	if !e.canUseTool(toolName) {
		log.Printf("Tool %s is not allowed for this agent", toolName)
		return &CachedToolResult{
			Success: false,
			Error:   fmt.Sprintf("tool '%s' is not permitted", toolName),
		}, fmt.Errorf("tool '%s' is not permitted", toolName)
	}

	// Generate cache key first
	var cacheKey string
	if e.toolCache != nil {
		cacheKey = e.toolCache.GenerateCacheKey(toolName, arguments)
		// Check if already cached
		if cached, found := e.toolCache.Get(cacheKey); found {
			log.Printf("Cache hit for %s", cacheKey)
			return &CachedToolResult{
				Success:    true,
				CacheKey:   cacheKey,
				Summary:    cached.Summary,
				CacheHit:   true,
			}, nil
		}
	}

	var result *ToolExecutionResult
	var err error

	switch toolName {
	case "query_jobs":
		result, err = e.executeQueryJobs(ctx, arguments)
	case "get_candidate":
		result, err = e.executeGetCandidate(ctx, arguments)
	case "create_offer":
		result, err = e.executeCreateOffer(ctx, arguments)
	case "schedule_interview":
		result, err = e.executeScheduleInterview(ctx, arguments)
	case "search_candidates":
		result, err = e.executeSearchCandidates(ctx, arguments)
	default:
		err = fmt.Errorf("unknown tool: %s", toolName)
	}

	// Cache result
	var summary string
	if e.toolCache != nil && result != nil && result.Success && result.Data != nil {
		cached := e.toolCache.Set(cacheKey, result.Data)
		summary = cached.Summary
		log.Printf("Cached tool result: key=%s, summary=%s", cacheKey, summary)
	}

	// Record tool call
	e.recordToolCall(matchID, toolName, arguments, result, err)

	return &CachedToolResult{
		Success:  result != nil && result.Success,
		CacheKey: cacheKey,
		Summary:  summary,
		Error:    errStr(err, result),
	}, err
}

// CachedToolResult holds a cached tool execution result with key + summary
type CachedToolResult struct {
	Success  bool
	CacheKey string
	Summary  string
	CacheHit bool
	Error    string
}

// ToolExecutionResult holds the result of a tool execution
type ToolExecutionResult struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

func (e *ToolExecutor) executeQueryJobs(ctx context.Context, args map[string]interface{}) (*ToolExecutionResult, error) {
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

	var jobList []*model.Job
	var err error

	if e.jobRepo != nil {
		jobList, err = e.jobRepo.Search(ctx, "", skills, location, int(salaryMin), jobType, int(limit))
		if err != nil {
			log.Printf("JobRepository.Search error: %v", err)
			jobList = []*model.Job{}
		}
	} else {
		jobList = []*model.Job{}
	}

	// Convert to map format for tool result
	jobs := make([]map[string]interface{}, len(jobList))
	for i, job := range jobList {
		// Parse StructuredJSON to extract job details
		var structuredData map[string]interface{}
		if job.StructuredJSON != "" {
			json.Unmarshal([]byte(job.StructuredJSON), &structuredData)
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
		jobs[i] = map[string]interface{}{
			"id":        job.ID.String(),
			"title":     title,
			"location":  jobLocation,
			"salary":    salary,
			"skills":    skills,
			"job_type": jobTypeStr,
		}
	}

	return &ToolExecutionResult{
		Success: true,
		Data: map[string]interface{}{
			"jobs":  jobs,
			"count": len(jobs),
		},
	}, nil
}

func (e *ToolExecutor) executeGetCandidate(ctx context.Context, args map[string]interface{}) (*ToolExecutionResult, error) {
	candidateIDStr, _ := args["candidate_id"].(string)
	if candidateIDStr == "" {
		return nil, fmt.Errorf("candidate_id required")
	}

	candidateID, err := uuid.Parse(candidateIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid candidate_id: %w", err)
	}

	var candidate map[string]interface{}

	if e.agentRepo != nil {
		agent, err := e.agentRepo.GetByID(candidateID)
		if err != nil {
			return &ToolExecutionResult{
				Success: false,
				Error:   fmt.Sprintf("failed to get candidate: %v", err),
			}, nil
		}
		// Parse ConfigJSON to extract candidate details
		var configData map[string]interface{}
		if agent.ConfigJSON != "" {
			json.Unmarshal([]byte(agent.ConfigJSON), &configData)
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
		candidate = map[string]interface{}{
			"id":         agent.ID.String(),
			"name":       name,
			"skills":     skills,
			"experience": experience,
		}
	} else {
		return nil, fmt.Errorf("agent repository not configured")
	}

	return &ToolExecutionResult{
		Success: true,
		Data:    candidate,
	}, nil
}

func (e *ToolExecutor) executeCreateOffer(ctx context.Context, args map[string]interface{}) (*ToolExecutionResult, error) {
	matchIDStr, _ := args["match_id"].(string)
	salary, _ := args["salary"].(float64)
	startDateStr, _ := args["start_date"].(string)
	notes, _ := args["notes"].(string)

	if matchIDStr == "" {
		return nil, fmt.Errorf("match_id required")
	}
	if salary == 0 {
		return nil, fmt.Errorf("salary required")
	}

	// Parse match ID
	matchID, err := uuid.Parse(matchIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid match_id: %w", err)
	}

	// Parse start date - accept both YYYY-MM-DD and RFC3339 formats
	var startDate time.Time
	if startDateStr != "" {
		// Try RFC3339 first (full datetime), then date only
		startDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			startDate, err = time.Parse("2006-01-02", startDateStr)
		}
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format: %w", err)
		}
	}

	// Create compensation JSON
	compensation := map[string]interface{}{
		"base_salary": salary,
		"notes":       notes,
	}
	compensationJSON, _ := json.Marshal(compensation)

	// Create offer record
	offer := &model.Offer{
		ID:               uuid.New(),
		MatchID:          matchID,
		CompensationJSON: string(compensationJSON),
		StartDate:        startDate,
		Status:           model.OfferStatusPending,
	}

	if e.offerRepo != nil {
		if err := e.offerRepo.Create(offer); err != nil {
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

func (e *ToolExecutor) executeScheduleInterview(ctx context.Context, args map[string]interface{}) (*ToolExecutionResult, error) {
	matchIDStr, _ := args["match_id"].(string)
	datetimeStr, _ := args["datetime"].(string)
	_, _ = args["duration_minutes"].(float64) // duration tracked separately if needed
	interviewType, _ := args["interview_type"].(string)

	if matchIDStr == "" {
		return nil, fmt.Errorf("match_id required")
	}
	if datetimeStr == "" {
		return nil, fmt.Errorf("datetime required")
	}

	// Parse datetime
	datetime, err := time.Parse(time.RFC3339, datetimeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid datetime format: %w", err)
	}

	// Parse match ID
	matchID, err := uuid.Parse(matchIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid match_id: %w", err)
	}

	// Map interview type string to InterviewFormat
	var format model.InterviewFormat
	switch interviewType {
	case "video":
		format = model.InterviewFormatVideo
	case "phone":
		format = model.InterviewFormatPhone
	case "onsite":
		format = model.InterviewFormatOnsite
	default:
		format = model.InterviewFormatVideo
	}

	// Create interview record
	interview := &model.Interview{
		ID:          uuid.New(),
		MatchID:     matchID,
		ScheduledAt: datetime,
		Format:      format,
		Status:      model.InterviewStatusScheduled,
	}

	if e.interviewRepo != nil {
		if err := e.interviewRepo.Create(interview); err != nil {
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

func (e *ToolExecutor) executeSearchCandidates(ctx context.Context, args map[string]interface{}) (*ToolExecutionResult, error) {
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

	var candidateList []*model.Agent
	var err error

	if e.agentRepo != nil {
		candidateList, err = e.agentRepo.SearchBySkills(ctx, skills, location, int(experienceMin), int(limit))
		if err != nil {
			log.Printf("AgentRepository.SearchBySkills error: %v", err)
			candidateList = []*model.Agent{}
		}
	} else {
		candidateList = []*model.Agent{}
	}

	// Convert to map format for tool result
	candidates := make([]map[string]interface{}, len(candidateList))
	for i, agent := range candidateList {
		// Parse ConfigJSON to extract candidate details
		var configData map[string]interface{}
		if agent.ConfigJSON != "" {
			json.Unmarshal([]byte(agent.ConfigJSON), &configData)
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

func (e *ToolExecutor) recordToolCall(matchID uuid.UUID, toolName string, args map[string]interface{}, result *ToolExecutionResult, err error) {
	argumentsJSON, _ := json.Marshal(args)

	status := model.ToolStatusSuccess
	if err != nil || (result != nil && !result.Success) {
		status = model.ToolStatusFailed
	}

	toolCall := &model.AgentToolCall{
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

	// Persist to database
	if e.matchRepo != nil {
		if dbErr := e.matchRepo.CreateToolCall(context.Background(), toolCall); dbErr != nil {
			log.Printf("Failed to persist tool call: %v", dbErr)
		} else {
			log.Printf("Tool call recorded: %s.%s - %s", matchID, toolName, status)
		}
	} else {
		log.Printf("Tool call recorded (no DB): %s.%s - %s", matchID, toolName, status)
	}

	// Observability: Track errors
	if status == model.ToolStatusFailed && e.onError != nil {
		errMsg := "tool_execution_failed"
		if err != nil {
			errMsg = err.Error()
		}
		e.onError("tool_error", errMsg, map[string]interface{}{
			"tool_name": toolName,
			"match_id":  matchID.String(),
			"arguments": args,
		})
	}
}

func errStr(err error, result *ToolExecutionResult) string {
	if err != nil {
		return err.Error()
	}
	if result != nil && result.Error != "" {
		return result.Error
	}
	return ""
}

// ToolCallParams holds parameters for tool execution
type ToolCallParams struct {
	MatchID   uuid.UUID
	ToolName  string
	Arguments map[string]interface{}
}

// GetTools returns all tools as AI Tool definitions for ChatWithTools
func (e *ToolExecutor) GetTools() []ai.Tool {
	return []ai.Tool{
		{
			Type: "function",
			Function: ai.ToolFunction{
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
			Function: ai.ToolFunction{
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
			Function: ai.ToolFunction{
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
			Function: ai.ToolFunction{
				Name:        "schedule_interview",
				Description: "Schedule an interview between candidate and recruiter",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"match_id":        map[string]interface{}{"type": "string"},
						"datetime":       map[string]interface{}{"type": "string"},
						"interview_type":  map[string]interface{}{"type": "string"},
					},
					"required": []string{"match_id", "datetime"},
				},
			},
		},
		{
			Type: "function",
			Function: ai.ToolFunction{
				Name:        "search_candidates",
				Description: "Search for candidates matching job requirements",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"skills":        map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
						"location":      map[string]interface{}{"type": "string"},
						"experience_min": map[string]interface{}{"type": "number"},
						"limit":         map[string]interface{}{"type": "number"},
					},
					"required": []string{"skills"},
				},
			},
		},
	}
}

// ExecuteToolForAI executes a tool and returns JSON string for AI function calling
func (e *ToolExecutor) ExecuteToolForAI(toolName string, args map[string]interface{}) (string, error) {
	result, err := e.ExecuteTool(context.Background(), uuid.Nil, toolName, args)
	if err != nil {
		return "", err
	}
	if !result.Success {
		return "", fmt.Errorf("%s", result.Error)
	}
	data, _ := json.Marshal(result.Data)
	return string(data), nil
}