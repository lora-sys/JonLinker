package integration

import (
	"encoding/json"
	"fmt"
	"time"

	"joblinker/internal/model"

	"github.com/google/uuid"
)

// SeedData contains all pre-populated test data for E2E testing
type SeedData struct {
	Organizations []model.Organization
	Users        []model.User
	Agents       []model.Agent
	Jobs         []model.Job
	Resumes      []model.Resume
	Matches      []model.Match
	Messages     []model.Message
	Interviews   []model.Interview
	Offers       []model.Offer
}

// TestOrganization creates a test organization
func (s *SeedData) TestOrganization(name string) model.Organization {
	org := model.Organization{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: time.Now(),
	}
	s.Organizations = append(s.Organizations, org)
	return org
}

// TestUser creates a test user with hashed password
func (s *SeedData) TestUser(email string, role string, orgID *uuid.UUID) model.User {
	user := model.User{
		ID:             uuid.New(),
		Email:          email,
		PasswordHash:   "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // "password123"
		Role:           model.UserRole(role),
		OrganizationID: orgID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	s.Users = append(s.Users, user)
	return user
}

// TestAgent creates a test agent
func (s *SeedData) TestAgent(userID uuid.UUID, agentType model.AgentType, status model.AgentStatus, config map[string]interface{}) model.Agent {
	configJSON, _ := json.Marshal(config)
	agent := model.Agent{
		ID:         uuid.New(),
		UserID:     userID,
		Type:       agentType,
		Status:     status,
		ConfigJSON: string(configJSON),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	s.Agents = append(s.Agents, agent)
	return agent
}

// TestJob creates a test job posting
func (s *SeedData) TestJob(agentID uuid.UUID, structured map[string]interface{}, status model.JobStatus) model.Job {
	structuredJSON, _ := json.Marshal(structured)
	job := model.Job{
		ID:             uuid.New(),
		AgentID:        agentID,
		StructuredJSON: string(structuredJSON),
		Status:         status,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	s.Jobs = append(s.Jobs, job)
	return job
}

// TestResume creates a test resume
func (s *SeedData) TestResume(agentID uuid.UUID, structured map[string]interface{}) model.Resume {
	structuredJSON, _ := json.Marshal(structured)
	resume := model.Resume{
		ID:             uuid.New(),
		AgentID:        agentID,
		StructuredJSON: string(structuredJSON),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	s.Resumes = append(s.Resumes, resume)
	return resume
}

// TestMatch creates a test match between seeker and job
func (s *SeedData) TestMatch(seekerAgentID, jobID uuid.UUID, score float64, status model.MatchStatus) model.Match {
	match := model.Match{
		ID:            uuid.New(),
		SeekerAgentID: seekerAgentID,
		JobID:         jobID,
		Score:         score,
		Status:        status,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	s.Matches = append(s.Matches, match)
	return match
}

// TestMessage creates a test A2A message
func (s *SeedData) TestMessage(matchID, senderAgentID uuid.UUID, contentXML, intentType string) model.Message {
	msg := model.Message{
		ID:            uuid.New(),
		MatchID:       matchID,
		SenderAgentID: senderAgentID,
		ContentXML:    contentXML,
		IntentType:    intentType,
		CreatedAt:     time.Now(),
	}
	s.Messages = append(s.Messages, msg)
	return msg
}

// TestInterview creates a test interview
func (s *SeedData) TestInterview(matchID uuid.UUID, scheduledAt time.Time, format model.InterviewFormat, status model.InterviewStatus) model.Interview {
	interview := model.Interview{
		ID:          uuid.New(),
		MatchID:     matchID,
		ScheduledAt: scheduledAt,
		Format:      format,
		Location:    "Video Call",
		Status:      status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	s.Interviews = append(s.Interviews, interview)
	return interview
}

// TestOffer creates a test offer
func (s *SeedData) TestOffer(matchID uuid.UUID, compensation map[string]interface{}, status model.OfferStatus) model.Offer {
	compensationJSON, _ := json.Marshal(compensation)
	startDate := time.Now().AddDate(0, 1, 0)
	offer := model.Offer{
		ID:               uuid.New(),
		MatchID:          matchID,
		CompensationJSON: string(compensationJSON),
		StartDate:        startDate,
		Status:           status,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	s.Offers = append(s.Offers, offer)
	return offer
}

// CreateComprehensiveSeedData creates a full set of seed data for E2E testing
// This simulates a realistic recruitment platform with multiple companies and candidates
func CreateComprehensiveSeedData() *SeedData {
	seed := &SeedData{}

	// ==================== ORGANIZATIONS ====================
	techCorp := seed.TestOrganization("TechCorp Inc.")
	startupHub := seed.TestOrganization("StartupHub")
	enterpriseCo := seed.TestOrganization("Enterprise Co.")

	// ==================== HR AGENTS (Recruiters) ====================
	// TechCorp HR Agents
	hrAlice := seed.TestUser("hr.alice@techcorp.com", "recruiter", &techCorp.ID)
	hrBob := seed.TestUser("hr.bob@techcorp.com", "recruiter", &techCorp.ID)
	_ = seed.TestAgent(hrAlice.ID, model.AgentTypeRecruiter, model.AgentStatusActive, map[string]interface{}{
		"name":       "Alice Chen",
		"title":      "Senior Recruiter",
		"company":    "TechCorp Inc.",
		"skills":     []string{"sourcing", "interviewing", "onboarding"},
		"department": "Human Resources",
	})

	agentBob := seed.TestAgent(hrBob.ID, model.AgentTypeRecruiter, model.AgentStatusActive, map[string]interface{}{
		"name":       "Bob Smith",
		"title":      "Technical Recruiter",
		"company":    "TechCorp Inc.",
		"skills":     []string{"technical screening", "code review", "system design"},
		"department": "Engineering Recruiting",
	})

	// StartupHub HR Agents
	hrCarol := seed.TestUser("hr.carol@startuphub.com", "recruiter", &startupHub.ID)
	agentCarol := seed.TestAgent(hrCarol.ID, model.AgentTypeRecruiter, model.AgentStatusActive, map[string]interface{}{
		"name":       "Carol Johnson",
		"title":      "Head of People",
		"company":    "StartupHub",
		"skills":     []string{"culture fit", "startup experience", "full-cycle recruiting"},
		"department": "People Operations",
	})

	// EnterpriseCo HR Agents
	hrDavid := seed.TestUser("hr.david@enterprise.com", "recruiter", &enterpriseCo.ID)
	agentDavid := seed.TestAgent(hrDavid.ID, model.AgentTypeRecruiter, model.AgentStatusActive, map[string]interface{}{
		"name":       "David Lee",
		"title":      "Corporate Recruiter",
		"company":    "Enterprise Co.",
		"skills":     []string{"enterprise sales", "B2B", "corporate recruiting"},
		"department": "Corporate Recruitment",
	})

	// ==================== BOSS AGENTS (Organization Admins) ====================
	// TechCorp Bosses
	ceoAlice := seed.TestUser("ceo.alice@techcorp.com", "recruiter", &techCorp.ID)
	ctoBob := seed.TestUser("cto.bob@techcorp.com", "recruiter", &techCorp.ID)
	_ = seed.TestAgent(ceoAlice.ID, model.AgentTypeRecruiter, model.AgentStatusActive, map[string]interface{}{
		"name":       "Alice Wong",
		"title":      "CEO",
		"company":     "TechCorp Inc.",
		"role":       "executive",
		"department": "Executive",
		"can_approve": true,
	})

	_ = seed.TestAgent(ctoBob.ID, model.AgentTypeRecruiter, model.AgentStatusActive, map[string]interface{}{
		"name":       "Bob Martinez",
		"title":      "CTO",
		"company":    "TechCorp Inc.",
		"role":       "executive",
		"department": "Engineering",
		"can_approve": true,
	})

	// StartupHub Boss
	ceoEve := seed.TestUser("ceo.eve@startuphub.com", "recruiter", &startupHub.ID)
	_ = seed.TestAgent(ceoEve.ID, model.AgentTypeRecruiter, model.AgentStatusActive, map[string]interface{}{
		"name":        "Eve Chen",
		"title":       "Founder & CEO",
		"company":      "StartupHub",
		"role":        "executive",
		"department":   "Executive",
		"can_approve":  true,
	})

	// ==================== SEEKER AGENTS (Job Seekers) ====================
	// TechCorp seekers
	seekerFrank := seed.TestUser("frank@email.com", "seeker", nil)
	seekerGrace := seed.TestUser("grace@email.com", "seeker", nil)
	seekerHenry := seed.TestUser("henry@email.com", "seeker", nil)

	// StartupHub seekers
	seekerIvy := seed.TestUser("ivy@email.com", "seeker", nil)
	seekerJack := seed.TestUser("jack@email.com", "seeker", nil)

	// EnterpriseCo seekers
	seekerKate := seed.TestUser("kate@email.com", "seeker", nil)

	// Independent seeker
	seekerLiam := seed.TestUser("liam@email.com", "seeker", nil)

	// Create seeker agents
	agentFrank := seed.TestAgent(seekerFrank.ID, model.AgentTypeSeeker, model.AgentStatusActive, map[string]interface{}{
		"name":       "Frank Garcia",
		"title":      "Senior Software Engineer",
		"skills":     []string{"Go", "Python", "Kubernetes", "AWS", "Microservices"},
		"experience": "8 years",
		"education":  "MS Computer Science, Stanford",
	})

	agentGrace := seed.TestAgent(seekerGrace.ID, model.AgentTypeSeeker, model.AgentStatusActive, map[string]interface{}{
		"name":       "Grace Kim",
		"title":      "Full Stack Developer",
		"skills":     []string{"React", "Node.js", "PostgreSQL", "TypeScript"},
		"experience": "5 years",
		"education":  "BS Computer Science, MIT",
	})

	agentHenry := seed.TestAgent(seekerHenry.ID, model.AgentTypeSeeker, model.AgentStatusActive, map[string]interface{}{
		"name":       "Henry Patel",
		"title":      "DevOps Engineer",
		"skills":     []string{"Terraform", "Docker", "CI/CD", "Linux", "Monitoring"},
		"experience": "6 years",
		"education":  "BS Information Technology, Berkeley",
	})

	agentIvy := seed.TestAgent(seekerIvy.ID, model.AgentTypeSeeker, model.AgentStatusActive, map[string]interface{}{
		"name":       "Ivy Zhang",
		"title":      "Frontend Engineer",
		"skills":     []string{"React", "Vue", "CSS", "Figma", "Animation"},
		"experience": "4 years",
		"education":  "BS Design, RISD",
	})

	agentJack := seed.TestAgent(seekerJack.ID, model.AgentTypeSeeker, model.AgentStatusActive, map[string]interface{}{
		"name":       "Jack Thompson",
		"title":      "Backend Engineer",
		"skills":     []string{"Java", "Spring Boot", "Kafka", "Redis"},
		"experience": "7 years",
		"education":  "MS Software Engineering, Carnegie Mellon",
	})

	agentKate := seed.TestAgent(seekerKate.ID, model.AgentTypeSeeker, model.AgentStatusActive, map[string]interface{}{
		"name":       "Kate Brown",
		"title":      "Data Scientist",
		"skills":     []string{"Python", "TensorFlow", "SQL", "Statistics", "Machine Learning"},
		"experience": "5 years",
		"education":  "PhD Statistics, Harvard",
	})

	_ = seed.TestAgent(seekerLiam.ID, model.AgentTypeSeeker, model.AgentStatusActive, map[string]interface{}{
		"name":       "Liam O'Connor",
		"title":      "Product Manager",
		"skills":     []string{"Product Strategy", "Agile", "User Research", "Analytics", "SQL"},
		"experience": "6 years",
		"education":  "MBA, Wharton",
	})

	// ==================== RESUMES ====================
	seed.TestResume(agentFrank.ID, map[string]interface{}{
		"summary": "Experienced software engineer specializing in distributed systems and cloud infrastructure",
		"skills":  []string{"Go", "Python", "Kubernetes", "AWS", "Microservices", "gRPC", "PostgreSQL"},
		"experience": []map[string]interface{}{
			{"title": "Senior Software Engineer", "company": "BigTech", "years": "2018-2026"},
			{"title": "Software Engineer", "company": "MedTech", "years": "2016-2018"},
		},
		"education": map[string]interface{}{
			"degree": "MS Computer Science",
			"school": "Stanford University",
			"year":   2016,
		},
	})

	seed.TestResume(agentGrace.ID, map[string]interface{}{
		"summary": "Full stack developer with expertise in modern web frameworks and scalable applications",
		"skills":  []string{"React", "Node.js", "PostgreSQL", "TypeScript", "GraphQL", "AWS"},
		"experience": []map[string]interface{}{
			{"title": "Full Stack Developer", "company": "WebScale", "years": "2019-2026"},
		},
		"education": map[string]interface{}{
			"degree": "BS Computer Science",
			"school": "MIT",
			"year":   2019,
		},
	})

	// ==================== JOBS (Posted by HR Agents) ====================
	// TechCorp Jobs
	job1 := seed.TestJob(agentBob.ID, map[string]interface{}{
		"title":       "Senior Software Engineer",
		"department":  "Engineering",
		"location":    "San Francisco, CA",
		"type":        "Full-time",
		"remote":      true,
		"salary_min":  180000,
		"salary_max":  250000,
		"skills":      []string{"Go", "Python", "Kubernetes", "AWS"},
		"description": "We are looking for a senior engineer to build our next-generation platform...",
	}, model.JobStatusActive)

	job2 := seed.TestJob(agentBob.ID, map[string]interface{}{
		"title":       "DevOps Engineer",
		"department":  "Infrastructure",
		"location":    "Remote",
		"type":        "Full-time",
		"remote":      true,
		"salary_min":  150000,
		"salary_max":  200000,
		"skills":      []string{"Terraform", "Docker", "CI/CD", "Linux"},
		"description": "Join our infrastructure team to scale our systems...",
	}, model.JobStatusActive)

	job3 := seed.TestJob(agentBob.ID, map[string]interface{}{
		"title":       "Frontend Engineer",
		"department":  "Product",
		"location":    "New York, NY",
		"type":        "Full-time",
		"remote":      true,
		"salary_min":  140000,
		"salary_max":  190000,
		"skills":      []string{"React", "TypeScript", "CSS", "Animation"},
		"description": "Build beautiful, performant user interfaces...",
	}, model.JobStatusActive)

	// StartupHub Jobs
	job4 := seed.TestJob(agentCarol.ID, map[string]interface{}{
		"title":       "Full Stack Developer",
		"department":  "Engineering",
		"location":    "Austin, TX",
		"type":        "Full-time",
		"remote":      true,
		"salary_min":  130000,
		"salary_max":  170000,
		"skills":      []string{"React", "Node.js", "PostgreSQL", "TypeScript"},
		"description": "Be part of a fast-paced startup building the future of work...",
	}, model.JobStatusActive)

	job5 := seed.TestJob(agentCarol.ID, map[string]interface{}{
		"title":       "Backend Engineer",
		"department":  "Engineering",
		"location":    "Austin, TX",
		"type":        "Full-time",
		"remote":      false,
		"salary_min":  140000,
		"salary_max":  180000,
		"skills":      []string{"Java", "Spring Boot", "Kafka", "Redis"},
		"description": "Build scalable backend systems for our enterprise clients...",
	}, model.JobStatusActive)

	// EnterpriseCo Jobs
	job6 := seed.TestJob(agentDavid.ID, map[string]interface{}{
		"title":       "Data Scientist",
		"department":  "Analytics",
		"location":    "Chicago, IL",
		"type":        "Full-time",
		"remote":      true,
		"salary_min":  160000,
		"salary_max":  220000,
		"skills":      []string{"Python", "TensorFlow", "SQL", "Statistics"},
		"description": "Drive data-informed decisions across the organization...",
	}, model.JobStatusActive)

	// ==================== MATCHES (Seeker-Job Matches) ====================
	// Frank matches with Senior SE job (high score)
	match1 := seed.TestMatch(agentFrank.ID, job1.ID, 0.92, model.MatchStatusPending)

	// Grace matches with Frontend job (high score)
	_ = seed.TestMatch(agentGrace.ID, job3.ID, 0.88, model.MatchStatusPending)

	// Henry matches with DevOps job (high score)
	_ = seed.TestMatch(agentHenry.ID, job2.ID, 0.85, model.MatchStatusPending)

	// Ivy matches with Full Stack job (medium score - will proceed)
	match4 := seed.TestMatch(agentIvy.ID, job4.ID, 0.78, model.MatchStatusMutualInterest)

	// Jack matches with Backend job (high score)
	match5 := seed.TestMatch(agentJack.ID, job5.ID, 0.90, model.MatchStatusNegotiating)

	// Kate matches with Data Scientist job
	match6 := seed.TestMatch(agentKate.ID, job6.ID, 0.87, model.MatchStatusOffered)

	// ==================== MESSAGES (A2A Conversations) ====================
	// Frank's conversation with TechCorp HR (INTRODUCTION -> INTEREST)
	seed.TestMessage(match1.ID, agentFrank.ID, `<message>
		<header>
			<message_id>msg-001</message_id>
			<timestamp>2026-04-20T10:00:00Z</timestamp>
			<sender_id>`+agentFrank.ID.String()+`</sender_id>
			<receiver_id>`+agentBob.ID.String()+`</receiver_id>
		</header>
		<payload>
			<intent>INTRODUCTION</intent>
			<parameters>
				<role>seeker</role>
				<name>Frank Garcia</name>
				<title>Senior Software Engineer</title>
				<company>Independent</company>
				<skills>Go, Python, Kubernetes, AWS, Microservices</skills>
				<experience>8 years</experience>
			</parameters>
		</payload>
	</message>`, "INTRODUCTION")

	seed.TestMessage(match1.ID, agentBob.ID, `<message>
		<header>
			<message_id>msg-002</message_id>
			<timestamp>2026-04-20T10:05:00Z</timestamp>
			<sender_id>`+agentBob.ID.String()+`</sender_id>
			<receiver_id>`+agentFrank.ID.String()+`</receiver_id>
			<reply_to>msg-001</reply_to>
		</header>
		<payload>
			<intent>INTEREST</intent>
			<parameters>
				<role>recruiter</role>
				<name>Bob Smith</name>
				<title>Technical Recruiter</title>
				<company>TechCorp Inc.</company>
				<interest_level>high</interest_level>
			</parameters>
		</payload>
	</message>`, "INTEREST")

	// Jack's negotiation conversation (NEGOTIATION in progress)
	seed.TestMessage(match5.ID, agentJack.ID, `<message>
		<header>
			<message_id>msg-010</message_id>
			<timestamp>2026-04-18T14:00:00Z</timestamp>
			<sender_id>`+agentJack.ID.String()+`</sender_id>
			<receiver_id>`+agentCarol.ID.String()+`</receiver_id>
		</header>
		<payload>
			<intent>NEGOTIATION</intent>
			<negotiation round="1" type="salary" current="140000" target="160000" currency="USD" concessions="0">
				<compensation>
					<base_salary>140000</base_salary>
					<currency>USD</currency>
				</compensation>
			</negotiation>
		</payload>
	</message>`, "NEGOTIATION")

	seed.TestMessage(match5.ID, agentCarol.ID, `<message>
		<header>
			<message_id>msg-011</message_id>
			<timestamp>2026-04-18T14:30:00Z</timestamp>
			<sender_id>`+agentCarol.ID.String()+`</sender_id>
			<receiver_id>`+agentJack.ID.String()+`</receiver_id>
			<reply_to>msg-010</reply_to>
		</header>
		<payload>
			<intent>NEGOTIATION</intent>
			<negotiation round="2" type="salary" current="150000" target="160000" currency="USD" concessions="1">
				<compensation>
					<base_salary>150000</base_salary>
					<currency>USD</currency>
					<bonus>10000</bonus>
				</compensation>
			</negotiation>
		</payload>
	</message>`, "NEGOTIATION")

	// Kate's offer conversation (OFFER stage)
	seed.TestMessage(match6.ID, agentDavid.ID, `<message>
		<header>
			<message_id>msg-020</message_id>
			<timestamp>2026-04-19T09:00:00Z</timestamp>
			<sender_id>`+agentDavid.ID.String()+`</sender_id>
			<receiver_id>`+agentKate.ID.String()+`</receiver_id>
		</header>
		<payload>
			<intent>OFFER</intent>
			<negotiation round="0" type="compensation" current="180000" target="200000" currency="USD">
				<compensation>
					<base_salary>180000</base_salary>
					<currency>USD</currency>
					<bonus>15000</bonus>
					<equity>0.05%</equity>
				</compensation>
				<start_date>2026-06-01</start_date>
			</negotiation>
		</payload>
	</message>`, "OFFER")

	// ==================== INTERVIEWS ====================
	// Ivy's first round interview (SCHEDULED)
	scheduledTime := time.Now().AddDate(0, 0, 5) // 5 days from now
	seed.TestInterview(match4.ID, scheduledTime, model.InterviewFormatVideo, model.InterviewStatusScheduled)

	// Kate's second round interview (completed with feedback)
	interviewTime := time.Now().AddDate(0, 0, -2) // 2 days ago
	kateInterview := seed.TestInterview(match6.ID, interviewTime, model.InterviewFormatOnsite, model.InterviewStatusCompleted)
	kateInterview.Location = "Enterprise Co. HQ, Chicago"
	fb := `{"rating": 4.5, "strengths": ["Technical depth", "Communication", "Problem solving"], "concerns": ["Salary expectations"]}`
	kateInterview.Feedback = fb

	// ==================== OFFERS ====================
	// Kate's offer (PENDING response)
	seed.TestOffer(match6.ID, map[string]interface{}{
		"base_salary": 180000,
		"currency":    "USD",
		"bonus":       15000,
		"equity":      "0.05%",
		"total_comp":  195000,
	}, model.OfferStatusPending)

	return seed
}

// GetAgentsByType returns all agents of a specific type
func (s *SeedData) GetAgentsByType(agentType model.AgentType) []model.Agent {
	var result []model.Agent
	for _, a := range s.Agents {
		if a.Type == agentType {
			result = append(result, a)
		}
	}
	return result
}

// GetAgentsByOrganization returns all agents belonging to an organization
func (s *SeedData) GetAgentsByOrganization(orgID uuid.UUID) []model.Agent {
	var result []model.Agent
	for _, u := range s.Users {
		if u.OrganizationID != nil && *u.OrganizationID == orgID {
			for _, a := range s.Agents {
				if a.UserID == u.ID {
					result = append(result, a)
				}
			}
		}
	}
	return result
}

// GetMatchesBySeeker returns all matches for a seeker agent
func (s *SeedData) GetMatchesBySeeker(seekerAgentID uuid.UUID) []model.Match {
	var result []model.Match
	for _, m := range s.Matches {
		if m.SeekerAgentID == seekerAgentID {
			result = append(result, m)
		}
	}
	return result
}

// GetMessagesByMatch returns all messages for a match
func (s *SeedData) GetMessagesByMatch(matchID uuid.UUID) []model.Message {
	var result []model.Message
	for _, m := range s.Messages {
		if m.MatchID == matchID {
			result = append(result, m)
		}
	}
	return result
}

// PrintSeedSummary prints a summary of the seed data
func (s *SeedData) PrintSeedSummary() {
	fmt.Println("\n========== SEED DATA SUMMARY ==========")
	fmt.Printf("Organizations: %d\n", len(s.Organizations))
	fmt.Printf("Users: %d\n", len(s.Users))
	fmt.Printf("  - Seekers: %d\n", len(s.GetAgentsByType(model.AgentTypeSeeker)))
	fmt.Printf("  - Recruiters: %d\n", len(s.GetAgentsByType(model.AgentTypeRecruiter)))
	fmt.Printf("Agents: %d\n", len(s.Agents))
	fmt.Printf("Jobs: %d\n", len(s.Jobs))
	fmt.Printf("Resumes: %d\n", len(s.Resumes))
	fmt.Printf("Matches: %d\n", len(s.Matches))
	fmt.Printf("Messages: %d\n", len(s.Messages))
	fmt.Printf("Interviews: %d\n", len(s.Interviews))
	fmt.Printf("Offers: %d\n", len(s.Offers))
	fmt.Println("========================================")

	// Print organization breakdown
	for _, org := range s.Organizations {
		agents := s.GetAgentsByOrganization(org.ID)
		fmt.Printf("Organization: %s (%d agents)\n", org.Name, len(agents))
		for _, a := range agents {
			for _, u := range s.Users {
				if u.ID == a.UserID {
					fmt.Printf("  - %s (%s) - %s\n", u.Email, a.Type, a.Status)
					break
				}
			}
		}
	}
}
