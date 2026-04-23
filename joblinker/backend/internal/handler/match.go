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
	id := uuid.MustParse(c.GetString("id"))
	match, err := h.svc.ConfirmMatch(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm match"})
		return
	}
	c.JSON(http.StatusOK, match)
}

func (h *MatchHandler) Get(c *gin.Context) {
	id := uuid.MustParse(c.GetString("id"))
	match, err := h.svc.GetMatch(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
		return
	}
	c.JSON(http.StatusOK, match)
}
