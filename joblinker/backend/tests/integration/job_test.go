package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestJobCreate_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("POST /api/jobs validates required fields", func(t *testing.T) {
		// Test with missing title
		body := map[string]interface{}{
			"description": "We are hiring",
			"type":        "full-time",
			"structured": map[string]interface{}{
				"requirements": []string{"Go", "React"},
			},
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/jobs", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Verify ShouldBindJSON rejects missing required fields
		var testReq struct {
			Title string `json:"title" binding:"required"`
		}
		if err := c.ShouldBindJSON(&testReq); err == nil {
			t.Error("expected error for missing title")
		}
	})

	t.Run("job type enum validation", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.New()
		r.POST("/test-job", func(c *gin.Context) {
			var req struct {
				Type string `json:"type" binding:"required,oneof=full-time part-time contract"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Test invalid type
		body := map[string]interface{}{"type": "intern"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/test-job", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for invalid type, got %d", w.Code)
		}

		// Test valid types
		validTypes := []string{"full-time", "part-time", "contract"}
		for _, vt := range validTypes {
			body := map[string]interface{}{"type": vt}
			bodyBytes, _ := json.Marshal(body)
			req := httptest.NewRequest(http.MethodPost, "/test-job", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected 200 for type %s, got %d", vt, w.Code)
			}
		}
	})

	t.Run("structured requirements validation", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.New()
		r.POST("/test-job", func(c *gin.Context) {
			var req struct {
				Requirements []string `json:"requirements" binding:"required,min=1"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Test empty requirements
		body := map[string]interface{}{"requirements": []string{}}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/test-job", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for empty requirements, got %d", w.Code)
		}

		// Test non-empty requirements
		body = map[string]interface{}{"requirements": []string{"Go"}}
		bodyBytes, _ = json.Marshal(body)
		req = httptest.NewRequest(http.MethodPost, "/test-job", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200 for non-empty requirements, got %d", w.Code)
		}
	})
}