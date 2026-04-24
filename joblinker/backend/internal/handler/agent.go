package handler

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"joblinker/internal/model"
	"joblinker/internal/service"
	"joblinker/pkg/parser"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AgentHandler struct {
	svc *service.AgentService
}

func NewAgentHandler(svc *service.AgentService) *AgentHandler {
	return &AgentHandler{svc: svc}
}

type CreateAgentRequest struct {
	Type       string `json:"type" binding:"required,oneof=seeker recruiter"`
	ConfigJSON string `json:"config"`
	ResumeURL  string `json:"resume_url"` // URL to resume file for parsing
}

func (h *AgentHandler) Create(c *gin.Context) {
	userID := uuid.MustParse(c.GetString("userID"))
	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If resume URL provided, parse it to extract skills
	if req.ResumeURL != "" {
		resumeParser := parser.NewResumeParser()
		parsedResume, err := resumeParser.ParseFromURL(req.ResumeURL)
		if err == nil && parsedResume != nil {
			// Add parsed resume data to config
			config := map[string]interface{}{
				"resume_parsed": true,
				"skills":        parsedResume.Skills,
				"experience":    parsedResume.Experience,
				"education":    parsedResume.Education,
			}
			if req.ConfigJSON != "" {
				var existing map[string]interface{}
				if err := json.Unmarshal([]byte(req.ConfigJSON), &existing); err == nil {
					for k, v := range existing {
						config[k] = v
					}
				}
			}
			configJSON, _ := json.Marshal(config)
			req.ConfigJSON = string(configJSON)
		}
	}

	agent, err := h.svc.CreateAgent(userID, model.AgentType(req.Type), req.ConfigJSON)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create agent: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, agent)
}

func (h *AgentHandler) Get(c *gin.Context) {
	id := uuid.MustParse(c.GetString("id"))
	agent, err := h.svc.GetAgent(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}
	c.JSON(http.StatusOK, agent)
}

func (h *AgentHandler) List(c *gin.Context) {
	userID := uuid.MustParse(c.GetString("userID"))
	agents, err := h.svc.ListUserAgents(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list agents"})
		return
	}
	c.JSON(http.StatusOK, agents)
}

func (h *AgentHandler) Update(c *gin.Context) {
	id := uuid.MustParse(c.GetString("id"))
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	agent, err := h.svc.UpdateAgent(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update agent"})
		return
	}
	c.JSON(http.StatusOK, agent)
}

func (h *AgentHandler) Delete(c *gin.Context) {
	id := uuid.MustParse(c.GetString("id"))
	if err := h.svc.DeleteAgent(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete agent"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

type ParseResumeRequest struct {
	URL      string `json:"url"`
	FilePath string `json:"file_path"`
}

func (h *AgentHandler) ParseResume(c *gin.Context) {
	var req ParseResumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resumeParser := parser.NewResumeParser()
	var parsed *parser.ParsedResume
	var err error

	if req.URL != "" {
		parsed, err = resumeParser.ParseFromURL(req.URL)
	} else if req.FilePath != "" {
		parsed, err = resumeParser.ParseFile(req.FilePath)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either url or file_path is required"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse resume: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, parsed)
}

func getFileExtension(filePath string) string {
	ext := filepath.Ext(filePath)
	return ext
}
