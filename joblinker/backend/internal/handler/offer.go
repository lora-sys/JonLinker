package handler

import (
	"net/http"

	"joblinker/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OfferHandler struct {
	svc *service.OfferService
}

func NewOfferHandler(svc *service.OfferService) *OfferHandler {
	return &OfferHandler{svc: svc}
}

func (h *OfferHandler) Get(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	offer, err := h.svc.GetOffer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Offer not found"})
		return
	}
	c.JSON(http.StatusOK, offer)
}

func (h *OfferHandler) Respond(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	var req struct {
		Response string `json:"response" binding:"required,oneof=accept decline negotiate"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	offer, err := h.svc.RespondToOffer(id, req.Response)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to respond to offer"})
		return
	}
	c.JSON(http.StatusOK, offer)
}
