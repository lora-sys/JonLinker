// Package testutil provides shared test utilities for integration tests.
// All integration tests must use these helpers to ensure consistent setup.
package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ENV_MISSING_MSG is used when a required environment variable is not set.
const ENV_MISSING_MSG = "required environment variable not set — test must t.Fatal"

// RandomEmail generates a unique test email address.
func RandomEmail() string {
	return fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8])
}

// RandomUUID returns a new random UUID string.
func RandomUUID() string {
	return uuid.New().String()
}

// RandomSkills returns a random subset of common skills.
func RandomSkills() []string {
	all := []string{"go", "python", "javascript", "typescript", "react", "node", "postgresql", "redis", "kubernetes", "docker", "aws", "gcp"}
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(3) + 2
	selected := make([]string, n)
	for i := range selected {
		selected[i] = all[rand.Intn(len(all))]
	}
	return selected
}

// RandomRole returns either "seeker" or "recruiter".
func RandomRole() string {
	if rand.Intn(2) == 0 {
		return "seeker"
	}
	return "recruiter"
}

// RequireEnv fails the test if the given environment variable is not set.
func RequireEnv(t *testing.T, key string) string {
	val := os.Getenv(key)
	if val == "" {
		t.Fatalf("required environment variable %s is not set — %s", key, ENV_MISSING_MSG)
	}
	return val
}

// RequireServerURL returns TEST_SERVER_URL or fails.
func RequireServerURL(t *testing.T) string {
	return RequireEnv(t, "TEST_SERVER_URL")
}

// RequireDatabaseURL returns TEST_DATABASE_URL or fails.
func RequireDatabaseURL(t *testing.T) string {
	return RequireEnv(t, "TEST_DATABASE_URL")
}

// RequireJWTSecret returns JWT_SECRET or fails.
func RequireJWTSecret(t *testing.T) string {
	return RequireEnv(t, "JWT_SECRET")
}

// HTTPClient is a shared HTTP client for tests.
var HTTPClient = &http.Client{Timeout: 30 * time.Second}

// RegisterRequest is the POST /api/auth/register payload.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// RegisterResponse is the POST /api/auth/register response.
type RegisterResponse struct {
	Token        string `json:"token"`
	UserID       string `json:"user_id,omitempty"`
	Email        string `json:"email,omitempty"`
	PasswordHash string `json:"password_hash,omitempty"`
}

// MustRegister registers a new user and returns (token, userID).
// Fails the test if registration fails.
func MustRegister(t *testing.T, baseURL, role string) (token string, userID string) {
	email := RandomEmail()
	payload := RegisterRequest{Email: email, Password: "password123", Role: role}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal register payload: %v", err)
	}

	resp, err := HTTPClient.Post(baseURL+"/api/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to POST /api/auth/register: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		t.Fatalf("expected 201 from /api/auth/register, got %d", resp.StatusCode)
	}

	var result RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode register response: %v", err)
	}

	if result.Token == "" || len(result.Token) < 20 {
		t.Fatalf("expected non-empty token with len > 20, got %q (len=%d)", result.Token, len(result.Token))
	}

	if result.PasswordHash != "" {
		t.Fatal("FAKE: password_hash must not be present in register response")
	}

	return result.Token, result.UserID
}

// MustLogin attempts login and returns token. Fails if login fails.
func MustLogin(t *testing.T, baseURL, email, password string) string {
	payload := map[string]string{"email": email, "password": password}
	body, _ := json.Marshal(payload)
	resp, err := HTTPClient.Post(baseURL+"/api/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to POST /api/auth/login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 from /api/auth/login, got %d", resp.StatusCode)
	}

	var result RegisterResponse
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Token == "" {
		t.Fatal("expected non-empty token from login")
	}
	return result.Token
}

// AgentPayload is the POST /api/agents request body.
type AgentPayload struct {
	Name     string   `json:"name"`
	AgentType string  `json:"agent_type"`
	Skills   []string `json:"skills,omitempty"`
}

// MustCreateAgent creates an agent and returns its ID. Fails on error.
func MustCreateAgent(t *testing.T, baseURL, token, agentType string) string {
	payload := AgentPayload{
		Name:      fmt.Sprintf("Agent-%s", RandomUUID()[:8]),
		AgentType: agentType,
		Skills:    RandomSkills(),
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", baseURL+"/api/agents", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to POST /api/agents: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		t.Fatalf("expected 201 from /api/agents, got %d", resp.StatusCode)
	}

	var agent map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&agent); err != nil {
		t.Fatalf("failed to decode agent response: %v", err)
	}

	id, ok := agent["id"].(string)
	if !ok || id == "" {
		t.Fatal("expected non-empty agent id")
	}
	return id
}

// JobPayload is the POST /api/jobs request body.
type JobPayload struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Skills      []string `json:"skills,omitempty"`
	SalaryMin   int     `json:"salary_min,omitempty"`
	SalaryMax   int     `json:"salary_max,omitempty"`
	Location    string  `json:"location,omitempty"`
	IsRemote    bool    `json:"is_remote,omitempty"`
}

// MustCreateJob creates a job and returns its ID. Fails on error.
func MustCreateJob(t *testing.T, baseURL, token string) string {
	rand.Seed(time.Now().UnixNano())
	// Use random salary to avoid hardcoded test values
	salaryMin := 80000 + rand.Intn(50000)
	salaryMax := salaryMin + 20000 + rand.Intn(30000)

	payload := JobPayload{
		Title:       fmt.Sprintf("Engineer-%s", RandomUUID()[:8]),
		Description: fmt.Sprintf("Looking for a skilled developer with experience in %s", strings.Join(RandomSkills(), ", ")),
		Skills:      RandomSkills(),
		SalaryMin:   salaryMin,
		SalaryMax:   salaryMax,
		Location:    []string{"San Francisco", "New York", "London", "Remote"}[rand.Intn(4)],
		IsRemote:    rand.Intn(2) == 0,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", baseURL+"/api/jobs", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to POST /api/jobs: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		t.Fatalf("expected 201 from /api/jobs, got %d", resp.StatusCode)
	}

	var job map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		t.Fatalf("failed to decode job response: %v", err)
	}

	id, ok := job["id"].(string)
	if !ok || id == "" {
		t.Fatal("expected non-empty job id")
	}
	return id
}

// MustCreateMatch creates a match between seeker and recruiter agents.
func MustCreateMatch(t *testing.T, baseURL, seekerToken, recruiterToken, seekerAgentID, recruiterAgentID, jobID string) string {
	payload := map[string]interface{}{
		"seeker_agent_id":    seekerAgentID,
		"recruiter_agent_id": recruiterAgentID,
		"job_id":             jobID,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", baseURL+"/api/matches/auto", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+recruiterToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to POST /api/matches/auto: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		t.Fatalf("expected 201 from /api/matches/auto, got %d", resp.StatusCode)
	}

	var match map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&match); err != nil {
		t.Fatalf("failed to decode match response: %v", err)
	}

	id, ok := match["id"].(string)
	if !ok || id == "" {
		t.Fatal("expected non-empty match id")
	}
	return id
}
