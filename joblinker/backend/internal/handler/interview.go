package handler

import (
	"net/http"

	"joblinker/internal/model"
	"joblinker/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InterviewHandler struct {
	svc *service.InterviewService
}

func NewInterviewHandler(svc *service.InterviewService) *InterviewHandler {
	return &InterviewHandler{svc: svc}
}

func (h *InterviewHandler) List(c *gin.Context) {
	interviews, total, err := h.svc.ListAll(100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list interviews"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"interviews": interviews, "total": total})
}

func (h *InterviewHandler) Create(c *gin.Context) {
	var req struct {
		MatchID     string `json:"match_id" binding:"required"`
		ScheduledAt string `json:"scheduled_at" binding:"required"`
		Format      string `json:"format" binding:"required"`
		Location    string `json:"location"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	matchID := uuid.MustParse(req.MatchID)
	interview, err := h.svc.ScheduleInterview(matchID, req.ScheduledAt, model.InterviewFormat(req.Format), req.Location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to schedule interview"})
		return
	}
	c.JSON(http.StatusCreated, interview)
}

func (h *InterviewHandler) Update(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	interview, err := h.svc.UpdateInterview(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update interview"})
		return
	}
	c.JSON(http.StatusOK, interview)
}

func (h *InterviewHandler) GetByMatchID(c *gin.Context) {
	matchID := uuid.MustParse(c.Param("matchId"))
	interview, err := h.svc.GetInterviewByMatchID(matchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Interview not found"})
		return
	}
	c.JSON(http.StatusOK, interview)
}

func (h *InterviewHandler) Confirm(c *gin.Context) {
	matchID := uuid.MustParse(c.Param("matchId"))
	interview, err := h.svc.ConfirmInterview(matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm interview"})
		return
	}
	c.JSON(http.StatusOK, interview)
}

func (h *InterviewHandler) Cancel(c *gin.Context) {
	matchID := uuid.MustParse(c.Param("matchId"))
	interview, err := h.svc.CancelInterview(matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel interview"})
		return
	}
	c.JSON(http.StatusOK, interview)
}
