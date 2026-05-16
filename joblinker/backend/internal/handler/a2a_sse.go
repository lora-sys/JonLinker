package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"joblinker/internal/eino/a2ui"
	"joblinker/internal/eino/runner"
)

type A2ASSEHandler struct {
	runner *runner.ADKRunner
}

func NewA2ASSEHandler(runner *runner.ADKRunner) *A2ASSEHandler {
	return &A2ASSEHandler{runner: runner}
}

type a2aChatRequest struct {
	Message string `json:"message"`
}

func (h *A2ASSEHandler) Chat(c *gin.Context) {
	var req a2aChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message required"})
		return
	}

	sessionID := c.DefaultQuery("session_id", uuid.New().String())

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.WriteHeader(http.StatusOK)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return
	}

	events := h.runner.Query(c.Request.Context(), req.Message)

	if err := a2ui.StreamToWriter(c.Request.Context(), c.Writer, events, sessionID); err != nil && err != io.EOF {
		log.Printf("A2A SSE stream error: %v", err)
		errData, _ := json.Marshal(map[string]string{"error": err.Error()})
		fmt.Fprintf(c.Writer, "data: %s\n\n", errData)
	}

	flusher.Flush()
}
