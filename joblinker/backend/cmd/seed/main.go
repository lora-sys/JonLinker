package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"joblinker/internal/model"
)

func main() {
	godotenv.Load("../../../.env")

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=joblinker password=joblinker_dev dbname=joblinker port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate to ensure all tables exist
	db.AutoMigrate(
		&model.User{},
		&model.Organization{},
		&model.Agent{},
		&model.Resume{},
		&model.Job{},
		&model.Match{},
		&model.Message{},
		&model.Interview{},
		&model.Offer{},
		&model.SecurityEvent{},
	)

	log.Println("Starting seed data creation...")

	// Clear existing seed data
	db.Exec("DELETE FROM messages")
	db.Exec("DELETE FROM interviews")
	db.Exec("DELETE FROM offers")
	db.Exec("DELETE FROM matches")
	db.Exec("DELETE FROM jobs")
	db.Exec("DELETE FROM resumes")
	db.Exec("DELETE FROM agents")
	db.Exec("DELETE FROM users")
	db.Exec("DELETE FROM organizations")
	log.Println("Cleared existing data")

	// Create Organizations
	orgs := createOrganizations(db)
	log.Printf("Created %d organizations", len(orgs))

	// Create Users (recruiters and seekers)
	users := createUsers(db, orgs)
	log.Printf("Created %d users", len(users))

	// Create Agents
	agents := createAgents(db, users)
	log.Printf("Created %d agents", len(agents))

	// Create Jobs
	jobs := createJobs(db, agents)
	log.Printf("Created %d jobs", len(jobs))

	// Create Matches
	matches := createMatches(db, agents, jobs)
	log.Printf("Created %d matches", len(matches))

	// Create Messages
	createMessages(db, agents, matches)
	log.Println("Created messages")

	// Create Interviews
	createInterviews(db, matches)
	log.Println("Created interviews")

	// Create Offers
	createOffers(db, matches)
	log.Println("Created offers")

	log.Println("Seed data creation complete!")
	printSummary(users, agents, jobs, matches)
}

func createOrganizations(db *gorm.DB) []model.Organization {
	orgs := []model.Organization{
		{Name: "TechCorp Inc."},
		{Name: "StartupHub"},
		{Name: "Enterprise Co."},
		{Name: "Innovation Labs"},
	}

	for i := range orgs {
		orgs[i].ID = uuid.New()
		orgs[i].CreatedAt = time.Now()
		db.Create(&orgs[i])
	}
	return orgs
}

func createUsers(db *gorm.DB, orgs []model.Organization) []model.User {
	// Generate proper bcrypt hash for "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	passwordHash := string(hashedPassword)

	users := []model.User{
		// TechCorp HR
		{Email: "hr.alice@techcorp.com", PasswordHash: passwordHash, Role: model.RoleRecruiter, OrganizationID: &orgs[0].ID},
		{Email: "hr.bob@techcorp.com", PasswordHash: passwordHash, Role: model.RoleRecruiter, OrganizationID: &orgs[0].ID},
		{Email: "ceo.alice@techcorp.com", PasswordHash: passwordHash, Role: model.RoleRecruiter, OrganizationID: &orgs[0].ID},

		// StartupHub HR
		{Email: "hr.carol@startuphub.com", PasswordHash: passwordHash, Role: model.RoleRecruiter, OrganizationID: &orgs[1].ID},
		{Email: "ceo.eve@startuphub.com", PasswordHash: passwordHash, Role: model.RoleRecruiter, OrganizationID: &orgs[1].ID},

		// Enterprise Co HR
		{Email: "hr.david@enterprise.com", PasswordHash: passwordHash, Role: model.RoleRecruiter, OrganizationID: &orgs[2].ID},

		// Innovation Labs HR
		{Email: "hr.frank@innovation.com", PasswordHash: passwordHash, Role: model.RoleRecruiter, OrganizationID: &orgs[3].ID},

		// Seekers
		{Email: "seeker.frank@email.com", PasswordHash: passwordHash, Role: model.RoleSeeker},
		{Email: "seeker.grace@email.com", PasswordHash: passwordHash, Role: model.RoleSeeker},
		{Email: "seeker.henry@email.com", PasswordHash: passwordHash, Role: model.RoleSeeker},
		{Email: "seeker.ivy@email.com", PasswordHash: passwordHash, Role: model.RoleSeeker},
		{Email: "seeker.jack@email.com", PasswordHash: passwordHash, Role: model.RoleSeeker},
		{Email: "seeker.kate@email.com", PasswordHash: passwordHash, Role: model.RoleSeeker},
		{Email: "seeker.liam@email.com", PasswordHash: passwordHash, Role: model.RoleSeeker},
	}

	for i := range users {
		users[i].ID = uuid.New()
		users[i].CreatedAt = time.Now()
		users[i].UpdatedAt = time.Now()
		db.Create(&users[i])
	}
	return users
}

func createAgents(db *gorm.DB, users []model.User) map[string]model.Agent {
	agents := make(map[string]model.Agent)

	agentConfigs := []struct {
		userEmail       string
		agentType       model.AgentType
		name            string
		title           string
		company         string
		skills          []string
		experience      string
		experienceYears int
		location        string
	}{
		// Recruiter Agents
		{"hr.alice@techcorp.com", model.AgentTypeRecruiter, "Alice Chen", "Senior Recruiter", "TechCorp Inc.", []string{"sourcing", "interviewing", "onboarding"}, "5 years", 5, "San Francisco, CA"},
		{"hr.bob@techcorp.com", model.AgentTypeRecruiter, "Bob Smith", "Technical Recruiter", "TechCorp Inc.", []string{"technical screening", "code review", "system design"}, "4 years", 4, "San Francisco, CA"},
		{"hr.carol@startuphub.com", model.AgentTypeRecruiter, "Carol Johnson", "Head of People", "StartupHub", []string{"culture fit", "startup experience", "full-cycle recruiting"}, "6 years", 6, "Austin, TX"},
		{"hr.david@enterprise.com", model.AgentTypeRecruiter, "David Lee", "Corporate Recruiter", "Enterprise Co.", []string{"enterprise sales", "B2B", "corporate recruiting"}, "7 years", 7, "Chicago, IL"},
		{"hr.frank@innovation.com", model.AgentTypeRecruiter, "Frank Miller", "Lead Recruiter", "Innovation Labs", []string{"AI/ML", "startups", "technical recruiting"}, "8 years", 8, "Boston, MA"},

		// Seeker Agents
		{"seeker.frank@email.com", model.AgentTypeSeeker, "Frank Garcia", "Senior Software Engineer", "Independent", []string{"Go", "Python", "Kubernetes", "AWS", "Microservices"}, "8 years", 8, "San Francisco, CA"},
		{"seeker.grace@email.com", model.AgentTypeSeeker, "Grace Kim", "Full Stack Developer", "Independent", []string{"React", "Node.js", "PostgreSQL", "TypeScript", "GraphQL"}, "5 years", 5, "Austin, TX"},
		{"seeker.henry@email.com", model.AgentTypeSeeker, "Henry Patel", "DevOps Engineer", "Independent", []string{"Terraform", "Docker", "CI/CD", "Linux", "Monitoring"}, "6 years", 6, "Remote"},
		{"seeker.ivy@email.com", model.AgentTypeSeeker, "Ivy Zhang", "Frontend Engineer", "Independent", []string{"React", "Vue", "CSS", "Figma", "Animation"}, "4 years", 4, "New York, NY"},
		{"seeker.jack@email.com", model.AgentTypeSeeker, "Jack Thompson", "Backend Engineer", "Independent", []string{"Java", "Spring Boot", "Kafka", "Redis", "Microservices"}, "7 years", 7, "Austin, TX"},
		{"seeker.kate@email.com", model.AgentTypeSeeker, "Kate Brown", "Data Scientist", "Independent", []string{"Python", "TensorFlow", "SQL", "Statistics", "Machine Learning"}, "5 years", 5, "Chicago, IL"},
		{"seeker.liam@email.com", model.AgentTypeSeeker, "Liam O'Connor", "Product Manager", "Independent", []string{"Product Strategy", "Agile", "User Research", "Analytics", "SQL"}, "6 years", 6, "Boston, MA"},
	}

	for _, config := range agentConfigs {
		var userID uuid.UUID
		for _, u := range users {
			if u.Email == config.userEmail {
				userID = u.ID
				break
			}
		}

		if userID == uuid.Nil {
			continue
		}

		configJSON, _ := json.Marshal(map[string]interface{}{
			"name":             config.name,
			"title":            config.title,
			"company":          config.company,
			"skills":           config.skills,
			"experience_years": config.experienceYears,
			"location":         config.location,
		})

		agent := model.Agent{
			ID:         uuid.New(),
			UserID:     userID,
			Type:       config.agentType,
			Status:     model.AgentStatusActive,
			ConfigJSON: json.RawMessage(configJSON),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		db.Create(&agent)
		agents[config.userEmail] = agent
	}

	return agents
}

func createJobs(db *gorm.DB, agents map[string]model.Agent) []model.Job {
	jobsData := []struct {
		recruiterEmail string
		title          string
		description    string
		requirements   []string
		location       string
		salaryMin      int
		salaryMax      int
		workType       string
	}{
		{"hr.bob@techcorp.com", "Senior Software Engineer", "Build next-generation distributed systems for our cloud platform. Work with Go, Python, Kubernetes, and AWS to scale services to millions of users.", []string{"Go", "Python", "Kubernetes", "AWS", "Microservices"}, "San Francisco, CA", 180000, 250000, "hybrid"},
		{"hr.bob@techcorp.com", "DevOps Engineer", "Scale our cloud infrastructure and improve CI/CD pipelines. Lead infrastructure automation using Terraform and Docker.", []string{"Terraform", "Docker", "CI/CD", "Linux", "AWS"}, "Remote", 150000, 200000, "remote"},
		{"hr.bob@techcorp.com", "Frontend Engineer", "Build beautiful user experiences with React and TypeScript. Collaborate with designers to create stunning interfaces.", []string{"React", "TypeScript", "CSS", "Figma"}, "New York, NY", 140000, 190000, "onsite"},
		{"hr.alice@techcorp.com", "Senior Backend Engineer", "Build scalable microservices architecture using Go and Rust. Design and implement high-performance APIs handling millions of requests per day.", []string{"Go", "Rust", "PostgreSQL", "gRPC", "Kubernetes"}, "San Francisco, CA", 200000, 280000, "hybrid"},
		{"hr.alice@techcorp.com", "Site Reliability Engineer", "Design and implement monitoring, alerting, and chaos engineering practices. Build resilient infrastructure that scales globally.", []string{"Terraform", "Prometheus", "Grafana", "Kubernetes", "Linux"}, "Remote", 170000, 230000, "remote"},
		{"hr.carol@startuphub.com", "Full Stack Developer", "Join our engineering team to build the future of work. Work across the entire stack with React, Node.js, and PostgreSQL.", []string{"React", "Node.js", "PostgreSQL", "GraphQL"}, "Austin, TX", 130000, 170000, "hybrid"},
		{"hr.carol@startuphub.com", "Backend Engineer", "Scale backend systems to handle rapid growth. Experience with Java, Spring Boot, and event-driven architecture preferred.", []string{"Java", "Spring Boot", "Kafka", "Redis"}, "Austin, TX", 140000, 180000, "hybrid"},
		{"hr.david@enterprise.com", "Data Scientist", "Drive data-informed decisions using machine learning and statistical analysis. Build predictive models at enterprise scale.", []string{"Python", "TensorFlow", "SQL", "Statistics"}, "Chicago, IL", 160000, 220000, "hybrid"},
		{"hr.david@enterprise.com", "Machine Learning Engineer", "Build AI-powered products and services. Push the boundaries of what's possible with large-scale ML systems.", []string{"Python", "PyTorch", "Kubernetes", "MLOps"}, "Remote", 180000, 250000, "remote"},
		{"hr.frank@innovation.com", "AI Research Scientist", "Push the boundaries of AI research. Work on cutting-edge problems in natural language processing and computer vision.", []string{"Python", "Research", "Publications", "NLP"}, "Boston, MA", 200000, 300000, "hybrid"},
		{"hr.frank@innovation.com", "Product Manager", "Lead product strategy for our AI platform. Combine technical knowledge with business acumen to drive product success.", []string{"Strategy", "Agile", "Analytics", "SQL"}, "Boston, MA", 150000, 200000, "hybrid"},
	}

	var createdJobs []model.Job
	for _, j := range jobsData {
		agent, ok := agents[j.recruiterEmail]
		if !ok {
			continue
		}

		structuredJSON, _ := json.Marshal(map[string]interface{}{
			"title":            j.title,
			"description":      j.description,
			"skills":           j.requirements,
			"location":         j.location,
			"salary_range": map[string]interface{}{
				"min":      j.salaryMin,
				"max":      j.salaryMax,
				"currency": "USD",
			},
			"work_type":        j.workType,
			"experience_years": 5,
		})

		job := model.Job{
			ID:              uuid.New(),
			AgentID:         agent.ID,
			StructuredJSON:  json.RawMessage(structuredJSON),
			Status:          model.JobStatusActive,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		db.Create(&job)
		createdJobs = append(createdJobs, job)
	}

	return createdJobs
}

func createMatches(db *gorm.DB, agents map[string]model.Agent, jobs []model.Job) []model.Match {
	type matchSpec struct {
		seekerEmail  string
		jobIndex     int
		score        float64
		status       model.MatchStatus
		reasoning    string
	}

	matchSpecs := []matchSpec{
		{"seeker.frank@email.com", 0, 0.92, model.MatchStatusPending, "Strong alignment: 8 years Go/Python/Kubernetes, extensive AWS experience matches requirements for Senior SWE role"},
		{"seeker.frank@email.com", 1, 0.75, model.MatchStatusPending, "Partial DevOps alignment: Go experience relevant, but seeking more infrastructure-focused role"},
		{"seeker.frank@email.com", 3, 0.88, model.MatchStatusPending, "Excellent Go/Rust backend match: 8 years microservices experience directly aligns with Senior Backend Engineer requirements"},
		{"seeker.frank@email.com", 4, 0.82, model.MatchStatusMutualInterest, "Strong SRE alignment: 8 years Go experience + DevOps skills match well with SRE role"},
		{"seeker.grace@email.com", 2, 0.88, model.MatchStatusPending, "Excellent frontend skills: React + TypeScript experience directly aligns with job needs"},
		{"seeker.grace@email.com", 3, 0.78, model.MatchStatusPending, "Partial match: React experience relevant but role focuses on backend microservices"},
		{"seeker.henry@email.com", 1, 0.95, model.MatchStatusPending, "Perfect DevOps match: 6 years Terraform/Docker/CI/CD experience exceeds requirements"},
		{"seeker.henry@email.com", 4, 0.90, model.MatchStatusNegotiating, "Excellent SRE fit: strong infrastructure background + Kubernetes + monitoring experience"},
		{"seeker.henry@email.com", 6, 0.65, model.MatchStatusPending, "Limited ML alignment: background more DevOps-focused than ML-focused"},
		{"seeker.ivy@email.com", 2, 0.91, model.MatchStatusPending, "Great fit: 4 years React/Vue/Figma experience, strong design background matches Frontend Engineer role"},
		{"seeker.ivy@email.com", 3, 0.72, model.MatchStatusPending, "Frontend skills less relevant for backend microservices role"},
		{"seeker.jack@email.com", 4, 0.94, model.MatchStatusNegotiating, "Excellent backend match: 7 years Java/Spring Boot/Kafka experience highly relevant for Backend Engineer role"},
		{"seeker.jack@email.com", 5, 0.85, model.MatchStatusOffered, "Good backend match with Java/Kafka background"},
		{"seeker.kate@email.com", 5, 0.89, model.MatchStatusOffered, "Strong data science background: Python/TensorFlow/SQL expertise directly matches Data Scientist requirements"},
		{"seeker.kate@email.com", 6, 0.85, model.MatchStatusMutualInterest, "Good ML potential: Python/PyTorch background relevant for ML Engineer role"},
		{"seeker.liam@email.com", 3, 0.80, model.MatchStatusPending, "PM background plus SQL skills could work for tech lead role"},
		{"seeker.liam@email.com", 8, 0.87, model.MatchStatusPending, "Good PM fit: Product strategy experience + SQL skills align with Product Manager requirements"},
	}

	var matches []model.Match
	for _, spec := range matchSpecs {
		seekerAgent, ok := agents[spec.seekerEmail]
		if !ok {
			continue
		}

		if spec.jobIndex >= len(jobs) {
			continue
		}
		job := jobs[spec.jobIndex]

		match := model.Match{
			ID:              uuid.New(),
			SeekerAgentID:   seekerAgent.ID,
			RecruiterAgentID: &job.AgentID,
			JobID:           job.ID,
			Score:           spec.score,
			Status:          spec.status,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		db.Create(&match)
		matches = append(matches, match)
	}

	return matches
}

func createMessages(db *gorm.DB, agents map[string]model.Agent, matches []model.Match) {
	// No hardcoded seed messages - AI agents generate real conversation autonomously
	log.Printf("createMessages: skipped hardcoded messages (A2A agents generate real dialogue)")
}

func createInterviews(db *gorm.DB, matches []model.Match) {
	if len(matches) < 4 {
		return
	}

	interviews := []struct {
		matchIndex    int
		daysFromNow  int
		format       model.InterviewFormat
		status       model.InterviewStatus
		location     string
	}{
		{3, 5, model.InterviewFormatVideo, model.InterviewStatusScheduled, "Video Call - Zoom"},
		{8, -2, model.InterviewFormatOnsite, model.InterviewStatusCompleted, "StartupHub HQ, Austin TX"},
		{9, 7, model.InterviewFormatVideo, model.InterviewStatusScheduled, "Video Call - Google Meet"},
	}

	for _, i := range interviews {
		if i.matchIndex >= len(matches) {
			continue
		}

		scheduledAt := time.Now().AddDate(0, 0, i.daysFromNow)
		interview := model.Interview{
			ID:          uuid.New(),
			MatchID:     matches[i.matchIndex].ID,
			ScheduledAt: scheduledAt,
			Format:      i.format,
			Location:    i.location,
			Status:      i.status,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if i.status == model.InterviewStatusCompleted {
			feedback := map[string]interface{}{
				"rating":   4.5,
				"strengths": []string{"Technical depth", "Communication"},
				"concerns":  []string{},
			}
			feedbackJSON, _ := json.Marshal(feedback)
			interview.Feedback = json.RawMessage(feedbackJSON)
		}

		db.Create(&interview)
	}
}

func createOffers(db *gorm.DB, matches []model.Match) {
	if len(matches) < 9 {
		return
	}

	offers := []struct {
		matchIndex      int
		baseSalary      int
		currency       string
		bonus          int
		equityShares   int
		status         model.OfferStatus
		expiresInHours int
	}{
		{9, 180000, "USD", 15000, 5000, model.OfferStatusPending, 72},
	}

	for _, o := range offers {
		if o.matchIndex >= len(matches) {
			continue
		}

		compensationJSON, _ := json.Marshal(map[string]interface{}{
			"base_salary": o.baseSalary,
			"currency":    o.currency,
			"bonus": map[string]interface{}{
				"amount":      o.bonus,
				"description": "Annual performance bonus",
			},
			"equity": map[string]interface{}{
				"shares":         o.equityShares,
				"vesting_period": "4 years",
			},
			"benefits": []string{"Health insurance", "401k matching", "Unlimited PTO"},
		})

		startDate := time.Now().AddDate(0, 1, 0)

		offer := model.Offer{
			ID:               uuid.New(),
			MatchID:          matches[o.matchIndex].ID,
			CompensationJSON: json.RawMessage(compensationJSON),
			StartDate:        startDate,
			Status:           o.status,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		db.Create(&offer)
	}
}

func printSummary(users []model.User, agents map[string]model.Agent, jobs []model.Job, matches []model.Match) {
	fmt.Println("\n========== SEED DATA SUMMARY ==========")
	fmt.Printf("Organizations: 4 (TechCorp Inc., StartupHub, Enterprise Co., Innovation Labs)\n")
	fmt.Printf("Users: %d total\n", len(users))
	fmt.Printf("  - Recruiters: 7\n")
	fmt.Printf("  - Seekers: 7\n")
	fmt.Printf("Agents: %d total\n", len(agents))
	fmt.Printf("Jobs: %d total\n", len(jobs))
	fmt.Printf("Matches: %d total\n", len(matches))
	fmt.Println("\nLogin credentials (all use password: password123):")
	fmt.Println("  Recruiters:")
	fmt.Println("    hr.alice@techcorp.com, hr.bob@techcorp.com, hr.carol@startuphub.com, hr.david@enterprise.com, hr.frank@innovation.com")
	fmt.Println("  Seekers:")
	fmt.Println("    seeker.frank@email.com, seeker.grace@email.com, seeker.henry@email.com, seeker.ivy@email.com")
	fmt.Println("    seeker.jack@email.com, seeker.kate@email.com, seeker.liam@email.com")
	fmt.Println("========================================")
}
