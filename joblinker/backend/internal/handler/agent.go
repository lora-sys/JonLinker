package handler

import (
	"net/http"

	"joblinker/internal/model"
	"joblinker/internal/service"

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
	Type      string `json:"type" binding:"required,oneof=seeker recruiter"`
	ConfigJSON string `json:"config"`
}

func (h *AgentHandler) Create(c *gin.Context) {
	userID := uuid.MustParse(c.GetString("userID"))
	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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
