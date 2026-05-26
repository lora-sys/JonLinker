package rest

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"joblinker/internal/adapters"
	"joblinker/internal/model"
	"joblinker/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AgentHandler handles agent CRUD endpoints.
type AgentHandler struct {
	svc    *service.AgentService
	parser *adapters.ResumeParser
}

// NewAgentHandler creates a new AgentHandler.
// It uses the new adapters layer, replacing direct pkg/parser dependency.
func NewAgentHandler(svc *service.AgentService, parser *adapters.ResumeParser) *AgentHandler {
	return &AgentHandler{svc: svc, parser: parser}
}

// CreateAgentRequest is the JSON payload for creating an agent.
type CreateAgentRequest struct {
	Type       string `json:"type" binding:"required,oneof=seeker recruiter"`
	ConfigJSON string `json:"config"`
	ResumeURL  string `json:"resume_url"`
}

// Create handles POST /api/agents.
func (h *AgentHandler) Create(c *gin.Context) {
	userID := uuid.MustParse(c.GetString("userID"))
	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ResumeURL != "" {
		parsedResume, err := h.parser.ParseFromURL(req.ResumeURL)
		if err == nil && parsedResume != nil {
			config := map[string]interface{}{
				"resume_parsed": true,
				"skills":        parsedResume.Skills,
				"experience":    parsedResume.Experience,
				"education":     parsedResume.Education,
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

// Get handles GET /api/agents/:id.
func (h *AgentHandler) Get(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	agent, err := h.svc.GetAgent(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}
	c.JSON(http.StatusOK, agent)
}

// List handles GET /api/agents.
func (h *AgentHandler) List(c *gin.Context) {
	userID := uuid.MustParse(c.GetString("userID"))
	agents, err := h.svc.ListUserAgents(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list agents"})
		return
	}
	c.JSON(http.StatusOK, agents)
}

// Update handles PATCH /api/agents/:id.
func (h *AgentHandler) Update(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
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

// Delete handles DELETE /api/agents/:id.
func (h *AgentHandler) Delete(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	if err := h.svc.DeleteAgent(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete agent"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ParseResumeRequest is the JSON payload for parsing a resume.
type ParseResumeRequest struct {
	URL      string `json:"url"`
	FilePath string `json:"file_path"`
}

// ParseResume handles POST /api/resume/parse.
func (h *AgentHandler) ParseResume(c *gin.Context) {
	var req ParseResumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var parsed *adapters.ParsedResume
	var err error

	if req.URL != "" {
		parsed, err = h.parser.ParseFromURL(req.URL)
	} else if req.FilePath != "" {
		parsed, err = h.parser.ParseFile(req.FilePath)
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
