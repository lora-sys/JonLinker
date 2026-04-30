package handler

import (
	"net/http"

	"joblinker/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MatchHandler struct {
	svc *service.MatchService
}

func NewMatchHandler(svc *service.MatchService) *MatchHandler {
	return &MatchHandler{svc: svc}
}

func (h *MatchHandler) List(c *gin.Context) {
	userID := uuid.MustParse(c.GetString("userID"))
	matches, err := h.svc.ListUserMatches(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list matches"})
		return
	}
	c.JSON(http.StatusOK, matches)
}

func (h *MatchHandler) Confirm(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	match, err := h.svc.ConfirmMatch(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm match"})
		return
	}
	c.JSON(http.StatusOK, match)
}

func (h *MatchHandler) Get(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	match, err := h.svc.GetMatch(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
		return
	}
	c.JSON(http.StatusOK, match)
}

func (h *MatchHandler) AutoCreate(c *gin.Context) {
	var req struct {
		JobIDs []string `json:"job_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "job_ids required"})
		return
	}

	userID := uuid.MustParse(c.GetString("userID"))
	jobUUIDs := make([]uuid.UUID, len(req.JobIDs))
	for i, id := range req.JobIDs {
		jobUUIDs[i] = uuid.MustParse(id)
	}

	result, err := h.svc.AutoCreateMatches(userID, jobUUIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create matches"})
		return
	}
	c.JSON(http.StatusOK, result)
}
