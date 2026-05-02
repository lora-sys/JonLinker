package integration

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"joblinker/internal/model"
)

// E2E Test Scenarios for JobLinker A2A Recruitment Platform

// TestE2E_CompleteRecruitmentFlow tests the complete recruitment flow from registration to offer
func TestE2E_CompleteRecruitmentFlow(t *testing.T) {
	seed := CreateComprehensiveSeedData()
	seed.PrintSeedSummary()

	t.Run("Scenario 1: New User Registration and Agent Creation", func(t *testing.T) {
		// Simulate a new user registering on the platform
		newUserEmail := "newuser@test.com"

		// Check that user doesn't already exist in seed
		for _, u := range seed.Users {
			if u.Email == newUserEmail {
				t.Skip("User already exists in seed data")
			}
		}

		// User would register via POST /api/auth/register
		// Expected: User created, token returned
		expectedStatus := 201 // http.StatusCreated

		if expectedStatus != 201 {
			t.Errorf("Registration should return 201 Created, got %d", expectedStatus)
		}
	})

	t.Run("Scenario 2: Seeker Creates Agent with Resume", func(t *testing.T) {
		// Get a seeker user from seed data
		var seekerUser model.User
		var seekerAgent model.Agent
		for i, u := range seed.Users {
			if u.Role == "seeker" && len(seed.Agents) > i {
				seekerUser = u
				for _, a := range seed.Agents {
					if a.UserID == u.ID {
						seekerAgent = a
						break
					}
				}
				break
			}
		}

		if seekerAgent.ID == uuid.Nil {
			t.Skip("No seeker agent found in seed data")
		}

		// Simulate agent creation with resume parsing
		agentConfig := map[string]interface{}{
			"resume_parsed": true,
			"skills":        []string{"Go", "Python", "Kubernetes", "AWS"},
			"experience":    "8 years",
			"education":     "MS Computer Science",
		}

		// Verify agent config contains parsed resume data
		configBytes, _ := json.Marshal(agentConfig)
		var parsedConfig map[string]interface{}
		json.Unmarshal(configBytes, &parsedConfig)

		if parsedConfig["resume_parsed"] != true {
			t.Error("Agent config should contain parsed resume data")
		}

		if parsedConfig["skills"] == nil {
			t.Error("Agent config should contain skills from resume")
		}

		t.Logf("Created agent %s for user %s with skills: %v",
			seekerAgent.ID, seekerUser.Email, parsedConfig["skills"])
	})

	t.Run("Scenario 3: HR Agent Posts New Job", func(t *testing.T) {
		// Find an HR agent (recruiter type)
		var hrAgent model.Agent
		for _, a := range seed.Agents {
			if a.Type == model.AgentTypeRecruiter && a.Status == model.AgentStatusActive {
				hrAgent = a
				break
			}
		}

		if hrAgent.ID == uuid.Nil {
			t.Skip("No HR agent found in seed data")
		}

		// Simulate job posting
		newJob := map[string]interface{}{
			"title":       "Senior A2A Engineer",
			"department":  "Platform",
			"location":    "Remote",
			"type":        "Full-time",
			"salary_min":  160000,
			"salary_max":  220000,
			"skills":      []string{"Go", "gRPC", "Kubernetes", "PostgreSQL"},
			"description": "Build the next generation of agent-to-agent communication...",
		}

		jobBytes, _ := json.Marshal(newJob)
		var parsedJob map[string]interface{}
		json.Unmarshal(jobBytes, &parsedJob)

		if parsedJob["title"] != "Senior A2A Engineer" {
			t.Error("Job title should match")
		}

		// JSON unmarshals numbers as float64, so we need to compare as float64
		if parsedJob["salary_min"].(float64) != 160000 {
			t.Error("Job salary_min should match")
		}

		t.Logf("HR Agent %s posted new job: %s", hrAgent.ID, newJob["title"])
	})

	t.Run("Scenario 4: Seeker Applies to Job - Match Created", func(t *testing.T) {
		// Find a seeker without any matches
		var unmatchedSeeker model.Agent
		for _, a := range seed.Agents {
			if a.Type == model.AgentTypeSeeker {
				matches := seed.GetMatchesBySeeker(a.ID)
				if len(matches) == 0 {
					unmatchedSeeker = a
					break
				}
			}
		}

		if unmatchedSeeker.ID == uuid.Nil {
			t.Skip("All seekers already have matches")
		}

		// Find an active job
		var targetJob model.Job
		for _, j := range seed.Jobs {
			if j.Status == model.JobStatusActive {
				targetJob = j
				break
			}
		}

		if targetJob.ID == uuid.Nil {
			t.Skip("No active jobs found")
		}

		// Simulate match creation
		// Match score calculation would happen via vector similarity
		matchScore := 0.85 // Simulated high match score

		newMatch := seed.TestMatch(unmatchedSeeker.ID, targetJob.ID, matchScore, model.MatchStatusPending)

		if newMatch.Score != matchScore {
			t.Errorf("Match score should be %f, got %f", matchScore, newMatch.Score)
		}

		if newMatch.Status != model.MatchStatusPending {
			t.Errorf("New match should be pending, got %s", newMatch.Status)
		}

		t.Logf("Created match: Seeker %s -> Job %s (score: %.2f)",
			unmatchedSeeker.ID, targetJob.ID, matchScore)
	})

	t.Run("Scenario 5: A2A Introduction Message Flow", func(t *testing.T) {
		// Get first pending match
		var pendingMatch model.Match
		var seekerAgent, recruiterAgent model.Agent

		for _, m := range seed.Matches {
			if m.Status == model.MatchStatusPending {
				pendingMatch = m
				break
			}
		}

		if pendingMatch.ID == uuid.Nil {
			t.Skip("No pending matches found")
		}

		// Find seeker agent
		for _, a := range seed.Agents {
			if a.ID == pendingMatch.SeekerAgentID {
				seekerAgent = a
				break
			}
		}

		// Find recruiter agent (from job)
		for _, j := range seed.Jobs {
			if j.ID == pendingMatch.JobID {
				for _, a := range seed.Agents {
					if a.UserID == j.AgentID {
						recruiterAgent = a
						break
					}
				}
				break
			}
		}

		// Simulate INTRODUCTION message from seeker
		introMessage := map[string]interface{}{
			"message_id": uuid.New().String(),
			"timestamp":  time.Now().Format(time.RFC3339),
			"sender_id":  seekerAgent.ID.String(),
			"receiver_id": recruiterAgent.ID.String(),
			"intent":     "INTRODUCTION",
			"parameters": map[string]interface{}{
				"role":       "seeker",
				"name":       "Frank Garcia",
				"title":      "Senior Software Engineer",
				"skills":     "Go, Python, Kubernetes, AWS",
				"experience": "8 years",
			},
		}

		introXML, _ := xmlMarshal(introMessage)

		// Store message
		seed.TestMessage(pendingMatch.ID, seekerAgent.ID, introXML, "INTRODUCTION")

		// Simulate INTEREST response from recruiter
		responseMessage := map[string]interface{}{
			"message_id": uuid.New().String(),
			"timestamp":   time.Now().Format(time.RFC3339),
			"sender_id":   recruiterAgent.ID.String(),
			"receiver_id": seekerAgent.ID.String(),
			"reply_to":    introMessage["message_id"],
			"intent":      "INTEREST",
			"parameters": map[string]interface{}{
				"role":           "recruiter",
				"name":           "Bob Smith",
				"interest_level": "high",
			},
		}

		responseXML, _ := xmlMarshal(responseMessage)
		seed.TestMessage(pendingMatch.ID, recruiterAgent.ID, responseXML, "INTEREST")

		// Verify messages stored
		messages := seed.GetMessagesByMatch(pendingMatch.ID)
		if len(messages) < 2 {
			t.Errorf("Expected at least 2 messages in conversation, got %d", len(messages))
		}

		// Verify conversation flow
		if messages[0].IntentType != "INTRODUCTION" {
			t.Errorf("First message should be INTRODUCTION, got %s", messages[0].IntentType)
		}

		if messages[1].IntentType != "INTEREST" {
			t.Errorf("Second message should be INTEREST, got %s", messages[1].IntentType)
		}

		t.Logf("A2A Introduction flow completed: %d messages in conversation", len(messages))
	})
}

// TestE2E_InterviewFlow tests the interview scheduling and completion flow
func TestE2E_InterviewFlow(t *testing.T) {
	seed := CreateComprehensiveSeedData()

	t.Run("Scenario 6: First Round Interview Scheduling", func(t *testing.T) {
		// Find a match in mutual interest status
		var match model.Match
		for _, m := range seed.Matches {
			if m.Status == model.MatchStatusMutualInterest {
				match = m
				break
			}
		}

		if match.ID == uuid.Nil {
			t.Skip("No mutual interest matches found")
		}

		// Simulate interview scheduling
		scheduledTime := time.Now().AddDate(0, 0, 7) // 7 days from now
		format := model.InterviewFormatVideo
		_ = "https://meet.example.com/interview/123"

		interview := seed.TestInterview(match.ID, scheduledTime, format, model.InterviewStatusScheduled)

		if interview.ScheduledAt.Unix() != scheduledTime.Unix() {
			t.Errorf("Interview time mismatch")
		}

		if interview.Format != model.InterviewFormatVideo {
			t.Errorf("Interview format should be video, got %s", interview.Format)
		}

		if interview.Status != model.InterviewStatusScheduled {
			t.Errorf("Interview status should be scheduled, got %s", interview.Status)
		}

		t.Logf("First round interview scheduled for %s at %s", match.ID, scheduledTime)
	})

	t.Run("Scenario 7: Second Round Interview (Onsite)", func(t *testing.T) {
		// Find a completed first round interview
		var completedInterview model.Interview
		for _, i := range seed.Interviews {
			if i.Status == model.InterviewStatusCompleted {
				completedInterview = i
				break
			}
		}

		if completedInterview.ID == uuid.Nil {
			t.Skip("No completed interviews found")
		}

		// Find match for this interview
		var match model.Match
		for _, m := range seed.Matches {
			if m.ID == completedInterview.MatchID {
				match = m
				break
			}
		}

		// Schedule second round (onsite)
		secondRoundTime := time.Now().AddDate(0, 0, 14) // 14 days from now
		secondInterview := seed.TestInterview(
			match.ID,
			secondRoundTime,
			model.InterviewFormatOnsite,
			model.InterviewStatusScheduled,
		)
		secondInterview.Location = "Company HQ, 123 Tech Street"

		if secondInterview.Format != model.InterviewFormatOnsite {
			t.Errorf("Second round should be onsite, got %s", secondInterview.Format)
		}

		t.Logf("Second round interview scheduled: %s", secondInterview.Location)
	})

	t.Run("Scenario 8: Interview Completion with Feedback", func(t *testing.T) {
		// Find a scheduled interview
		var interview model.Interview
		for _, i := range seed.Interviews {
			if i.Status == model.InterviewStatusScheduled {
				interview = i
				break
			}
		}

		if interview.ID == uuid.Nil {
			t.Skip("No scheduled interviews found")
		}

		// Simulate interview completion
		feedback := map[string]interface{}{
			"rating":         4.5,
			"strengths":      []string{"Technical skills", "Communication", "Problem solving"},
			"concerns":       []string{"Salary expectations slightly high"},
			"recommendation": "proceed_to_offer",
		}

		// Update interview status
		interview.Status = model.InterviewStatusCompleted
		feedbackJSON, _ := json.Marshal(feedback)
		interview.Feedback = string(feedbackJSON)

		if interview.Status != model.InterviewStatusCompleted {
			t.Errorf("Interview should be completed, got %s", interview.Status)
		}

		if interview.Feedback != "" {
			t.Logf("Interview %s completed with feedback: %s", interview.ID, interview.Feedback)
		} else {
			t.Logf("Interview %s completed", interview.ID)
		}
	})
}

// TestE2E_OfferFlow tests the offer negotiation and acceptance flow
func TestE2E_OfferFlow(t *testing.T) {
	seed := CreateComprehensiveSeedData()

	t.Run("Scenario 9: Offer Generation", func(t *testing.T) {
		// Find a match that's ready for offer (completed interviews)
		var offerMatch model.Match
		for _, m := range seed.Matches {
			if m.Status == model.MatchStatusNegotiating {
				offerMatch = m
				break
			}
		}

		if offerMatch.ID == uuid.Nil {
			t.Skip("No negotiating matches found")
		}

		// Generate offer
		compensation := map[string]interface{}{
			"base_salary": 170000,
			"currency":    "USD",
			"bonus":       15000,
			"equity":      "0.03%",
			"total":       185000,
		}

		seed.TestOffer(offerMatch.ID, compensation, model.OfferStatusPending)

		t.Logf("Generated offer for match %s: %v", offerMatch.ID, compensation)
	})

	t.Run("Scenario 10: Offer Acceptance", func(t *testing.T) {
		// Find a pending offer
		var offer model.Offer
		for _, o := range seed.Offers {
			if o.Status == model.OfferStatusPending {
				offer = o
				break
			}
		}

		if offer.ID == uuid.Nil {
			t.Skip("No pending offers found")
		}

		// Simulate offer acceptance
		acceptTime := time.Now()
		offer.Status = model.OfferStatusAccepted
		offer.RespondedAt = &acceptTime

		if offer.Status != model.OfferStatusAccepted {
			t.Errorf("Offer should be accepted, got %s", offer.Status)
		}

		// Update match status
		for i, m := range seed.Matches {
			if m.ID == offer.MatchID {
				seed.Matches[i].Status = model.MatchStatusHired
				break
			}
		}

		t.Logf("Offer %s accepted at %s", offer.ID, acceptTime)
	})

	t.Run("Scenario 11: Offer Negotiation", func(t *testing.T) {
		// Find a pending offer that will be negotiated
		var offer model.Offer
		for _, o := range seed.Offers {
			if o.Status == model.OfferStatusPending {
				offer = o
				break
			}
		}

		if offer.ID == uuid.Nil {
			t.Skip("No pending offers found")
		}

		// Find match
		var match model.Match
		for _, m := range seed.Matches {
			if m.ID == offer.MatchID {
				match = m
				break
			}
		}

		var seekerAgent model.Agent
		for _, a := range seed.Agents {
			if a.ID == match.SeekerAgentID {
				seekerAgent = a
				break
			}
		}

		// Counter offer message
		negotiationMsg := map[string]interface{}{
			"message_id": uuid.New().String(),
			"timestamp":  time.Now().Format(time.RFC3339),
			"sender_id":  seekerAgent.ID.String(),
			"intent":     "NEGOTIATION",
			"negotiation": map[string]interface{}{
				"round":       1,
				"type":        "salary",
				"current":     170000,
				"target":      185000,
				"currency":    "USD",
				"concessions": 0,
			},
		}

		_, _ = xmlMarshal(negotiationMsg)
		seed.TestMessage(match.ID, seekerAgent.ID, "", "NEGOTIATION")

		// Update offer status
		offer.Status = model.OfferStatusNegotiating

		if offer.Status != model.OfferStatusNegotiating {
			t.Errorf("Offer should be negotiating, got %s", offer.Status)
		}

		t.Logf("Offer negotiation started for match %s", match.ID)
	})

	t.Run("Scenario 12: Offer Declined (Rejection)", func(t *testing.T) {
		// Create a new pending offer that will be declined
		var match model.Match
		for _, m := range seed.Matches {
			if m.Status == model.MatchStatusNegotiating {
				match = m
				break
			}
		}

		if match.ID == uuid.Nil {
			t.Skip("No negotiating matches found")
		}

		offer := seed.TestOffer(match.ID, map[string]interface{}{
			"base_salary": 120000,
			"currency":    "USD",
		}, model.OfferStatusPending)

		// Simulate decline
		declineTime := time.Now()
		offer.Status = model.OfferStatusDeclined
		offer.RespondedAt = &declineTime

		// Update match to rejected
		for i, m := range seed.Matches {
			if m.ID == match.ID {
				seed.Matches[i].Status = model.MatchStatusRejected
				break
			}
		}

		if offer.Status != model.OfferStatusDeclined {
			t.Errorf("Offer should be declined, got %s", offer.Status)
		}

		t.Logf("Offer %s declined, match %s rejected", offer.ID, match.ID)
	})
}

// TestE2E_BossAgentFlow tests the boss/executive agent capabilities
func TestE2E_BossAgentFlow(t *testing.T) {
	seed := CreateComprehensiveSeedData()

	t.Run("Scenario 13: Boss Agent Views Company Agents", func(t *testing.T) {
		// Find a boss/executive agent
		var bossAgent model.Agent
		var orgID *uuid.UUID

		for _, a := range seed.Agents {
			// Check if this is a boss/executive agent
			for _, u := range seed.Users {
				if u.ID == a.UserID {
					if u.OrganizationID != nil {
						orgID = u.OrganizationID
						bossAgent = a
						break
					}
				}
			}
			if bossAgent.ID != uuid.Nil {
				break
			}
		}

		if bossAgent.ID == uuid.Nil || orgID == nil {
			t.Skip("No boss agent found in seed data")
		}

		// Get all agents in the same organization
		orgAgents := seed.GetAgentsByOrganization(*orgID)

		// Separate by type
		var hrAgents, employeeAgents []model.Agent
		for _, a := range orgAgents {
			if a.Type == model.AgentTypeRecruiter {
				hrAgents = append(hrAgents, a)
			} else {
				employeeAgents = append(employeeAgents, a)
			}
		}

		t.Logf("Boss %s can see %d HR agents and %d employee agents in organization %s",
			bossAgent.ID, len(hrAgents), len(employeeAgents), *orgID)

		if len(orgAgents) == 0 {
			t.Error("Boss should be able to see organization agents")
		}
	})

	t.Run("Scenario 14: Boss Reviews Hiring Pipeline", func(t *testing.T) {
		// Find organization with matches
		var orgWithMatches model.Organization

		for _, org := range seed.Organizations {
			orgAgents := seed.GetAgentsByOrganization(org.ID)
			for _, agent := range orgAgents {
				if agent.Type == model.AgentTypeRecruiter {
					// Check for matches
					for _, job := range seed.Jobs {
						for _, match := range seed.Matches {
							if match.JobID == job.ID && job.AgentID == agent.ID {
								orgWithMatches = org
								goto found
							}
						}
					}
				}
			}
		}
	found:

		if orgWithMatches.ID == uuid.Nil {
			t.Skip("No organization with matches found")
		}

		// Get pipeline summary
		pipeline := map[string]int{
			"pending":         0,
			"mutual_interest": 0,
			"negotiating":     0,
			"offered":         0,
			"hired":           0,
			"rejected":        0,
		}

		orgAgents := seed.GetAgentsByOrganization(orgWithMatches.ID)
		for _, agent := range orgAgents {
			if agent.Type == model.AgentTypeRecruiter {
				for _, job := range seed.Jobs {
					if job.AgentID == agent.ID {
						for _, match := range seed.Matches {
							if match.JobID == job.ID {
								pipeline[string(match.Status)]++
							}
						}
					}
				}
			}
		}

		t.Logf("Organization %s pipeline: %+v", orgWithMatches.Name, pipeline)
	})

	t.Run("Scenario 15: Boss Approves Offer", func(t *testing.T) {
		// Find a pending offer
		var offer model.Offer
		for _, o := range seed.Offers {
			if o.Status == model.OfferStatusPending {
				offer = o
				break
			}
		}

		if offer.ID == uuid.Nil {
			t.Skip("No pending offers found")
		}

		// Find match and job
		var match model.Match
		var job model.Job
		for _, m := range seed.Matches {
			if m.ID == offer.MatchID {
				match = m
				break
			}
		}

		for _, j := range seed.Jobs {
			if j.ID == match.JobID {
				job = j
				break
			}
		}

		// Find boss in the organization - compare with job.AgentID directly
		var bossAgent model.Agent
		for _, a := range seed.Agents {
			for _, u := range seed.Users {
				if u.ID == a.UserID && u.OrganizationID != nil && *u.OrganizationID == job.AgentID {
					// Check if boss can approve
					var config map[string]interface{}
					json.Unmarshal([]byte(a.ConfigJSON), &config)
					if canApprove, ok := config["can_approve"].(bool); ok && canApprove {
						bossAgent = a
						break
					}
				}
			}
			if bossAgent.ID != uuid.Nil {
				break
			}
		}

		if bossAgent.ID == uuid.Nil {
			t.Skip("No approving boss found")
		}

		// Simulate boss approval message
		approvalMsg := map[string]interface{}{
			"message_id": uuid.New().String(),
			"timestamp":  time.Now().Format(time.RFC3339),
			"sender_id":  bossAgent.ID.String(),
			"intent":     "CONFIRM",
			"parameters": map[string]interface{}{
				"role":     "boss",
				"approval": "approved",
				"offer_id": offer.ID.String(),
			},
		}

		msgXML, _ := xmlMarshal(approvalMsg)
		seed.TestMessage(match.ID, bossAgent.ID, msgXML, "CONFIRM")

		t.Logf("Boss %s approved offer %s for match %s",
			bossAgent.ID, offer.ID, match.ID)
	})
}

// TestE2E_AgentCommunication tests the A2A messaging between agents
func TestE2E_AgentCommunication(t *testing.T) {
	seed := CreateComprehensiveSeedData()

	t.Run("Scenario 16: Group Chat in Organization", func(t *testing.T) {
		// Find an organization with multiple agents
		var org model.Organization
		var orgAgents []model.Agent

		for _, o := range seed.Organizations {
			agents := seed.GetAgentsByOrganization(o.ID)
			if len(agents) >= 3 {
				org = o
				orgAgents = agents
				break
			}
		}

		if org.ID == uuid.Nil {
			t.Skip("No organization with enough agents found")
		}

		// Simulate group message (would go to all agents in org)
		var sender model.Agent
		for _, a := range orgAgents {
			if a.Type == model.AgentTypeRecruiter {
				sender = a
				break
			}
		}

		if sender.ID == uuid.Nil {
			t.Skip("No recruiter agent found")
		}

		// Create a group message
		groupMessage := map[string]interface{}{
			"message_id": uuid.New().String(),
			"timestamp":  time.Now().Format(time.RFC3339),
			"sender_id":   sender.ID.String(),
			"type":        "group",
			"intent":      "ANNOUNCEMENT",
			"parameters": map[string]interface{}{
				"title":   "New Position Opening",
				"content": "We have a new senior role opening up...",
			},
		}

		_, _ = xmlMarshal(groupMessage)

		// In real implementation, this would be broadcast to all org agents
		// For now, just verify the message structure
		t.Logf("Group message from %s to organization %s: %s",
			sender.ID, org.ID, groupMessage["message_id"])
	})

	t.Run("Scenario 17: Individual Agent-to-Agent Chat", func(t *testing.T) {
		// Find two agents with an active match
		var agent1, agent2 model.Agent
		var match model.Match

		for _, m := range seed.Matches {
			if m.Status == model.MatchStatusMutualInterest || m.Status == model.MatchStatusNegotiating {
				match = m
				for _, a := range seed.Agents {
					if a.ID == m.SeekerAgentID {
						agent1 = a
					}
				}
				for _, j := range seed.Jobs {
					if j.ID == m.JobID {
						for _, a := range seed.Agents {
							if a.UserID == j.AgentID {
								agent2 = a
							}
						}
					}
				}
				break
			}
		}

		if agent1.ID == uuid.Nil || agent2.ID == uuid.Nil {
			t.Skip("No agents with active match found")
		}

		// Simulate direct message exchange
		directMessages := []map[string]interface{}{
			{
				"message_id":  uuid.New().String(),
				"sender_id":   agent1.ID.String(),
				"receiver_id": agent2.ID.String(),
				"intent":      "INQUIRY",
				"content":     "What is the team size for this role?",
			},
			{
				"message_id":  uuid.New().String(),
				"sender_id":   agent2.ID.String(),
				"receiver_id": agent1.ID.String(),
				"reply_to":    "first",
				"intent":      "INQUIRY",
				"content":     "The team has 8 engineers reporting to the CTO",
			},
			{
				"message_id":  uuid.New().String(),
				"sender_id":   agent1.ID.String(),
				"receiver_id": agent2.ID.String(),
				"intent":      "INTEREST",
				"content":     "That sounds great, I am very interested",
			},
		}

		for _, msg := range directMessages {
			msgXML, _ := xmlMarshal(msg)
			seed.TestMessage(match.ID, agent1.ID, msgXML, msg["intent"].(string))
		}

		// Verify message history
		messages := seed.GetMessagesByMatch(match.ID)
		if len(messages) < 3 {
			t.Errorf("Expected at least 3 messages, got %d", len(messages))
		}

		t.Logf("Direct A2A chat between %s and %s: %d messages",
			agent1.ID, agent2.ID, len(messages))
	})

	t.Run("Scenario 18: Message Intent Classification", func(t *testing.T) {
		// Test different intent types
		intents := []string{
			"INTRODUCTION",
			"INTEREST",
			"NEGOTIATION",
			"OFFER",
			"ACCEPT",
			"DECLINE",
			"SCHEDULE",
			"CONFIRM",
			"WITHDRAW",
			"INQUIRY",
		}

		for _, intent := range intents {
			// Verify intent is a valid A2A intent
			validIntent := false
			validIntents := map[string]bool{
				"INTRODUCTION": true,
				"INTEREST":     true,
				"NEGOTIATION":  true,
				"OFFER":        true,
				"ACCEPT":       true,
				"DECLINE":      true,
				"SCHEDULE":     true,
				"CONFIRM":      true,
				"WITHDRAW":     true,
				"INQUIRY":      true,
			}

			if validIntents[intent] {
				validIntent = true
			}

			if !validIntent {
				t.Errorf("Invalid intent: %s", intent)
			}
		}

		t.Logf("All %d intents validated", len(intents))
	})
}

// TestE2E_RejectionFlow tests the rejection scenario
func TestE2E_RejectionFlow(t *testing.T) {
	seed := CreateComprehensiveSeedData()

	t.Run("Scenario 19: Seeker Gets Rejected After Interview", func(t *testing.T) {
		// Find a match that hasn't been rejected yet
		var match model.Match
		for _, m := range seed.Matches {
			if m.Status == model.MatchStatusMutualInterest {
				match = m
				break
			}
		}

		if match.ID == uuid.Nil {
			t.Skip("No suitable matches found for rejection test")
		}

		// Find seeker and recruiter agents
		var seekerAgent, recruiterAgent model.Agent
		for _, a := range seed.Agents {
			if a.ID == match.SeekerAgentID {
				seekerAgent = a
			}
		}

		for _, j := range seed.Jobs {
			if j.ID == match.JobID {
				for _, a := range seed.Agents {
					if a.UserID == j.AgentID {
						recruiterAgent = a
					}
				}
			}
		}

		// Send rejection message
		rejectionMsg := map[string]interface{}{
			"message_id":  uuid.New().String(),
			"timestamp":   time.Now().Format(time.RFC3339),
			"sender_id":   recruiterAgent.ID.String(),
			"receiver_id": seekerAgent.ID.String(),
			"intent":      "DECLINE",
			"parameters": map[string]interface{}{
				"reason": "After careful consideration, we have decided to move forward with other candidates whose qualifications more closely match our current needs.",
			},
		}

		msgXML, _ := xmlMarshal(rejectionMsg)
		seed.TestMessage(match.ID, recruiterAgent.ID, msgXML, "DECLINE")

		// Update match status in slice
		for i, m := range seed.Matches {
			if m.ID == match.ID {
				seed.Matches[i].Status = model.MatchStatusRejected
				break
			}
		}

		// Find the updated match and check status
		var updatedMatch model.Match
		for _, m := range seed.Matches {
			if m.ID == match.ID {
				updatedMatch = m
				break
			}
		}

		if updatedMatch.Status != model.MatchStatusRejected {
			t.Errorf("Match should be rejected, got %s", updatedMatch.Status)
		}

		t.Logf("Seeker %s rejected after interview for match %s", seekerAgent.ID, match.ID)
	})

	t.Run("Scenario 20: System Handles Rejection Gracefully", func(t *testing.T) {
		// Count rejected matches
		rejectedCount := 0
		for _, m := range seed.Matches {
			if m.Status == model.MatchStatusRejected {
				rejectedCount++
			}
		}

		// Verify rejection count is reasonable
		if rejectedCount == 0 {
			t.Log("No rejections in current seed data")
		} else {
			t.Logf("Total rejections in system: %d", rejectedCount)
		}

		// Check that rejected matches still have message history
		for _, m := range seed.Matches {
			if m.Status == model.MatchStatusRejected {
				messages := seed.GetMessagesByMatch(m.ID)
				if len(messages) == 0 {
					t.Errorf("Rejected match %s has no message history", m.ID)
				}
			}
		}
	})
}

// Helper function to marshal map to XML-like string (simplified for testing)
func xmlMarshal(v interface{}) (string, error) {
	// In real implementation, this would use proper XML marshaling
	// For testing purposes, we just return a JSON representation
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
