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

	"github.com/google/uuid"
)

// ToolExecutor handles agent tool/function execution
type ToolExecutor struct {
	jobRepo       *repository.JobRepository
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
	matchRepo *repository.MatchRepository,
	offerRepo *repository.OfferRepository,
	interviewRepo *repository.InterviewRepository,
) *ToolExecutor {
	return &ToolExecutor{
		jobRepo:       jobRepo,
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
	log.Printf("Executing tool: %s with args: %v", toolName, arguments)

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
	log.Printf("Executing tool with cache: %s with args: %v", toolName, arguments)

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

	// Use job repository to search (would need to implement Search method)
	jobs := []map[string]interface{}{
		{
			"id":       uuid.New().String(),
			"title":    "Sample Job",
			"location": location,
			"salary":   int(salaryMin),
			"skills":   skills,
			"job_type": jobType,
		},
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
	candidateID, _ := args["candidate_id"].(string)
	if candidateID == "" {
		return nil, fmt.Errorf("candidate_id required")
	}

	// Would fetch from candidate repository
	candidate := map[string]interface{}{
		"id":       candidateID,
		"name":     "Sample Candidate",
		"skills":   []string{"golang", "python"},
		"experience": 5,
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

	// Parse start date
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format: %w", err)
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

	// Would use vector similarity search in production
	candidates := []map[string]interface{}{
		{
			"id":         uuid.New().String(),
			"skills":     skills,
			"location":   location,
			"experience": int(experienceMin),
		},
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

	// Record to repository (would save to DB)
	log.Printf("Tool call recorded: %s.%s - %s", matchID, toolName, status)

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