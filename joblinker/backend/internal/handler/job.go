package handler

import (
	"encoding/json"
	"net/http"

	"joblinker/internal/model"
	"joblinker/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type JobHandler struct {
	jobRepo   *repository.JobRepository
	agentRepo *repository.AgentRepository
}

func NewJobHandler(jobRepo *repository.JobRepository, agentRepo *repository.AgentRepository) *JobHandler {
	return &JobHandler{jobRepo: jobRepo, agentRepo: agentRepo}
}

type CreateJobRequest struct {
	StructuredJSON string `json:"structured" binding:"required"`
	VectorID       string `json:"vector_id"`
}

func (h *JobHandler) Create(c *gin.Context) {
	userID := uuid.MustParse(c.GetString("userID"))
	var req CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	agents, _ := h.agentRepo.ListByUserID(userID, "")
	var agentID uuid.UUID
	for _, a := range agents {
		if a.Type == model.AgentTypeRecruiter {
			agentID = a.ID
			break
		}
	}
	if agentID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No recruiter agent found"})
		return
	}
	job := &model.Job{
		AgentID:        agentID,
		StructuredJSON: json.RawMessage(req.StructuredJSON),
		VectorID:       req.VectorID,
		Status:         model.JobStatusActive,
	}
	if err := h.jobRepo.Create(job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job"})
		return
	}
	c.JSON(http.StatusCreated, job)
}

func (h *JobHandler) Get(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	job, err := h.jobRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}
	c.JSON(http.StatusOK, job)
}

func (h *JobHandler) List(c *gin.Context) {
	jobs, _, err := h.jobRepo.ListAll(100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list jobs"})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

func (h *JobHandler) Update(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	job, err := h.jobRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if status, ok := updates["status"].(string); ok {
		job.Status = model.JobStatus(status)
	}
	if err := h.jobRepo.Update(job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update job"})
		return
	}
	c.JSON(http.StatusOK, job)
}

func (h *JobHandler) Delete(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	if err := h.jobRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete job"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Job deleted"})
}
