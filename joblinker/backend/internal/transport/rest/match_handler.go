package rest

import (
	"net/http"

	"joblinker/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MatchHandler handles match CRUD and auto-match endpoints.
type MatchHandler struct {
	svc *service.MatchService
}

// NewMatchHandler creates a new MatchHandler.
func NewMatchHandler(svc *service.MatchService) *MatchHandler {
	return &MatchHandler{svc: svc}
}

// List handles GET /api/matches.
func (h *MatchHandler) List(c *gin.Context) {
	userID := uuid.MustParse(c.GetString("userID"))
	matches, err := h.svc.ListUserMatches(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list matches"})
		return
	}
	c.JSON(http.StatusOK, matches)
}

// Get handles GET /api/matches/:id.
func (h *MatchHandler) Get(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	match, err := h.svc.GetMatch(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
		return
	}
	c.JSON(http.StatusOK, match)
}

// AutoCreate handles POST /api/matches/auto.
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

// Confirm handles POST /api/matches/:id/confirm.
func (h *MatchHandler) Confirm(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	match, err := h.svc.ConfirmMatch(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm match"})
		return
	}
	c.JSON(http.StatusOK, match)
}

// Decline handles POST /api/matches/:id/decline.
func (h *MatchHandler) Decline(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	match, err := h.svc.DeclineMatch(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decline match"})
		return
	}
	c.JSON(http.StatusOK, match)
}
