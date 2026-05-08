//go:build integration
// +build integration

// Real integration tests for auth endpoints.
// Uses real HTTP requests to TEST_SERVER_URL. No anonymous handlers.
// No t.Skip — missing env vars cause t.Fatal.
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"joblinker/tests/testutil"
)

func TestRegister_Real(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)

	// Use random email to avoid collision
	email := testutil.RandomEmail()
	payload := testutil.RegisterRequest{
		Email:    email,
		Password: "password123",
		Role:     "seeker",
	}
	body, _ := json.Marshal(payload)

	resp, err := testutil.HTTPClient.Post(baseURL+"/api/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to POST /api/auth/register: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	token, ok := result["token"].(string)
	if !ok || len(token) < 20 {
		t.Errorf("expected token string with len > 20, got %v", result["token"])
	}

	if _, exists := result["password_hash"]; exists {
		t.Fatal("FAKE: password_hash must not be present in register response")
	}

	// Verify token works for a protected endpoint
	if resp.StatusCode == 201 {
		req, _ := http.NewRequest("GET", baseURL+"/api/agents", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		agentResp, _ := testutil.HTTPClient.Do(req)
		if agentResp.StatusCode != 200 {
			t.Errorf("token should be valid for /api/agents, got %d", agentResp.StatusCode)
		}
		agentResp.Body.Close()
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)

	email := testutil.RandomEmail()
	regPayload := testutil.RegisterRequest{Email: email, Password: "password123", Role: "seeker"}
	body, _ := json.Marshal(regPayload)

	resp1, _ := testutil.HTTPClient.Post(baseURL+"/api/auth/register", "application/json", bytes.NewReader(body))
	resp1.Body.Close()
	if resp1.StatusCode != 201 {
		t.Fatalf("first register should succeed, got %d", resp1.StatusCode)
	}

	// Register same email again — should return 409
	resp2, _ := testutil.HTTPClient.Post(baseURL+"/api/auth/register", "application/json", bytes.NewReader(body))
	defer resp2.Body.Close()

	if resp2.StatusCode != 409 {
		t.Errorf("expected 409 for duplicate email, got %d", resp2.StatusCode)
	}
}

func TestLogin_Real(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)

	// Register first
	email := testutil.RandomEmail()
	regPayload := testutil.RegisterRequest{Email: email, Password: "securepass99", Role: "recruiter"}
	body, _ := json.Marshal(regPayload)
	resp, _ := testutil.HTTPClient.Post(baseURL+"/api/auth/register", "application/json", bytes.NewReader(body))
	resp.Body.Close()
	if resp.StatusCode != 201 {
		t.Fatalf("register failed, got %d", resp.StatusCode)
	}

	// Login with correct password
	loginPayload := map[string]string{"email": email, "password": "securepass99"}
	body, _ = json.Marshal(loginPayload)
	resp2, _ := testutil.HTTPClient.Post(baseURL+"/api/auth/login", "application/json", bytes.NewReader(body))
	defer resp2.Body.Close()

	if resp2.StatusCode != 200 {
		t.Errorf("expected 200 for correct login, got %d", resp2.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp2.Body).Decode(&result)
	token, ok := result["token"].(string)
	if !ok || token == "" {
		t.Error("expected non-empty token from login")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)

	email := testutil.RandomEmail()
	regPayload := testutil.RegisterRequest{Email: email, Password: "correctpass", Role: "seeker"}
	body, _ := json.Marshal(regPayload)
	resp, _ := testutil.HTTPClient.Post(baseURL+"/api/auth/register", "application/json", bytes.NewReader(body))
	resp.Body.Close()

	wrongPayload := map[string]string{"email": email, "password": "wrongpassword"}
	body, _ = json.Marshal(wrongPayload)
	resp2, _ := testutil.HTTPClient.Post(baseURL+"/api/auth/login", "application/json", bytes.NewReader(body))
	defer resp2.Body.Close()

	if resp2.StatusCode != 401 {
		t.Errorf("expected 401 for wrong password, got %d", resp2.StatusCode)
	}
}

func TestProtectedEndpoint_NoToken(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)

	req, _ := http.NewRequest("GET", baseURL+"/api/agents", nil)
	resp, _ := testutil.HTTPClient.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode != 401 {
		t.Errorf("expected 401 for no token, got %d", resp.StatusCode)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)

	payload := map[string]string{"email": "not-an-email", "password": "password123", "role": "seeker"}
	body, _ := json.Marshal(payload)
	resp, _ := testutil.HTTPClient.Post(baseURL+"/api/auth/register", "application/json", bytes.NewReader(body))
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Errorf("expected 400 for invalid email, got %d", resp.StatusCode)
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)

	payload := map[string]string{"email": testutil.RandomEmail(), "password": "short", "role": "seeker"}
	body, _ := json.Marshal(payload)
	resp, _ := testutil.HTTPClient.Post(baseURL+"/api/auth/register", "application/json", bytes.NewReader(body))
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Errorf("expected 400 for short password, got %d", resp.StatusCode)
	}
}

func TestRegister_MissingRole(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)

	payload := map[string]string{"email": testutil.RandomEmail(), "password": "password123"}
	body, _ := json.Marshal(payload)
	resp, _ := testutil.HTTPClient.Post(baseURL+"/api/auth/register", "application/json", bytes.NewReader(body))
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Errorf("expected 400 for missing role, got %d", resp.StatusCode)
	}
}

func TestRegister_InvalidRole(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)

	payload := map[string]string{"email": testutil.RandomEmail(), "password": "password123", "role": "admin"}
	body, _ := json.Marshal(payload)
	resp, _ := testutil.HTTPClient.Post(baseURL+"/api/auth/register", "application/json", bytes.NewReader(body))
	defer resp.Body.Close()

	// Invalid role should return 400 or 422
	if resp.StatusCode != 400 && resp.StatusCode != 422 {
		t.Errorf("expected 400/422 for invalid role, got %d", resp.StatusCode)
	}
}

func TestLogin_NonExistentEmail(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)

	payload := map[string]string{"email": "nonexistent+" + testutil.RandomEmail(), "password": "anypass"}
	body, _ := json.Marshal(payload)
	resp, _ := testutil.HTTPClient.Post(baseURL+"/api/auth/login", "application/json", bytes.NewReader(body))
	defer resp.Body.Close()

	if resp.StatusCode != 401 {
		t.Errorf("expected 401 for non-existent email, got %d", resp.StatusCode)
	}
}

// TestRegister_VerifyEmailNotLeaked verifies that when registering,
// the response does not echo back the password or include it in any field.
func TestRegister_VerifyEmailNotLeaked(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)

	password := "supersecret123"
	payload := testutil.RegisterRequest{Email: testutil.RandomEmail(), Password: password, Role: "seeker"}
	body, _ := json.Marshal(payload)

	resp, _ := testutil.HTTPClient.Post(baseURL+"/api/auth/register", "application/json", bytes.NewReader(body))
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	respBody, _ := json.Marshal(result)
	if strings.Contains(string(respBody), password) {
		t.Fatal("FAKE: password appears in register response body")
	}
}
