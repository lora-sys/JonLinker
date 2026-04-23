package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestHealthEndpoint(t *testing.T) {
	r := setupTestRouter()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %s", resp["status"])
	}
}

func TestAuthRegisterValidation(t *testing.T) {
	r := setupTestRouter()
	r.POST("/api/auth/register", func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=8"`
			Role     string `json:"role" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "user created"})
	})

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{
			name:       "valid request",
			body:       map[string]string{"email": "test@example.com", "password": "password123", "role": "seeker"},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid email",
			body:       map[string]string{"email": "notanemail", "password": "password123", "role": "seeker"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "short password",
			body:       map[string]string{"email": "test@example.com", "password": "short", "role": "seeker"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing role",
			body:       map[string]string{"email": "test@example.com", "password": "password123"},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestAuthLoginValidation(t *testing.T) {
	r := setupTestRouter()
	r.POST("/api/auth/login", func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": "fake-token"})
	})

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{
			name:       "valid request",
			body:       map[string]string{"email": "test@example.com", "password": "password123"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing password",
			body:       map[string]string{"email": "test@example.com"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid email",
			body:       map[string]string{"email": "notanemail", "password": "password123"},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestInterviewFormatValidation(t *testing.T) {
	r := setupTestRouter()
	r.POST("/api/interviews", func(c *gin.Context) {
		var req struct {
			MatchID     string `json:"match_id" binding:"required"`
			ScheduledAt string `json:"scheduled_at" binding:"required"`
			Format      string `json:"format" binding:"required,oneof=video phone onsite"`
			Location    string `json:"location"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "interview created"})
	})

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{
			name:       "valid video interview",
			body:       map[string]string{"match_id": "123", "scheduled_at": "2025-01-01T10:00:00Z", "format": "video"},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "valid phone interview",
			body:       map[string]string{"match_id": "123", "scheduled_at": "2025-01-01T10:00:00Z", "format": "phone"},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "valid onsite interview",
			body:       map[string]string{"match_id": "123", "scheduled_at": "2025-01-01T10:00:00Z", "format": "onsite", "location": "123 Office St"},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid format",
			body:       map[string]string{"match_id": "123", "scheduled_at": "2025-01-01T10:00:00Z", "format": "invalid"},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/api/interviews", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestOfferResponseValidation(t *testing.T) {
	r := setupTestRouter()
	r.POST("/api/offers/:id/respond", func(c *gin.Context) {
		var req struct {
			Response string `json:"response" binding:"required,oneof=accept decline negotiate"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "response recorded"})
	})

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{
			name:       "accept",
			body:       map[string]string{"response": "accept"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "decline",
			body:       map[string]string{"response": "decline"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "negotiate",
			body:       map[string]string{"response": "negotiate"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid response",
			body:       map[string]string{"response": "maybe"},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/api/offers/123/respond", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}
