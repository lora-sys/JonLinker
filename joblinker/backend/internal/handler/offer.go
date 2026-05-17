package handler

import (
	"net/http"
	"time"

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

func (h *OfferHandler) List(c *gin.Context) {
	offers, total, err := h.svc.ListAll(100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list offers"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"offers": offers, "total": total})
}

func (h *OfferHandler) Create(c *gin.Context) {
	var req struct {
		MatchID         string `json:"match_id" binding:"required"`
		CompensationJSON string `json:"compensation" binding:"required"`
		StartDate       string `json:"start_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	matchID := uuid.MustParse(req.MatchID)
	startDate := req.StartDate
	if startDate == "" {
		startDate = time.Now().AddDate(0, 1, 0).Format("2006-01-02")
	}
	offer, err := h.svc.GenerateOffer(matchID, req.CompensationJSON, startDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create offer"})
		return
	}
	c.JSON(http.StatusCreated, offer)
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
		Response      string  `json:"response" binding:"required,oneof=accept decline negotiate"`
		CounterAmount float64 `json:"counter_amount,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	offer, err := h.svc.RespondToOffer(id, req.Response, req.CounterAmount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to respond to offer"})
		return
	}
	c.JSON(http.StatusOK, offer)
}

func (h *OfferHandler) GetByMatchID(c *gin.Context) {
	matchID := uuid.MustParse(c.Param("matchId"))
	offer, err := h.svc.GetOfferByMatchID(matchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Offer not found"})
		return
	}
	c.JSON(http.StatusOK, offer)
}

func (h *OfferHandler) Accept(c *gin.Context) {
	matchID := uuid.MustParse(c.Param("matchId"))
	offer, err := h.svc.AcceptOffer(matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept offer"})
		return
	}
	c.JSON(http.StatusOK, offer)
}

func (h *OfferHandler) Decline(c *gin.Context) {
	matchID := uuid.MustParse(c.Param("matchId"))
	offer, err := h.svc.DeclineOffer(matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decline offer"})
		return
	}
	c.JSON(http.StatusOK, offer)
}
