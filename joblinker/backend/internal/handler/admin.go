package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct{}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

func (h *AdminHandler) GetDashboard(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"total_agents":       0,
		"active_jobs":        0,
		"pending_matches":    0,
		"completed_offers":   0,
		"total_interviews":   0,
		"recent_activity":    []interface{}{},
	})
}
