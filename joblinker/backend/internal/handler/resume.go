package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"joblinker/pkg/ai"

	"github.com/gin-gonic/gin"
)

type ResumeHandler struct{}

func NewResumeHandler() *ResumeHandler {
	return &ResumeHandler{}
}

type GenerateResumeRequest struct {
	UserInfo string `json:"user_info" binding:"required"`
}

type GenerateResumeResponse struct {
	Resume any    `json:"resume"`
	Raw    string `json:"raw"`
}

// Generate creates an AI-generated resume from user info
func (h *ResumeHandler) Generate(c *gin.Context) {
	var req GenerateResumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	aiClient := ai.NewClient()
	if aiClient == nil || aiClient.APIKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI service not configured"})
		return
	}

	// Build prompt for resume generation
	systemPrompt := `You are a professional resume writer. Generate a structured resume in JSON format from the user's information.

Return ONLY valid JSON with this exact structure - no markdown, no explanation:
{
  "summary": "2-3 sentence professional summary",
  "skills": ["skill1", "skill2", ...],
  "experience": [
    {"title": "Job Title", "company": "Company Name", "duration": "2020-2024", "bullets": ["Achievement 1", "Achievement 2"]}
  ],
  "education": [
    {"degree": "Degree Name", "institution": "School Name", "year": "2020"}
  ],
  "achievements": ["Achievement 1", "Achievement 2"]
}`

	response, err := aiClient.Chat(systemPrompt, req.UserInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("AI generation failed: %v", err)})
		return
	}

	// Parse the response as JSON
	var resumeData map[string]interface{}
	if err := json.Unmarshal([]byte(response), &resumeData); err != nil {
		// If parsing fails, return raw response
		c.JSON(http.StatusOK, gin.H{
			"resume": nil,
			"raw":    response,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"resume": resumeData,
		"raw":    "",
	})
}