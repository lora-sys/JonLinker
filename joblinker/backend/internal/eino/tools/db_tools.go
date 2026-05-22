package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"

	"joblinker/internal/repository"
)

// GetMatchProgressInput represents input for the get_match_progress tool
type GetMatchProgressInput struct {
	MatchID string `json:"match_id"`
}

// GetMatchProgressOutput represents output from get_match_progress
type GetMatchProgressOutput struct {
	MatchID         string `json:"match_id"`
	Status          string `json:"status"`
	Score           float64 `json:"score"`
	SeekerAgentID   string `json:"seeker_agent_id"`
	RecruiterAgentID string `json:"recruiter_agent_id,omitempty"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// GetInterviewDetailsInput represents input for the get_interview_details tool
type GetInterviewDetailsInput struct {
	MatchID string `json:"match_id"`
}

// GetInterviewDetailsOutput represents output from get_interview_details
type GetInterviewDetailsOutput struct {
	InterviewID string `json:"interview_id"`
	ScheduledAt string `json:"scheduled_at"`
	Format      string `json:"format"`
	Status      string `json:"status"`
	Location    string `json:"location,omitempty"`
}

// GetOfferDetailsInput represents input for the get_offer_details tool
type GetOfferDetailsInput struct {
	MatchID string `json:"match_id"`
}

// GetOfferDetailsOutput represents output from get_offer_details
type GetOfferDetailsOutput struct {
	OfferID     string `json:"offer_id"`
	Salary      int    `json:"salary"`
	Currency    string `json:"currency"`
	StartDate   string `json:"start_date"`
	Status      string `json:"status"`
}

// GetUserProfileInput represents input for the get_user_profile tool
type GetUserProfileInput struct {
	AgentID string `json:"agent_id"`
}

// GetUserProfileOutput represents output from get_user_profile
type GetUserProfileOutput struct {
	AgentID    string   `json:"agent_id"`
	Type       string   `json:"type"`
	Email      string   `json:"email,omitempty"`
	Skills     []string `json:"skills,omitempty"`
	Experience int      `json:"experience,omitempty"`
}

// GetMatchProgress creates the get_match_progress Eino tool
func GetMatchProgress(matchRepo *repository.MatchRepository) tool.InvokableTool {
	desc := &schema.ToolInfo{
		Name: "get_match_progress",
		Desc: "Get the current progress, status, and metadata of a match between a seeker and recruiter.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"match_id": {Type: schema.String, Desc: "The match UUID to query", Required: true},
		}),
	}
	return utils.NewTool(desc, func(ctx context.Context, input GetMatchProgressInput) (string, error) {
		matchID, err := uuid.Parse(input.MatchID)
		if err != nil {
			return "", fmt.Errorf("invalid match_id: %w", err)
		}
		match, err := matchRepo.GetByID(matchID)
		if err != nil {
			return "", fmt.Errorf("match not found: %w", err)
		}
		output := GetMatchProgressOutput{
			MatchID:         match.ID.String(),
			Status:          string(match.Status),
			Score:           match.Score,
			SeekerAgentID:   match.SeekerAgentID.String(),
			CreatedAt:       match.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:       match.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if match.RecruiterAgentID != nil {
			output.RecruiterAgentID = match.RecruiterAgentID.String()
		}
		data, _ := json.Marshal(output)
		return string(data), nil
	})
}

// GetInterviewDetails creates the get_interview_details Eino tool
func GetInterviewDetails(interviewRepo *repository.InterviewRepository) tool.InvokableTool {
	desc := &schema.ToolInfo{
		Name: "get_interview_details",
		Desc: "Get interview details for a match, including scheduled time, format, and status.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"match_id": {Type: schema.String, Desc: "The match UUID", Required: true},
		}),
	}
	return utils.NewTool(desc, func(ctx context.Context, input GetInterviewDetailsInput) (string, error) {
		matchID, err := uuid.Parse(input.MatchID)
		if err != nil {
			return "", fmt.Errorf("invalid match_id: %w", err)
		}
		interview, err := interviewRepo.GetByMatchID(matchID)
		if err != nil {
			return "", fmt.Errorf("interview not found: %w", err)
		}
		output := GetInterviewDetailsOutput{
			InterviewID: interview.ID.String(),
			ScheduledAt: interview.ScheduledAt.Format("2006-01-02T15:04:05Z"),
			Format:      string(interview.Format),
			Status:      string(interview.Status),
			Location:    interview.Location,
		}
		data, _ := json.Marshal(output)
		return string(data), nil
	})
}

// GetOfferDetails creates the get_offer_details Eino tool
func GetOfferDetails(offerRepo *repository.OfferRepository) tool.InvokableTool {
	desc := &schema.ToolInfo{
		Name: "get_offer_details",
		Desc: "Get offer details for a match, including salary, currency, start date, and status.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"match_id": {Type: schema.String, Desc: "The match UUID", Required: true},
		}),
	}
	return utils.NewTool(desc, func(ctx context.Context, input GetOfferDetailsInput) (string, error) {
		matchID, err := uuid.Parse(input.MatchID)
		if err != nil {
			return "", fmt.Errorf("invalid match_id: %w", err)
		}
		offer, err := offerRepo.GetByMatchID(matchID)
		if err != nil {
			return "", fmt.Errorf("offer not found: %w", err)
		}
		salary := 0
		currency := "USD"
		var comp map[string]interface{}
		if len(offer.CompensationJSON) > 0 {
			json.Unmarshal(offer.CompensationJSON, &comp)
			if s, ok := comp["base_salary"].(float64); ok {
				salary = int(s)
			}
			if c, ok := comp["currency"].(string); ok {
				currency = c
			}
		}
		output := GetOfferDetailsOutput{
			OfferID:   offer.ID.String(),
			Salary:    salary,
			Currency:  currency,
			StartDate: offer.StartDate.Format("2006-01-02"),
			Status:    string(offer.Status),
		}
		data, _ := json.Marshal(output)
		return string(data), nil
	})
}

// GetUserProfile creates the get_user_profile Eino tool
func GetUserProfile(agentRepo *repository.AgentRepository) tool.InvokableTool {
	desc := &schema.ToolInfo{
		Name: "get_user_profile",
		Desc: "Get user/agent profile details including type, contact info, skills, and experience.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"agent_id": {Type: schema.String, Desc: "The agent UUID to look up", Required: true},
		}),
	}
	return utils.NewTool(desc, func(ctx context.Context, input GetUserProfileInput) (string, error) {
		agentID, err := uuid.Parse(input.AgentID)
		if err != nil {
			return "", fmt.Errorf("invalid agent_id: %w", err)
		}
		agent, err := agentRepo.GetByID(agentID)
		if err != nil {
			return "", fmt.Errorf("agent not found: %w", err)
		}
		output := GetUserProfileOutput{
			AgentID: agent.ID.String(),
			Type:    string(agent.Type),
		}
		if agent.User != nil {
			output.Email = agent.User.Email
		}
		var cfg map[string]interface{}
		if len(agent.ConfigJSON) > 0 {
			json.Unmarshal(agent.ConfigJSON, &cfg)
			if s, ok := cfg["skills"].([]interface{}); ok {
				for _, sk := range s {
					if skStr, ok := sk.(string); ok {
						output.Skills = append(output.Skills, skStr)
					}
				}
			}
			if exp, ok := cfg["experience_years"].(float64); ok {
				output.Experience = int(exp)
			}
		}
		data, _ := json.Marshal(output)
		return string(data), nil
	})
}

// AddDBTools appends the 4 custom DB tools to a tool slice
func AddDBTools(tools []tool.InvokableTool, matchRepo *repository.MatchRepository, interviewRepo *repository.InterviewRepository, offerRepo *repository.OfferRepository, agentRepo *repository.AgentRepository) []tool.InvokableTool {
	tools = append(tools, GetMatchProgress(matchRepo))
	tools = append(tools, GetInterviewDetails(interviewRepo))
	tools = append(tools, GetOfferDetails(offerRepo))
	tools = append(tools, GetUserProfile(agentRepo))
	return tools
}
