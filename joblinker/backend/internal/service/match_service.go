package service

import (
	"encoding/json"
	"errors"
	"strings"

	"joblinker/internal/model"
	"joblinker/internal/repository"

	"github.com/google/uuid"
)

var ErrInvalidMatchState = errors.New("invalid match state transition")

type MatchService struct {
	matchRepo *repository.MatchRepository
	agentRepo *repository.AgentRepository
	jobRepo   *repository.JobRepository
}

func NewMatchService(matchRepo *repository.MatchRepository, agentRepo *repository.AgentRepository, jobRepo *repository.JobRepository) *MatchService {
	return &MatchService{
		matchRepo: matchRepo,
		agentRepo: agentRepo,
		jobRepo:   jobRepo,
	}
}

func (s *MatchService) CreateMatch(seekerAgentID, jobID uuid.UUID, score float64) (*model.Match, error) {
	match := &model.Match{
		SeekerAgentID: seekerAgentID,
		JobID:         jobID,
		Score:         score,
		Status:        model.MatchStatusPending,
	}
	if err := s.matchRepo.Create(match); err != nil {
		return nil, err
	}
	return match, nil
}

func (s *MatchService) GetMatch(id uuid.UUID) (*model.Match, error) {
	return s.matchRepo.GetByID(id)
}

func (s *MatchService) ConfirmMatch(id uuid.UUID) (*model.Match, error) {
	match, err := s.matchRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if match.Status != model.MatchStatusPending {
		return nil, ErrInvalidMatchState
	}
	match.Status = model.MatchStatusMutualInterest
	if err := s.matchRepo.Update(match); err != nil {
		return nil, err
	}
	return match, nil
}

func (s *MatchService) TransitionToNegotiating(id uuid.UUID) (*model.Match, error) {
	match, err := s.matchRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if match.Status != model.MatchStatusMutualInterest {
		return nil, ErrInvalidMatchState
	}
	match.Status = model.MatchStatusNegotiating
	if err := s.matchRepo.Update(match); err != nil {
		return nil, err
	}
	return match, nil
}

func (s *MatchService) ListUserMatches(userID uuid.UUID) ([]*model.Match, error) {
	agents, err := s.agentRepo.ListByUserID(userID, "")
	if err != nil {
		return nil, err
	}
	var allMatches []*model.Match
	for _, agent := range agents {
		if agent.Type == model.AgentTypeSeeker {
			matches, err := s.matchRepo.ListBySeekerAgentID(agent.ID)
			if err != nil {
				continue
			}
			allMatches = append(allMatches, matches...)
		}
	}
	return allMatches, nil
}

func (s *MatchService) FindMatches(agentID uuid.UUID, limit int) ([]*model.Match, error) {
	return s.matchRepo.ListBySeekerAgentID(agentID)
}

func (s *MatchService) CalculateScore(seekerAgentID, jobID uuid.UUID) (float64, error) {
	seeker, err := s.agentRepo.GetByID(seekerAgentID)
	if err != nil {
		return 0, err
	}
	job, err := s.jobRepo.GetByID(jobID)
	if err != nil {
		return 0, err
	}

	// Parse seeker skills
	var seekerSkills []string
	var seekerExpYears int
	var seekerLocation string
	var config map[string]interface{}
	if json.Unmarshal([]byte(seeker.ConfigJSON), &config) == nil {
		if skills, ok := config["skills"].([]interface{}); ok {
			for _, skill := range skills {
				if s, ok := skill.(string); ok {
					seekerSkills = append(seekerSkills, s)
				}
			}
		}
		if exp, ok := config["experience_years"].(float64); ok {
			seekerExpYears = int(exp)
		}
		if loc, ok := config["location"].(string); ok {
			seekerLocation = loc
		}
	}

	// Parse job requirements
	var jobSkills []string
	var jobExpYears int
	var jobLocation string
	var jobData map[string]interface{}
	if json.Unmarshal([]byte(job.StructuredJSON), &jobData) == nil {
		if skills, ok := jobData["skills"].([]interface{}); ok {
			for _, skill := range skills {
				if s, ok := skill.(string); ok {
					jobSkills = append(jobSkills, s)
				}
			}
		}
		if exp, ok := jobData["experience_years"].(float64); ok {
			jobExpYears = int(exp)
		}
		if loc, ok := jobData["location"].(string); ok {
			jobLocation = loc
		}
	}

	// Calculate skills match (40% weight)
	skillsMatch := calculateSkillsMatch(seekerSkills, jobSkills)

	// Calculate location match (30% weight)
	locationMatch := calculateLocationMatch(seekerLocation, jobLocation)

	// Calculate experience match (30% weight)
	experienceMatch := calculateExperienceMatch(seekerExpYears, jobExpYears)

	// Weighted total score
	totalScore := (skillsMatch * 0.4) + (locationMatch * 0.3) + (experienceMatch * 0.3)

	return totalScore, nil
}

func calculateSkillsMatch(seekerSkills, jobSkills []string) float64 {
	if len(jobSkills) == 0 {
		return 1.0
	}
	matchCount := 0
	for _, seekerSkill := range seekerSkills {
		for _, jobSkill := range jobSkills {
			if strings.EqualFold(seekerSkill, jobSkill) {
				matchCount++
				break
			}
		}
	}
	return float64(matchCount) / float64(len(jobSkills))
}

func calculateLocationMatch(seekerLoc, jobLoc string) float64 {
	if jobLoc == "" || seekerLoc == "" {
		return 1.0
	}
	// Remote-friendly jobs match any location
	if strings.Contains(strings.ToLower(jobLoc), "remote") ||
		strings.Contains(strings.ToLower(jobLoc), "anywhere") {
		return 1.0
	}
	// Exact match
	if strings.EqualFold(seekerLoc, jobLoc) {
		return 1.0
	}
	// Same region (crude check)
	if strings.HasPrefix(seekerLoc, jobLoc[:min(len(seekerLoc), len(jobLoc))]) {
		return 0.7
	}
	return 0.3
}

func calculateExperienceMatch(seekerYears, jobYears int) float64 {
	if jobYears == 0 {
		return 1.0
	}
	// Within ±2 years of requirement = full match
	diff := seekerYears - jobYears
	if diff < 0 {
		diff = -diff
	}
	if diff <= 2 {
		return 1.0
	}
	if diff <= 4 {
		return 0.7
	}
	return 0.4
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type AutoMatchResult struct {
	Matches      []*model.Match `json:"matches"`
	CreatedCount int            `json:"created_count"`
	SkippedCount int            `json:"skipped_count"`
}

func (s *MatchService) AutoCreateMatches(userID uuid.UUID, jobIDs []uuid.UUID) (*AutoMatchResult, error) {
	agents, err := s.agentRepo.ListByUserID(userID, "")
	if err != nil {
		return nil, err
	}
	var seekerAgent *model.Agent
	for _, agent := range agents {
		if agent.Type == model.AgentTypeSeeker && agent.Status == model.AgentStatusActive {
			seekerAgent = agent
			break
		}
	}
	if seekerAgent == nil {
		return &AutoMatchResult{}, nil
	}

	result := &AutoMatchResult{}

	for _, jobID := range jobIDs {
		existing, _ := s.matchRepo.FindBySeekerAndJob(seekerAgent.ID, jobID)
		if existing != nil {
			result.SkippedCount++
			result.Matches = append(result.Matches, existing)
			continue
		}

		_, err := s.jobRepo.GetByID(jobID)
		if err != nil {
			result.SkippedCount++
			continue
		}

		// Use weighted hybrid score calculation
		score, err := s.CalculateScore(seekerAgent.ID, jobID)
		if err != nil {
			result.SkippedCount++
			continue
		}

		if score > 0.5 {
			match := &model.Match{
				SeekerAgentID: seekerAgent.ID,
				JobID:         jobID,
				Score:         score,
				Status:        model.MatchStatusPending,
			}
			if err := s.matchRepo.Create(match); err != nil {
				continue
			}
			result.CreatedCount++
			result.Matches = append(result.Matches, match)
		} else {
			result.SkippedCount++
		}
	}

	return result, nil
}
