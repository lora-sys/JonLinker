package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"joblinker/internal/model"
	"joblinker/internal/repository"
	"joblinker/pkg/rabbitmq"

	"github.com/google/uuid"
)

var ErrInvalidMatchState = errors.New("invalid match state transition")

type MatchService struct {
	matchRepo *repository.MatchRepository
	agentRepo *repository.AgentRepository
	jobRepo   *repository.JobRepository
	rmq       *rabbitmq.RabbitMQ
}

func NewMatchService(matchRepo *repository.MatchRepository, agentRepo *repository.AgentRepository, jobRepo *repository.JobRepository, rmq *rabbitmq.RabbitMQ) *MatchService {
	return &MatchService{
		matchRepo: matchRepo,
		agentRepo: agentRepo,
		jobRepo:   jobRepo,
		rmq:       rmq,
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

	// Auto-start A2A conversation immediately
	if err := s.autoStartA2A(match); err != nil {
		log.Printf("[CreateMatch] Failed to auto-start A2A for match %s: %v", match.ID, err)
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

	// Auto-start A2A conversation (handles pending → mutual_interest transition)
	if err := s.autoStartA2A(match); err != nil {
		return nil, fmt.Errorf("failed to start A2A: %w", err)
	}

	return match, nil
}

func (s *MatchService) DeclineMatch(id uuid.UUID) (*model.Match, error) {
	match, err := s.matchRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if match.Status != model.MatchStatusPending {
		return nil, ErrInvalidMatchState
	}
	match.Status = model.MatchStatusRejected
	if err := s.matchRepo.Update(match); err != nil {
		return nil, err
	}
	return match, nil
}

func (s *MatchService) autoStartA2A(match *model.Match) error {
	if s.rmq == nil {
		return errors.New("rabbitmq not available")
	}

	// Transition match from pending → mutual_interest so A2A conversation starts
	if match.Status == model.MatchStatusPending {
		match.Status = model.MatchStatusMutualInterest
		if err := s.matchRepo.Update(match); err != nil {
			log.Printf("[autoStartA2A] Failed to update match %s status: %v", match.ID, err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	seeker, err := s.agentRepo.GetByID(match.SeekerAgentID)
	if err != nil {
		return fmt.Errorf("failed to get seeker agent: %w", err)
	}

	seekerName := "Candidate"
	var seekerSkills []string
	seekerYearsExp := 0
	seekerLocation := ""
	seekerTitle := ""
	var config map[string]interface{}
	if seeker.ConfigJSON != nil {
		if json.Unmarshal(seeker.ConfigJSON, &config) == nil {
			if name, ok := config["name"].(string); ok && name != "" {
				seekerName = name
			}
			if skills, ok := config["skills"].([]interface{}); ok {
				for _, sk := range skills {
					if s, ok := sk.(string); ok {
						seekerSkills = append(seekerSkills, s)
					}
				}
			}
			if exp, ok := config["experience_years"].(float64); ok {
				seekerYearsExp = int(exp)
			}
			if loc, ok := config["location"].(string); ok {
				seekerLocation = loc
			}
			if title, ok := config["title"].(string); ok {
				seekerTitle = title
			}
		}
	}

	jobTitle, jobCompany := "", ""
	salaryMin, salaryMax := 0, 0
	jobLocation := ""
	var jobSkills []string
	job, _ := s.jobRepo.GetByID(match.JobID)
	if job != nil {
		var jd map[string]interface{}
		if json.Unmarshal(job.StructuredJSON, &jd) == nil {
			if t, ok := jd["title"].(string); ok {
				jobTitle = t
			}
			if c, ok := jd["company"].(string); ok {
				jobCompany = c
			}
			if loc, ok := jd["location"].(string); ok {
				jobLocation = loc
			}
			if min, ok := jd["salary_min"].(float64); ok {
				salaryMin = int(min)
			}
			if max, ok := jd["salary_max"].(float64); ok {
				salaryMax = int(max)
			}
			if skills, ok := jd["skills"].([]interface{}); ok {
				for _, sk := range skills {
					if s, ok := sk.(string); ok {
						jobSkills = append(jobSkills, s)
					}
				}
			}
			if jobCompany == "" {
				jobCompany = "TechCorp"
			}
		}
	}

	var intro string
	if salaryMin > 0 || salaryMax > 0 {
		intro = fmt.Sprintf(
			"Hi, I'm %s. I'm very interested in the %s position at %s. "+
				"I'm a %s based in %s with %d years of experience in %s. "+
				"I see the role is looking for %s in %s, offering $%d-$%d — this aligns strongly with my background. "+
				"I'd love to discuss how my experience can contribute to the team.",
			seekerName, jobTitle, jobCompany,
			seekerTitle, seekerLocation, seekerYearsExp, strings.Join(seekerSkills, ", "),
			strings.Join(jobSkills, ", "), jobLocation, salaryMin, salaryMax,
		)
	} else {
		intro = fmt.Sprintf(
			"Hi, I'm %s. I'm very interested in the %s position at %s. "+
				"I'm a %s based in %s with %d years of experience in %s. "+
				"My background in %s seems like a great match for this role, and I'd love to discuss it further.",
			seekerName, jobTitle, jobCompany,
			seekerTitle, seekerLocation, seekerYearsExp, strings.Join(seekerSkills, ", "),
			strings.Join(jobSkills, ", "),
		)
	}

	msg := &rabbitmq.AgentMessage{
		MessageID:  uuid.New().String(),
		SenderID:   match.SeekerAgentID.String(),
		ReceiverID: "",
		Intent:     "INQUIRY",
		MatchID:    match.ID.String(),
		Payload: map[string]interface{}{
			"message": intro,
			"content": intro,
		},
		Timestamp: time.Now(),
	}

	log.Printf("[autoStartA2A] Publishing initial INQUIRY for match %s from seeker %s", match.ID, seekerName)
	return s.rmq.PublishAgentMessage(ctx, msg)
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

	// Collect all agent IDs for this user
	var agentIDs []uuid.UUID
	for _, agent := range agents {
		agentIDs = append(agentIDs, agent.ID)
	}

	if len(agentIDs) == 0 {
		return []*model.Match{}, nil
	}

	// Get all matches where user owns the seeker agent OR owns the job's recruiter agent
	allMatches, err := s.matchRepo.ListByAgentIDs(agentIDs)
	if err != nil {
		return nil, err
	}

	// Filter matches: keep if seeker_agent_id belongs to user OR job's agent_id belongs to user
	var filtered []*model.Match
	agentMap := make(map[uuid.UUID]bool)
	for _, a := range agents {
		agentMap[a.ID] = true
	}

	for _, match := range allMatches {
		if agentMap[match.SeekerAgentID] {
			filtered = append(filtered, match)
			continue
		}
		if match.Job != nil && agentMap[match.Job.AgentID] {
			filtered = append(filtered, match)
		}
	}

	return filtered, nil
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
	if json.Unmarshal(seeker.ConfigJSON, &config) == nil {
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
	if json.Unmarshal(job.StructuredJSON, &jobData) == nil {
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
		log.Printf("[score] calculateSkillsMatch: jobSkills empty, falling back to 1.0 (seekerSkills=%v)", seekerSkills)
		return 1.0
	}
	if len(seekerSkills) == 0 {
		log.Printf("[score] calculateSkillsMatch: seekerSkills empty, falling back to 0")
		return 0
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
		log.Printf("[score] calculateLocationMatch: empty seekerLoc=%q jobLoc=%q, falling back to 1.0", seekerLoc, jobLoc)
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
		log.Printf("[score] calculateExperienceMatch: jobYears=0, falling back to 1.0 (seekerYears=%d)", seekerYears)
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

		job, err := s.jobRepo.GetByID(jobID)
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
				SeekerAgentID:   seekerAgent.ID,
				RecruiterAgentID: &job.AgentID,
				JobID:           jobID,
				Score:           score,
				Status:          model.MatchStatusPending,
			}
			if err := s.matchRepo.Create(match); err != nil {
				continue
			}
			result.CreatedCount++
			result.Matches = append(result.Matches, match)

			// Auto-start A2A conversation for this match
			if err := s.autoStartA2A(match); err != nil {
				log.Printf("[AutoCreateMatches] Failed to start A2A for match %s: %v", match.ID, err)
			}
		} else {
			result.SkippedCount++
		}
	}

	return result, nil
}
