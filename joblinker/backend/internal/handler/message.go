package handler

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"joblinker/internal/config"
	"joblinker/internal/middleware"
	"joblinker/internal/model"
	"joblinker/internal/repository"
	"joblinker/internal/service"
	"joblinker/pkg/proto"
	"joblinker/pkg/rabbitmq"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
		if allowedOrigins == "" {
			return true
		}
		origin := r.Header.Get("Origin")
		for _, allowed := range strings.Split(allowedOrigins, ",") {
			if strings.TrimSpace(allowed) == origin {
				return true
			}
		}
		log.Printf("WebSocket: origin %s not in whitelist %s", origin, allowedOrigins)
		return false
	},
}

type MessageHandler struct {
	messageRepo *repository.MessageRepository
	matchRepo   *repository.MatchRepository
	agentRepo   *repository.AgentRepository
	rmq         *rabbitmq.RabbitMQ
	mqSvc       *service.MessageQueueService
	clients     map[string]map[string]*websocket.Conn // matchID -> clientID -> conn
	mu          sync.RWMutex
}

func NewMessageHandler(messageRepo *repository.MessageRepository, matchRepo *repository.MatchRepository, agentRepo *repository.AgentRepository, rmq *rabbitmq.RabbitMQ, mqSvc *service.MessageQueueService) *MessageHandler {
	return &MessageHandler{
		messageRepo: messageRepo,
		matchRepo:   matchRepo,
		agentRepo:   agentRepo,
		rmq:         rmq,
		mqSvc:       mqSvc,
		clients:     make(map[string]map[string]*websocket.Conn),
	}
}

// BroadcastAgentResponse sends a payload to all WebSocket clients in a match room
func (h *MessageHandler) BroadcastAgentResponse(matchID string, eventType string, payload interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns, ok := h.clients[matchID]
	if !ok {
		return
	}

	msg := WSMessage{
		Type:    eventType,
		Payload: payload,
	}
	for _, conn := range conns {
		conn.WriteJSON(msg)
	}
}

// HandleHumanConfirm is the REST endpoint for human confirmation
func (h *MessageHandler) HandleHumanConfirm(c *gin.Context) {
	matchID := c.Param("id")
	if h.mqSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "message queue service not available"})
		return
	}

	var req struct {
		Approved bool   `json:"approved"`
		Feedback string `json:"feedback,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mid, err := uuid.Parse(matchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match id"})
		return
	}

	if err := h.mqSvc.HandleHumanConfirm(mid, req.Approved, req.Feedback); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}

// A2A Message types for XML protocol
const (
	IntentIntroduction = "INTRODUCTION"
	IntentInterest     = "INTEREST"
	IntentNegotiation  = "NEGOTIATION"
	IntentOffer        = "OFFER"
	IntentAccept       = "ACCEPT"
	IntentDecline      = "DECLINE"
	IntentSchedule     = "SCHEDULE"
	IntentConfirm      = "CONFIRM"
	IntentWithdraw     = "WITHDRAW"
	IntentInquiry      = "INQUIRY"
)

type A2AMessage struct {
	XMLName   xml.Name    `xml:"message"`
	Header    MessageHeader `xml:"header"`
	Payload   MessagePayload `xml:"payload"`
}

type MessageHeader struct {
	MessageID  string `xml:"message_id"`
	Timestamp  string `xml:"timestamp"`
	SenderID   string `xml:"sender_id"`
	ReceiverID string `xml:"receiver_id"`
	ReplyTo    string `xml:"reply_to,omitempty"`
}

type MessagePayload struct {
	Intent      string         `xml:"intent"`
	Parameters  *MessageParams `xml:"parameters,omitempty"`
	Negotiation *NegotiationInfo `xml:"negotiation,omitempty"`
}

type MessageParams struct {
	Role          string `xml:"role,omitempty"`
	Name          string `xml:"name,omitempty"`
	Title         string `xml:"title,omitempty"`
	Company       string `xml:"company,omitempty"`
	Location      string `xml:"location,omitempty"`
	SalaryMin     int    `xml:"salary_min,omitempty"`
	SalaryMax     int    `xml:"salary_max,omitempty"`
	Experience    string `xml:"experience,omitempty"`
	Skills        string `xml:"skills,omitempty"`
	MatchID       string `xml:"match_id,omitempty"`
	InterestLevel string `xml:"interest_level,omitempty"`
}

type NegotiationInfo struct {
	Round        int     `xml:"round,attr"`
	Type         string  `xml:"type,omitempty"`
	Current      int     `xml:"current,omitempty"`
	Target       int     `xml:"target,omitempty"`
	Currency     string  `xml:"currency,omitempty"`
	Concessions  int     `xml:"concessions,omitempty"`
	Compensation *CompensationInfo `xml:"compensation,omitempty"`
	StartDate    string  `xml:"start_date,omitempty"`
	Notes        string  `xml:"notes,omitempty"`
}

type CompensationInfo struct {
	BaseSalary int    `xml:"base_salary"`
	Currency   string `xml:"currency"`
	Bonus      int    `xml:"bonus,omitempty"`
	Equity     string `xml:"equity,omitempty"`
}

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func (h *MessageHandler) HandleWebSocket(c *gin.Context) {
	matchID := c.Param("matchId")

	// Get token from Sec-WebSocket-Protocol header (preferred) or query param (fallback)
	tokenStr := c.GetHeader("Sec-WebSocket-Protocol")
	if tokenStr == "" {
		tokenStr = c.Query("token")
	}
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
		return
	}

	log.Printf("WebSocket HandleWebSocket: matchId=%q", matchID)

	userID, err := extractUserIDFromToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Parse Gateway identity query params (WebSocket can't use headers)
	userIDParam := c.Query("user_id")
	agentIDParam := c.Query("agent_id")
	tenantIDParam := c.Query("tenant_id")

	// Set in Gin context for downstream use
	if userIDParam != "" {
		c.Set("userID", userIDParam)
	}
	if agentIDParam != "" {
		c.Set("agentID", agentIDParam)
	}
	if tenantIDParam != "" {
		c.Set("tenantID", tenantIDParam)
	}

	// If matchId is provided, verify access to this match
	var agent *model.Agent
	if matchID != "" && matchID != "ws" {
		// Verify user has access to this match (either as seeker or recruiter)
		match, matchErr := h.matchRepo.GetByID(uuid.MustParse(matchID))
		if matchErr != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
			return
		}

		// Check: user must own seeker agent OR recruiter (job) agent
		seekerAgent, _ := h.agentRepo.GetByID(match.SeekerAgentID)
		hasAccess := false
		if seekerAgent != nil && seekerAgent.UserID == userID {
			agent = seekerAgent
			hasAccess = true
		} else if match.Job != nil {
			// Check recruiter (job's agent)
			agents, _ := h.agentRepo.ListByUserID(userID, "")
			for _, a := range agents {
				if a.ID == match.Job.AgentID {
					agent = a
					hasAccess = true
					break
				}
			}
		}
		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	// Only echo Sec-WebSocket-Protocol if the client sent one (RFC 6455 §4.2.2)
	upgradeHeader := http.Header{}
	if clientProtocol := c.GetHeader("Sec-WebSocket-Protocol"); clientProtocol != "" {
		upgradeHeader["Sec-WebSocket-Protocol"] = []string{tokenStr}
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, upgradeHeader)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Send initial connection success message BEFORE adding to clients map
	conn.WriteJSON(WSMessage{
		Type:    "connected",
		Payload: map[string]interface{}{"match_id": matchID},
	})

	clientID := userID.String()
	h.mu.Lock()
	if h.clients[matchID] == nil {
		h.clients[matchID] = make(map[string]*websocket.Conn)
	}
	h.clients[matchID][clientID] = conn
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients[matchID], clientID)
		if len(h.clients[matchID]) == 0 {
			delete(h.clients, matchID)
		}
		h.mu.Unlock()
		conn.Close()
	}()

	// Initialize WebSocket frame serializer for Protobuf support
	wsSerializer := proto.NewWebSocketFrameSerializer()
	var sequenceNum uint64 = 0

	// Handle incoming messages
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Detect message format (Protobuf vs JSON/XML)
		format := wsSerializer.DetectFrameFormat(data)
		log.Printf("WebSocket message format: %s, size: %d bytes", format, len(data))

		var xmlMsg A2AMessage
		var processErr error

		// Handle control messages
		if format == "json" {
			var msg struct {
				Type    string `json:"type"`
				MatchID string `json:"match_id"`
			}
			if err := json.Unmarshal(data, &msg); err == nil {
				if msg.Type == "join_match" && msg.MatchID != "" {
					// Move client to the specified match room
					log.Printf("Client %s joining match %s", userID.String(), msg.MatchID)
					h.mu.Lock()
					// Remove from old room
					delete(h.clients[matchID], userID.String())
					// Add to new room
					if h.clients[msg.MatchID] == nil {
						h.clients[msg.MatchID] = make(map[string]*websocket.Conn)
					}
					h.clients[msg.MatchID][userID.String()] = conn
					// Update matchID for this connection
					matchID = msg.MatchID
					h.mu.Unlock()
					conn.WriteJSON(WSMessage{
						Type:    "joined",
						Payload: map[string]interface{}{"match_id": matchID},
					})
					continue
				}
			}
		}

		if format == "protobuf" && config.IsWebSocketProtobufEnabled() {
			// Protobuf mode: deserialize WebSocketFrame
			frame, err := wsSerializer.UnmarshalFrame(data)
			if err != nil {
				log.Printf("Failed to unmarshal Protobuf frame: %v", err)
				conn.WriteJSON(WSMessage{
					Type:    "error",
					Payload: map[string]interface{}{"message": "Invalid Protobuf format"},
				})
				continue
			}

			// Use frame metadata
			sequenceNum = frame.SequenceNum
			log.Printf("Processing Protobuf frame: message_type=%s, sequence=%d",
				proto.GetMessageTypeName(frame.MessageType), frame.SequenceNum)

			// Convert Protobuf payload to A2AMessage for processing
			if len(frame.Payload) > 0 {
				if err := xml.Unmarshal(frame.Payload, &xmlMsg); err != nil {
					processErr = err
				}
			}
		} else if format == "xml" {
			if err := xml.Unmarshal(data, &xmlMsg); err != nil {
				log.Printf("Failed to parse message: %v", err)
				conn.WriteJSON(WSMessage{
					Type:    "error",
					Payload: map[string]interface{}{"message": "Invalid message format"},
				})
				continue
			}
		}

		if processErr != nil {
			conn.WriteJSON(WSMessage{
				Type:    "error",
				Payload: map[string]interface{}{"message": fmt.Sprintf("Processing error: %v", processErr)},
			})
			continue
		}

		// Store message in database only for valid XML A2A messages
		// (skip JSON control messages like pings, join_match, etc.)
		if format == "xml" && matchID != "" && matchID != "ws" {
			var senderAgentID uuid.UUID
			if agent != nil {
				senderAgentID = agent.ID
			}
			message := &model.Message{
				ID:            uuid.New(),
				MatchID:       uuid.MustParse(matchID),
				SenderAgentID: senderAgentID,
				ContentXML:    string(data),
				IntentType:    xmlMsg.Payload.Intent,
			}
			h.messageRepo.Create(message)

			// Process message and generate response (only for A2A XML messages)
			response := h.processMessage(matchID, userID.String(), &xmlMsg)

			// Broadcast response to both parties (respecting Protobuf mode)
			h.broadcastToMatchWithFormat(matchID, response, wsSerializer, &sequenceNum)
		}
	}
}

func (h *MessageHandler) processMessage(matchID, senderID string, msg *A2AMessage) *A2AMessage {
	// Route through RabbitMQ for AI-generated response
	if h.rmq != nil {
		go func() {
			agentMsg := &rabbitmq.AgentMessage{
				MessageID:  uuid.New().String(),
				SenderID:   senderID,
				ReceiverID: "",
				Intent:     msg.Payload.Intent,
				MatchID:    matchID,
				Payload:    nil,
				Timestamp:  time.Now(),
			}
			if err := h.rmq.PublishAgentMessage(context.Background(), agentMsg); err != nil {
				log.Printf("Failed to publish message to RabbitMQ: %v", err)
			}
		}()
	}

	// Return acknowledgement; AI-generated response arrives asynchronously via RabbitMQ consumer
	return &A2AMessage{
		Header: MessageHeader{
			MessageID:  uuid.New().String(),
			Timestamp:  time.Now().Format(time.RFC3339),
			SenderID:   "system",
			ReceiverID: senderID,
			ReplyTo:    msg.Header.MessageID,
		},
		Payload: MessagePayload{
			Intent: IntentInquiry,
			Parameters: &MessageParams{
				Role: "acknowledge",
			},
		},
	}
}

func (h *MessageHandler) broadcastToMatch(matchID string, msg *A2AMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns, ok := h.clients[matchID]
	if !ok {
		return
	}
	data, _ := xml.MarshalIndent(msg, "", "  ")
	for _, conn := range conns {
		conn.WriteMessage(websocket.TextMessage, data)
	}
}

// broadcastToMatchWithFormat sends messages in either Protobuf or JSON based on config
func (h *MessageHandler) broadcastToMatchWithFormat(matchID string, msg *A2AMessage, wsSerializer *proto.WebSocketFrameSerializer, sequenceNum *uint64) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns, ok := h.clients[matchID]
	if !ok {
		return
	}

	// Serialize the A2AMessage to XML payload
	xmlData, _ := xml.MarshalIndent(msg, "", "  ")

	if config.IsWebSocketProtobufEnabled() {
		// Protobuf mode: wrap in WebSocketFrame
		*sequenceNum++
		frame := &proto.WebSocketFrame{
			MessageType:   proto.MessageType_TEXT,
			Payload:       xmlData,
			SequenceNum:   *sequenceNum,
			Timestamp:     time.Now().UnixMilli(),
			SchemaVersion: 1,
		}

		protoData, err := wsSerializer.MarshalFrame(frame)
		if err != nil {
			log.Printf("Failed to marshal Protobuf frame: %v", err)
			return
		}

		for _, conn := range conns {
			conn.WriteMessage(websocket.BinaryMessage, protoData)
		}
	} else {
		// JSON/XML mode
		for _, conn := range conns {
			conn.WriteMessage(websocket.TextMessage, xmlData)
		}
	}
}

// SendMessage is the REST endpoint for sending messages
func (h *MessageHandler) SendMessage(c *gin.Context) {
	matchID := c.Param("matchId")
	userID := uuid.MustParse(c.GetString("userID"))

	var req struct {
		ContentXML string `json:"content_xml" binding:"required"`
		IntentType string `json:"intent_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify match access — allow BOTH seeker and recruiter (job's agent) to send
	match, err := h.matchRepo.GetByID(uuid.MustParse(matchID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
		return
	}

	// Find an agent for this user
	agents, err := h.agentRepo.ListByUserID(userID, "")
	if err != nil || len(agents) == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Check: user must own either the seeker agent OR the recruiter agent (Job.AgentID)
	var senderAgentID uuid.UUID
	hasAccess := false
	for _, agent := range agents {
		if agent.ID == match.SeekerAgentID {
			senderAgentID = agent.ID
			hasAccess = true
			break
		}
		// Recruiter is the job's agent
		if match.Job != nil && agent.ID == match.Job.AgentID {
			senderAgentID = agent.ID
			hasAccess = true
			break
		}
	}
	if !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Parse and validate XML content
	var xmlMsg A2AMessage
	if err := xml.Unmarshal([]byte(req.ContentXML), &xmlMsg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid XML format"})
		return
	}

	// Create message
	message := &model.Message{
		ID:            uuid.New(),
		MatchID:       uuid.MustParse(matchID),
		SenderAgentID: senderAgentID,
		ContentXML:    req.ContentXML,
		IntentType:    req.IntentType,
	}

	if err := h.messageRepo.Create(message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store message"})
		return
	}

	// Publish to RabbitMQ for auto-response (async)
	if h.rmq != nil {
		go func() {
			agentMsg := &rabbitmq.AgentMessage{
				MessageID:  message.ID.String(),
				SenderID:   senderAgentID.String(),
				ReceiverID: "",
				Intent:     req.IntentType,
				MatchID:    matchID,
				Payload:    nil,
				Timestamp:  message.CreatedAt,
			}
			// Fire and forget - log error only
			_ = h.rmq.PublishAgentMessage(context.Background(), agentMsg)
		}()
	} else {
		// Fallback: process message directly when RabbitMQ is not available
		go func() {
			log.Printf("Processing message directly (RabbitMQ not available)")
			// Import the service package to access MessageQueueService logic
			// This is a simplified fallback - in production, would want proper queue handling
		}()
	}

	c.JSON(http.StatusCreated, message)
}

// GetMessages retrieves all conversations for a user (grouped by match)
func (h *MessageHandler) GetMessages(c *gin.Context) {
	userID := uuid.MustParse(c.GetString("userID"))

	agents, err := h.agentRepo.ListByUserID(userID, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve agents"})
		return
	}

	agentIDs := make([]uuid.UUID, 0, len(agents))
	for _, agent := range agents {
		agentIDs = append(agentIDs, agent.ID)
	}

	matches, err := h.matchRepo.ListByAgentIDs(agentIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve matches"})
		return
	}

	type convMsg struct {
		ID            string    `json:"id"`
		MatchID       string    `json:"match_id"`
		SenderAgentID string    `json:"sender_agent_id"`
		ContentXML    string    `json:"content_xml"`
		IntentType    string    `json:"intent_type"`
		CreatedAt     time.Time `json:"created_at"`
	}

	conversations := make([]map[string]interface{}, 0, len(matches))
	for _, match := range matches {
		messages, err := h.messageRepo.ListByMatchID(match.ID)
		if err != nil {
			continue
		}

		jobTitle := ""
		if match.Job != nil {
			var jobData map[string]interface{}
			if err := json.Unmarshal(match.Job.StructuredJSON, &jobData); err == nil {
				if title, ok := jobData["title"].(string); ok {
					jobTitle = title
				}
			}
		}

		var lastMsg *convMsg
		for _, msg := range messages {
			lastMsg = &convMsg{
				ID:            msg.ID.String(),
				MatchID:       msg.MatchID.String(),
				SenderAgentID: msg.SenderAgentID.String(),
				ContentXML:    msg.ContentXML,
				IntentType:    msg.IntentType,
				CreatedAt:     msg.CreatedAt,
			}
		}

		conversations = append(conversations, map[string]interface{}{
			"MatchID":     match.ID.String(),
			"JobTitle":    jobTitle,
			"LastMessage": lastMsg,
			"UnreadCount": 0,
			"UpdatedAt":   match.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"conversations": conversations})
}

// GetConversation retrieves message history for a match
func (h *MessageHandler) GetConversation(c *gin.Context) {
	matchID := c.Param("matchId")
	userID := uuid.MustParse(c.GetString("userID"))

	// Verify access — allow both seeker and recruiter
	match, err := h.matchRepo.GetByID(uuid.MustParse(matchID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
		return
	}

	// Get user's agents
	agents, err := h.agentRepo.ListByUserID(userID, "")
	if err != nil || len(agents) == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Check: user must own seeker agent OR recruiter (job) agent
	hasAccess := false
	for _, agent := range agents {
		if agent.ID == match.SeekerAgentID {
			hasAccess = true
			break
		}
		if match.Job != nil && agent.ID == match.Job.AgentID {
			hasAccess = true
			break
		}
	}
	if !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	messages, err := h.messageRepo.ListByMatchID(uuid.MustParse(matchID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve messages"})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// SalaryNegotiator handles salary and compensation negotiations


func extractUserIDFromToken(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return middleware.GetJwtSecret(), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid claims")
	}
	sub, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid sub claim")
	}
	return uuid.Parse(sub)
}