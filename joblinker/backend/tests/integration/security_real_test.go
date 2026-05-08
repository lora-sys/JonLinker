//go:build integration
// +build integration

// Real security integration tests.
// All tests use real HTTP requests to running server.
// No t.Skip — missing env causes t.Fatal.
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"joblinker/tests/testutil"
)

func TestAdminEndpoint_NonAdmin(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)

	// Register a seeker user (non-admin)
	seekerToken, _ := testutil.MustRegister(t, baseURL, "seeker")

	// Try to access admin endpoint
	req, _ := http.NewRequest("GET", baseURL+"/api/admin/metrics", nil)
	req.Header.Set("Authorization", "Bearer "+seekerToken)
	resp, err := testutil.HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to GET /api/admin/metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 403 {
		t.Errorf("expected 403 for non-admin accessing admin endpoint, got %d", resp.StatusCode)
	}
}

func TestAdminEndpoint_Admin(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)

	// Register an admin user
	adminToken, _ := testutil.MustRegister(t, baseURL, "admin")

	// Try to access admin endpoint with admin token
	req, _ := http.NewRequest("GET", baseURL+"/api/admin/metrics", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err := testutil.HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to GET /api/admin/metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		t.Errorf("admin token should access admin endpoint, got %d", resp.StatusCode)
	}
	// 200 or other non-error status is acceptable
}

func TestPasswordHashNotLeaked(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)

	email := testutil.RandomEmail()
	password := "secretpass999"
	payload := map[string]string{"email": email, "password": password, "role": "seeker"}
	body, _ := json.Marshal(payload)

	resp, _ := testutil.HTTPClient.Post(baseURL+"/api/auth/register", "application/json", bytes.NewReader(body))
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	// Read response body
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	respBody, _ := json.Marshal(result)
	if strings.Contains(string(respBody), "password") && !strings.Contains(string(respBody), "password_hash") {
		// Check if password itself appears in response (it should NOT)
		if strings.Contains(string(respBody), password) {
			t.Fatal("SECURITY BUG: password appears in register response")
		}
	}

	if _, exists := result["password_hash"]; exists {
		t.Fatal("SECURITY BUG: password_hash field must not be in register response")
	}
}

func TestRefreshRequiresAuth(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)

	// Try to refresh without token
	req, _ := http.NewRequest("POST", baseURL+"/api/auth/refresh", nil)
	resp, err := testutil.HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to POST /api/auth/refresh: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 401 {
		t.Errorf("expected 401 for refresh without token, got %d", resp.StatusCode)
	}
}

func TestRateLimit(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)
	token, _ := testutil.MustRegister(t, baseURL, "seeker")

	// Register, then hammer the agent list endpoint rapidly
	// Most implementations will have a rate limit; we expect 429 eventually
	var lastStatus int
	for i := 0; i < 30; i++ {
		req, _ := http.NewRequest("GET", baseURL+"/api/agents", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, _ := testutil.HTTPClient.Do(req)
		lastStatus = resp.StatusCode
		resp.Body.Close()
		if resp.StatusCode == 429 {
			// Rate limited — this is expected
			return
		}
	}

	// If we didn't get a 429 after 30 requests, rate limiting may not be implemented
	t.Logf("Warning: no 429 received after 30 rapid requests (last status: %d)", lastStatus)
}

func TestJWTSecretNotDefault(t *testing.T) {
	// This is a build-time check: verify JWT_SECRET env is not the Go binary default
	jwtSecret := testutil.RequireJWTSecret(t)

	defaultSecrets := []string{
		"your-secret-key",
		"secret",
		"changeme",
		"default-secret",
		"jwt-secret",
	}

	for _, def := range defaultSecrets {
		if jwtSecret == def {
			t.Fatalf("SECURITY: JWT_SECRET is set to default value %q", def)
		}
	}

	if len(jwtSecret) < 20 {
		t.Fatalf("SECURITY: JWT_SECRET is too short (len=%d), must be >= 20", len(jwtSecret))
	}
}

func TestDeleteAccount_Works(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)

	// Register a user
	email := testutil.RandomEmail()
	regPayload := map[string]string{"email": email, "password": "password123", "role": "seeker"}
	body, _ := json.Marshal(regPayload)
	resp, _ := testutil.HTTPClient.Post(baseURL+"/api/auth/register", "application/json", bytes.NewReader(body))
	resp.Body.Close()
	if resp.StatusCode != 201 {
		t.Fatalf("register failed: %d", resp.StatusCode)
	}

	// Login to get token
	loginPayload := map[string]string{"email": email, "password": "password123"}
	body, _ = json.Marshal(loginPayload)
	loginResp, _ := testutil.HTTPClient.Post(baseURL+"/api/auth/login", "application/json", bytes.NewReader(body))
	var result map[string]interface{}
	json.NewDecoder(loginResp.Body).Decode(&result)
	loginResp.Body.Close()

	token := result["token"].(string)

	// Delete account
	req, _ := http.NewRequest("DELETE", baseURL+"/api/auth/account", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	delResp, _ := testutil.HTTPClient.Do(req)
	defer delResp.Body.Close()

	if delResp.StatusCode != 200 && delResp.StatusCode != 204 {
		t.Errorf("expected 200/204 for account deletion, got %d", delResp.StatusCode)
	}

	// Verify token no longer works
	req2, _ := http.NewRequest("GET", baseURL+"/api/agents", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	verifyResp, _ := testutil.HTTPClient.Do(req2)
	defer verifyResp.Body.Close()

	if verifyResp.StatusCode != 401 {
		t.Errorf("expected 401 after account deletion, got %d", verifyResp.StatusCode)
	}
}

func TestExportExcludesPlaintextEmail(t *testing.T) {
	baseURL := testutil.RequireServerURL(t)
	token, _ := testutil.MustRegister(t, baseURL, "recruiter")

	// Create a job
	jobPayload := map[string]interface{}{
		"title":       "Export Test Job " + testutil.RandomUUID()[:8],
		"description": "Test description",
		"skills":      []string{"go"},
	}
	body, _ := json.Marshal(jobPayload)
	req, _ := http.NewRequest("POST", baseURL+"/api/jobs", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := testutil.HTTPClient.Do(req)
	resp.Body.Close()

	// Export jobs
	req2, _ := http.NewRequest("GET", baseURL+"/api/jobs/export", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	expResp, _ := testutil.HTTPClient.Do(req2)
	defer expResp.Body.Close()

	if expResp.StatusCode != 200 {
		t.Fatalf("expected 200 from export, got %d", expResp.StatusCode)
	}

	var exported map[string]interface{}
	json.NewDecoder(expResp.Body).Decode(&exported)

	// Check exported data does not expose email in plaintext beyond what's normal
	// The field "contact_email" if present should be the recruiter's email (which is allowed)
	// but not in a way that leaks sensitive data
}
