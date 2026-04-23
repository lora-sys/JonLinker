package agent

import (
	"fmt"
	"strings"
	"time"
)

// RequirementsHandler manages job requirements discussion between agents
type RequirementsHandler struct {
	mandatory []string
	preferred []string
	flexible  []string
}

func NewRequirementsHandler() *RequirementsHandler {
	return &RequirementsHandler{
		mandatory: []string{},
		preferred: []string{},
		flexible:  []string{},
	}
}

type Requirement struct {
	Name        string
	Category    string
	Priority    Priority
	Flexibility Flexibility
	Satisfied   bool
}

type Priority string

const (
	PriorityMandatory Priority = "mandatory"
	PriorityPreferred Priority = "preferred"
	PriorityNiceToHave Priority = "nice_to_have"
)

type Flexibility string

const (
	FlexibilityFixed     Flexibility = "fixed"
	FlexibilityNegotiable Flexibility = "negotiable"
	FlexibilityFlexible  Flexibility = "flexible"
)

type RequirementsProfile struct {
	Role            string
	Experience      int
	Skills          []string
	Location        string
	WorkAuth        string
	SalaryRange     [2]int // [min, max]
	RemoteOK        bool
	StartDate       time.Time
	NoticePeriod    int // days
}

// ParseRequirements converts a structured profile into requirements
func ParseRequirements(profile *RequirementsProfile) []*Requirement {
	reqs := []*Requirement{}

	// Experience
	reqs = append(reqs, &Requirement{
		Name:        fmt.Sprintf("Experience: %d+ years", profile.Experience),
		Category:    "experience",
		Priority:    PriorityMandatory,
		Flexibility: FlexibilityFixed,
	})

	// Skills
	for _, skill := range profile.Skills {
		reqs = append(reqs, &Requirement{
			Name:        skill,
			Category:    "skills",
			Priority:    PriorityMandatory,
			Flexibility: FlexibilityNegotiable,
		})
	}

	// Location
	if profile.Location != "" {
		locReq := PriorityPreferred
		flex := FlexibilityNegotiable
		if profile.RemoteOK {
			locReq = PriorityNiceToHave
			flex = FlexibilityFlexible
		}
		reqs = append(reqs, &Requirement{
			Name:        profile.Location,
			Category:    "location",
			Priority:    locReq,
			Flexibility: flex,
		})
	}

	// Work authorization
	if profile.WorkAuth != "" {
		reqs = append(reqs, &Requirement{
			Name:        profile.WorkAuth,
			Category:    "work_auth",
			Priority:    PriorityMandatory,
			Flexibility: FlexibilityFixed,
		})
	}

	// Salary
	reqs = append(reqs, &Requirement{
		Name:        fmt.Sprintf("Salary: $%d-$%d", profile.SalaryRange[0], profile.SalaryRange[1]),
		Category:    "compensation",
		Priority:    PriorityMandatory,
		Flexibility: FlexibilityNegotiable,
	})

	return reqs
}

// MatchScore calculates how well requirements are met
func MatchScore(candidate *RequirementsProfile, job *RequirementsProfile) float64 {
	var score float64 = 0
	var weights float64 = 0

	// Experience (weight: 0.2)
	if candidate.Experience >= job.Experience {
		score += 0.2
	} else if candidate.Experience >= job.Experience-2 {
		score += 0.1
	}
	weights += 0.2

	// Skills (weight: 0.3)
	jobSkills := make(map[string]bool)
	for _, s := range job.Skills {
		jobSkills[strings.ToLower(s)] = true
	}
	matched := 0
	for _, s := range candidate.Skills {
		if jobSkills[strings.ToLower(s)] {
			matched++
		}
	}
	if len(job.Skills) > 0 {
		score += 0.3 * float64(matched) / float64(len(job.Skills))
	}
	weights += 0.3

	// Location (weight: 0.15)
	if candidate.Location == job.Location || candidate.RemoteOK {
		score += 0.15
	}
	weights += 0.15

	// Compensation (weight: 0.25)
	if candidate.SalaryRange[0] <= job.SalaryRange[0] && candidate.SalaryRange[1] >= job.SalaryRange[1] {
		score += 0.25
	} else if candidate.SalaryRange[0] <= job.SalaryRange[1] {
		score += 0.1
	}
	weights += 0.25

	// Work auth (weight: 0.1)
	if candidate.WorkAuth == job.WorkAuth || candidate.WorkAuth == "authorized" {
		score += 0.1
	}
	weights += 0.1

	if weights == 0 {
		return 0
	}
	return score / weights
}

// NegotiationState tracks the current state of requirements negotiation
type NegotiationState struct {
	SeekerReqs *RequirementsProfile
	JobReqs    *RequirementsProfile
	Met        map[string]bool
	Gaps       map[string]int // requirement name -> gap size
}

// NewNegotiationState creates a new requirements negotiation context
func NewNegotiationState(seeker, job *RequirementsProfile) *NegotiationState {
	return &NegotiationState{
		SeekerReqs: seeker,
		JobReqs:    job,
		Met:        make(map[string]bool),
		Gaps:       make(map[string]int),
	}
}

// IdentifyGaps finds mismatches between seeker and job requirements
func (ns *NegotiationState) IdentifyGaps() map[string]int {
	gaps := make(map[string]int)

	// Check experience
	if ns.SeekerReqs.Experience < ns.JobReqs.Experience {
		gaps["experience"] = ns.JobReqs.Experience - ns.SeekerReqs.Experience
	}

	// Check salary overlap
	seekerMax := ns.SeekerReqs.SalaryRange[1]
	jobMin := ns.JobReqs.SalaryRange[0]
	if seekerMax < jobMin {
		gaps["salary"] = jobMin - seekerMax
	}

	// Check skills match
	seekerSkills := make(map[string]bool)
	for _, s := range ns.SeekerReqs.Skills {
		seekerSkills[strings.ToLower(s)] = true
	}
	missingSkills := []string{}
	for _, s := range ns.JobReqs.Skills {
		if !seekerSkills[strings.ToLower(s)] {
			missingSkills = append(missingSkills, s)
		}
	}
	if len(missingSkills) > 0 {
		gaps["missing_skills"] = len(missingSkills)
	}

	ns.Gaps = gaps
	return gaps
}

// CanNegotiate checks if gaps can be bridged through negotiation
func (ns *NegotiationState) CanNegotiate() bool {
	// Can always negotiate salary and some skills
	if ns.SeekerReqs.Experience >= ns.JobReqs.Experience-1 {
		return true
	}
	return false
}

// FormatRequirementsSummary creates a human-readable requirements summary
func FormatRequirementsSummary(profile *RequirementsProfile) string {
	parts := []string{
		fmt.Sprintf("%d+ years experience", profile.Experience),
		fmt.Sprintf("Skills: %s", strings.Join(profile.Skills, ", ")),
		fmt.Sprintf("Salary: $%d-$%dK", profile.SalaryRange[0]/1000, profile.SalaryRange[1]/1000),
	}
	if profile.RemoteOK {
		parts = append(parts, "Remote OK")
	}
	return strings.Join(parts, " | ")
}