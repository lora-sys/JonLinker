package handler

import (
	"net/http"

	"joblinker/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PrivacyHandler struct {
	svc *service.PrivacyService
}

func NewPrivacyHandler(svc *service.PrivacyService) *PrivacyHandler {
	return &PrivacyHandler{svc: svc}
}

func (h *PrivacyHandler) Export(c *gin.Context) {
	userID := uuid.MustParse(c.GetString("userID"))
	data, err := h.svc.ExportUserData(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export data"})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *PrivacyHandler) DeleteAccount(c *gin.Context) {
	userID := uuid.MustParse(c.GetString("userID"))
	if err := h.svc.DeleteUserAccount(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Account deletion scheduled"})
}
