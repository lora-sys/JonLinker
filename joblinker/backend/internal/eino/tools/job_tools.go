package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"joblinker/internal/model"
	"joblinker/internal/repository"
)

// QueryJobsInput represents input for the query_jobs tool
type QueryJobsInput struct {
	Location string   `json:"location"`
	Skills   []string `json:"skills,omitempty"`
	SalaryMin int    `json:"salary_min,omitempty"`
	JobType  string   `json:"job_type,omitempty"`
	Limit    int      `json:"limit,omitempty"`
}

// QueryJobsOutput represents output from the query_jobs tool
type QueryJobsOutput struct {
	Jobs []JobInfo `json:"jobs"`
}

// JobInfo represents a job listing
type JobInfo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	SalaryRange string `json:"salary_range"`
	Description string `json:"description"`
}

// SearchCandidatesInput represents input for the search_candidates tool
type SearchCandidatesInput struct {
	Skills        []string `json:"skills"`
	Location      string   `json:"location,omitempty"`
	ExperienceMin int      `json:"experience_min,omitempty"`
	Limit         int      `json:"limit,omitempty"`
}

// SearchCandidatesOutput represents output from search_candidates
type SearchCandidatesOutput struct {
	Candidates []CandidateInfo `json:"candidates"`
}

// CandidateInfo represents a candidate profile
type CandidateInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Skills      []string `json:"skills"`
	Experience  int      `json:"experience"`
	Location    string   `json:"location"`
}

// GetCandidateInput represents input for get_candidate tool
type GetCandidateInput struct {
	CandidateID string `json:"candidate_id"`
}

// GetCandidateOutput represents output from get_candidate
type GetCandidateOutput struct {
	Candidate CandidateDetail `json:"candidate"`
}

// CandidateDetail represents detailed candidate info
type CandidateDetail struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Email        string   `json:"email"`
	Skills       []string `json:"skills"`
	Experience   int      `json:"experience"`
	Education    string   `json:"education"`
	Location     string   `json:"location"`
	SalaryExpect int      `json:"salary_expectation"`
}

// CreateOfferInput represents input for create_offer tool
type CreateOfferInput struct {
	MatchID  string `json:"match_id"`
	Salary   int    `json:"salary"`
	StartDate string `json:"start_date,omitempty"`
	Notes    string `json:"notes,omitempty"`
}

// CreateOfferOutput represents output from create_offer
type CreateOfferOutput struct {
	OfferID string `json:"offer_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ScheduleInterviewInput represents input for schedule_interview tool
type ScheduleInterviewInput struct {
	MatchID          string `json:"match_id"`
	Datetime         string `json:"datetime"`
	DurationMinutes  int    `json:"duration_minutes,omitempty"`
	InterviewType    string `json:"interview_type,omitempty"`
}

// ScheduleInterviewOutput represents output from schedule_interview
type ScheduleInterviewOutput struct {
	InterviewID string `json:"interview_id"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

// QueryJobs creates the query_jobs Eino tool
func QueryJobs(executor func(ctx context.Context, input QueryJobsInput) (QueryJobsOutput, error)) tool.InvokableTool {
	desc := &schema.ToolInfo{
		Name: "query_jobs",
		Desc: "Search for jobs based on location, skills, salary range, and job type. Returns matching job listings with details.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"location": {Type: schema.String, Desc: "Desired work location (city, region, or 'remote')", Required: true},
			"skills": {Type: schema.Array, ElemInfo: &schema.ParameterInfo{Type: schema.String}, Desc: "Required skills for the job"},
			"salary_min": {Type: schema.Integer, Desc: "Minimum annual salary in USD"},
			"job_type": {Type: schema.String, Enum: []string{"remote", "hybrid", "onsite"}},
			"limit": {Type: schema.Integer, Desc: "Maximum number of results to return"},
		}),
	}
	return utils.NewTool(desc, func(ctx context.Context, input QueryJobsInput) (string, error) {
		output, err := executor(ctx, input)
		if err != nil {
			return "", err
		}
		data, err := json.Marshal(output)
		if err != nil {
			return "", fmt.Errorf("failed to marshal output: %w", err)
		}
		return string(data), nil
	})
}

// SearchCandidates creates the search_candidates Eino tool
func SearchCandidates(executor func(ctx context.Context, input SearchCandidatesInput) (SearchCandidatesOutput, error)) tool.InvokableTool {
	desc := &schema.ToolInfo{
		Name: "search_candidates",
		Desc: "Find candidates matching job requirements using vector similarity search.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"skills": {Type: schema.Array, ElemInfo: &schema.ParameterInfo{Type: schema.String}, Desc: "Required skills", Required: true},
			"location": {Type: schema.String, Desc: "Preferred location"},
			"experience_min": {Type: schema.Integer, Desc: "Minimum years of experience"},
			"limit": {Type: schema.Integer, Desc: "Maximum candidates to return"},
		}),
	}
	return utils.NewTool(desc, func(ctx context.Context, input SearchCandidatesInput) (string, error) {
		output, err := executor(ctx, input)
		if err != nil {
			return "", err
		}
		data, err := json.Marshal(output)
		if err != nil {
			return "", fmt.Errorf("failed to marshal output: %w", err)
		}
		return string(data), nil
	})
}

// GetCandidate creates the get_candidate Eino tool
func GetCandidate(executor func(ctx context.Context, input GetCandidateInput) (GetCandidateOutput, error)) tool.InvokableTool {
	desc := &schema.ToolInfo{
		Name: "get_candidate",
		Desc: "Get detailed candidate profile including skills, experience, and preferences.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"candidate_id": {Type: schema.String, Desc: "Unique candidate identifier", Required: true},
		}),
	}
	return utils.NewTool(desc, func(ctx context.Context, input GetCandidateInput) (string, error) {
		output, err := executor(ctx, input)
		if err != nil {
			return "", err
		}
		data, err := json.Marshal(output)
		if err != nil {
			return "", fmt.Errorf("failed to marshal output: %w", err)
		}
		return string(data), nil
	})
}

// CreateOffer creates the create_offer Eino tool
func CreateOffer(executor func(ctx context.Context, input CreateOfferInput) (CreateOfferOutput, error)) tool.InvokableTool {
	desc := &schema.ToolInfo{
		Name: "create_offer",
		Desc: "Create a job offer for a candidate match. Sends offer details to candidate.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"match_id": {Type: schema.String, Desc: "The match identifier linking candidate and job", Required: true},
			"salary": {Type: schema.Integer, Desc: "Annual salary offer in USD", Required: true},
			"start_date": {Type: schema.String, Desc: "Proposed start date (YYYY-MM-DD)"},
			"notes": {Type: schema.String, Desc: "Additional notes for the candidate"},
		}),
	}
	return utils.NewTool(desc, func(ctx context.Context, input CreateOfferInput) (string, error) {
		output, err := executor(ctx, input)
		if err != nil {
			return "", err
		}
		data, err := json.Marshal(output)
		if err != nil {
			return "", fmt.Errorf("failed to marshal output: %w", err)
		}
		return string(data), nil
	})
}

// ScheduleInterview creates the schedule_interview Eino tool
func ScheduleInterview(executor func(ctx context.Context, input ScheduleInterviewInput) (ScheduleInterviewOutput, error)) tool.InvokableTool {
	desc := &schema.ToolInfo{
		Name: "schedule_interview",
		Desc: "Schedule an interview between the candidate and recruiter.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"match_id": {Type: schema.String, Desc: "The match identifier", Required: true},
			"datetime": {Type: schema.String, Desc: "Interview date and time (ISO 8601 format)", Required: true},
			"duration_minutes": {Type: schema.Integer, Desc: "Interview duration in minutes"},
			"interview_type": {Type: schema.String, Enum: []string{"video", "phone", "onsite"}},
		}),
	}
	return utils.NewTool(desc, func(ctx context.Context, input ScheduleInterviewInput) (string, error) {
		output, err := executor(ctx, input)
		if err != nil {
			return "", err
		}
		data, err := json.Marshal(output)
		if err != nil {
			return "", fmt.Errorf("failed to marshal output: %w", err)
		}
		return string(data), nil
	})
}

// GetAllTools returns all Eino tools
func GetAllTools() []tool.InvokableTool {
	tools := make([]tool.InvokableTool, 0)

	// QueryJobs - placeholder executor that returns mock data
	qj := QueryJobs(func(ctx context.Context, input QueryJobsInput) (QueryJobsOutput, error) {
		return QueryJobsOutput{
			Jobs: []JobInfo{
				{ID: "job-1", Title: "Senior Go Developer", Company: "TechCorp", Location: input.Location, SalaryRange: "$120k-$150k"},
			},
		}, nil
	})
	tools = append(tools, qj)

	// SearchCandidates
	sc := SearchCandidates(func(ctx context.Context, input SearchCandidatesInput) (SearchCandidatesOutput, error) {
		return SearchCandidatesOutput{
			Candidates: []CandidateInfo{
				{ID: "cand-1", Name: "John Doe", Skills: input.Skills, Experience: 5, Location: input.Location},
			},
		}, nil
	})
	tools = append(tools, sc)

	// GetCandidate
	gc := GetCandidate(func(ctx context.Context, input GetCandidateInput) (GetCandidateOutput, error) {
		return GetCandidateOutput{
			Candidate: CandidateDetail{
				ID: input.CandidateID, Name: "John Doe", Email: "john@example.com",
				Skills: []string{"Go", "Python"}, Experience: 5, Location: "NYC",
			},
		}, nil
	})
	tools = append(tools, gc)

	// CreateOffer
	co := CreateOffer(func(ctx context.Context, input CreateOfferInput) (CreateOfferOutput, error) {
		return CreateOfferOutput{
			OfferID: "offer-123", Status: "sent", Message: "Offer created successfully",
		}, nil
	})
	tools = append(tools, co)

	// ScheduleInterview
	si := ScheduleInterview(func(ctx context.Context, input ScheduleInterviewInput) (ScheduleInterviewOutput, error) {
		return ScheduleInterviewOutput{
			InterviewID: "interview-123", Status: "scheduled", Message: "Interview scheduled",
		}, nil
	})
	tools = append(tools, si)

	return tools
}

// NewRealTools creates production tools backed by real repository data
func NewRealTools(jobRepo *repository.JobRepository, agentRepo *repository.AgentRepository, matchRepo *repository.MatchRepository, offerRepo *repository.OfferRepository, interviewRepo *repository.InterviewRepository) []tool.InvokableTool {
	tools := make([]tool.InvokableTool, 0)

	qj := QueryJobs(func(ctx context.Context, input QueryJobsInput) (QueryJobsOutput, error) {
		jobs, err := jobRepo.Search(ctx, "", input.Skills, input.Location, input.SalaryMin, "", input.Limit)
		if err != nil {
			return QueryJobsOutput{Jobs: []JobInfo{}}, nil
		}
		var results []JobInfo
		for _, j := range jobs {
			var info JobInfo
			info.ID = j.ID.String()
			json.Unmarshal(j.StructuredJSON, &info)
			results = append(results, info)
		}
		return QueryJobsOutput{Jobs: results}, nil
	})
	tools = append(tools, qj)

	sc := SearchCandidates(func(ctx context.Context, input SearchCandidatesInput) (SearchCandidatesOutput, error) {
		agents, err := agentRepo.SearchBySkills(ctx, input.Skills, input.Location, input.ExperienceMin, input.Limit)
		if err != nil {
			return SearchCandidatesOutput{Candidates: []CandidateInfo{}}, nil
		}
		var results []CandidateInfo
		for _, a := range agents {
			results = append(results, CandidateInfo{
				ID: a.ID.String(), Name: a.ID.String(),
				Experience: 0, Location: "",
			})
		}
		return SearchCandidatesOutput{Candidates: results}, nil
	})
	tools = append(tools, sc)

	gc := GetCandidate(func(ctx context.Context, input GetCandidateInput) (GetCandidateOutput, error) {
		agentID, err := uuid.Parse(input.CandidateID)
		if err != nil {
			return GetCandidateOutput{}, fmt.Errorf("invalid candidate id")
		}
		agent, err := agentRepo.GetByID(agentID)
		if err != nil {
			return GetCandidateOutput{}, fmt.Errorf("candidate not found")
		}
		return GetCandidateOutput{
			Candidate: CandidateDetail{
				ID: agent.ID.String(), Name: agent.ID.String(),
			},
		}, nil
	})
	tools = append(tools, gc)

	co := CreateOffer(func(ctx context.Context, input CreateOfferInput) (CreateOfferOutput, error) {
		matchID, err := uuid.Parse(input.MatchID)
		if err != nil {
			return CreateOfferOutput{}, fmt.Errorf("invalid match id")
		}
		offer := &model.Offer{
			ID:      uuid.New(),
			MatchID: matchID,
			Status:  model.OfferStatusPending,
		}
		if err := offerRepo.Create(offer); err != nil {
			return CreateOfferOutput{}, fmt.Errorf("failed to create offer: %w", err)
		}
		return CreateOfferOutput{
			OfferID: offer.ID.String(), Status: string(offer.Status), Message: "Offer created",
		}, nil
	})
	tools = append(tools, co)

	si := ScheduleInterview(func(ctx context.Context, input ScheduleInterviewInput) (ScheduleInterviewOutput, error) {
		matchID, err := uuid.Parse(input.MatchID)
		if err != nil {
			return ScheduleInterviewOutput{}, fmt.Errorf("invalid match id")
		}
		scheduledAt, err := time.Parse(time.RFC3339, input.Datetime)
		if err != nil {
			scheduledAt = time.Now().Add(24 * time.Hour)
		}
		interview := &model.Interview{
			ID: uuid.New(), MatchID: matchID,
			ScheduledAt: scheduledAt,
			Status:      model.InterviewStatusScheduled,
		}
		if err := interviewRepo.Create(interview); err != nil {
			return ScheduleInterviewOutput{}, fmt.Errorf("failed to schedule interview: %w", err)
		}
		return ScheduleInterviewOutput{
			InterviewID: interview.ID.String(), Status: string(interview.Status), Message: "Interview scheduled",
		}, nil
	})
	tools = append(tools, si)

	return tools
}

// GetToolInfos returns schema.ToolInfo for all tools (for registration with ChatModel)
func GetToolInfos() []*schema.ToolInfo {
	tools := GetAllTools()
	infos := make([]*schema.ToolInfo, len(tools))
	for i, t := range tools {
		info, _ := t.Info(context.Background())
		infos[i] = info
	}
	return infos
}

// GetToolByName returns a tool by name
func GetToolByName(name string) tool.InvokableTool {
	switch name {
	case "query_jobs":
		t := QueryJobs(func(ctx context.Context, input QueryJobsInput) (QueryJobsOutput, error) {
			return QueryJobsOutput{Jobs: []JobInfo{{ID: "job-1", Title: "Mock Job"}}}, nil
		})
		return t
	case "search_candidates":
		t := SearchCandidates(func(ctx context.Context, input SearchCandidatesInput) (SearchCandidatesOutput, error) {
			return SearchCandidatesOutput{Candidates: []CandidateInfo{{ID: "cand-1", Name: "Mock"}}}, nil
		})
		return t
	case "get_candidate":
		t := GetCandidate(func(ctx context.Context, input GetCandidateInput) (GetCandidateOutput, error) {
			return GetCandidateOutput{Candidate: CandidateDetail{ID: input.CandidateID}}, nil
		})
		return t
	case "create_offer":
		t := CreateOffer(func(ctx context.Context, input CreateOfferInput) (CreateOfferOutput, error) {
			return CreateOfferOutput{OfferID: "offer-1", Status: "created"}, nil
		})
		return t
	case "schedule_interview":
		t := ScheduleInterview(func(ctx context.Context, input ScheduleInterviewInput) (ScheduleInterviewOutput, error) {
			return ScheduleInterviewOutput{InterviewID: "interview-1", Status: "scheduled"}, nil
		})
		return t
	default:
		return nil
	}
}

// ToolName constants
const (
	ToolNameQueryJobs        = "query_jobs"
	ToolNameSearchCandidates = "search_candidates"
	ToolNameGetCandidate     = "get_candidate"
	ToolNameCreateOffer      = "create_offer"
	ToolNameScheduleInterview = "schedule_interview"
)
