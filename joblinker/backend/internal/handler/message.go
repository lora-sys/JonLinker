package handler

import (
	"encoding/xml"
	"net/http"
	"sync"

	"joblinker/internal/model"
	"joblinker/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

type MessageHandler struct {
	messageRepo *repository.MessageRepository
	matchRepo   *repository.MatchRepository
	agentRepo   *repository.AgentRepository
	clients     map[string]*websocket.Conn
	mu          sync.RWMutex
}

func NewMessageHandler(messageRepo *repository.MessageRepository, matchRepo *repository.MatchRepository, agentRepo *repository.AgentRepository) *MessageHandler {
	return &MessageHandler{
		messageRepo: messageRepo,
		matchRepo:   matchRepo,
		agentRepo:   agentRepo,
		clients:     make(map[string]*websocket.Conn),
	}
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
	if matchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "match_id required"})
		return
	}

	userID := uuid.MustParse(c.GetString("userID"))

	// Verify user has access to this match
	match, err := h.matchRepo.GetByID(uuid.MustParse(matchID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
		return
	}

	// Verify agent ownership
	agent, err := h.agentRepo.GetByID(match.SeekerAgentID)
	if err != nil || agent.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection"})
		return
	}

	clientID := userID.String()
	h.mu.Lock()
	h.clients[clientID] = conn
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, clientID)
		h.mu.Unlock()
		conn.Close()
	}()

	// Send initial connection success message
	conn.WriteJSON(WSMessage{
		Type:    "connected",
		Payload: map[string]interface{}{"match_id": matchID},
	})

	// Handle incoming messages
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Parse incoming XML message
		var xmlMsg A2AMessage
		if err := xml.Unmarshal(data, &xmlMsg); err != nil {
			conn.WriteJSON(WSMessage{
				Type:    "error",
				Payload: map[string]interface{}{"message": "Invalid message format"},
			})
			continue
		}

		// Store message in database
		message := &model.Message{
			ID:            uuid.New(),
			MatchID:       uuid.MustParse(matchID),
			SenderAgentID: agent.ID,
			ContentXML:    string(data),
			IntentType:    xmlMsg.Payload.Intent,
		}
		h.messageRepo.Create(message)

		// Process message and generate response
		response := h.processMessage(matchID, userID.String(), &xmlMsg)

		// Broadcast response to both parties
		h.broadcastToMatch(matchID, response)
	}
}

func (h *MessageHandler) processMessage(matchID, senderID string, msg *A2AMessage) *A2AMessage {
	// Generate appropriate response based on intent
	response := &A2AMessage{
		Header: MessageHeader{
			MessageID:  uuid.New().String(),
			Timestamp:  "2026-04-23T00:00:00Z", // Would use time.Now().Format()
			SenderID:   "system",
			ReceiverID: senderID,
			ReplyTo:    msg.Header.MessageID,
		},
	}

	switch msg.Payload.Intent {
	case IntentIntroduction:
		response.Payload.Intent = IntentInterest
		response.Payload.Parameters = &MessageParams{
			Role: "system_response",
		}
	case IntentInterest:
		response.Payload.Intent = IntentNegotiation
		response.Payload.Negotiation = &NegotiationInfo{
			Type:     "salary",
			Current:  100000,
			Target:   120000,
			Currency: "USD",
		}
	case IntentNegotiation:
		// Use salary negotiator to generate response
		negotiator := NewSalaryNegotiator(80000, 150000, 120000)
		result := negotiator.SimulateNegotiation(100000)
		response.Payload.Negotiation = &NegotiationInfo{
			Type:        "salary",
			Current:     result.FinalOffer,
			Target:      result.AgreedSalary,
			Currency:    "USD",
			Concessions: result.CounterOffers,
		}
	case IntentOffer:
		response.Payload.Intent = IntentAccept
		response.Payload.Parameters = &MessageParams{
			Role: "accept",
		}
	default:
		response.Payload.Intent = IntentInquiry
		response.Payload.Parameters = &MessageParams{
			Role: "acknowledge",
		}
	}

	return response
}

func (h *MessageHandler) broadcastToMatch(matchID string, msg *A2AMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, _ := xml.MarshalIndent(msg, "", "  ")
	for _, conn := range h.clients {
		conn.WriteMessage(websocket.TextMessage, data)
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

	// Verify match access
	match, err := h.matchRepo.GetByID(uuid.MustParse(matchID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
		return
	}

	agent, err := h.agentRepo.GetByID(match.SeekerAgentID)
	if err != nil || agent.UserID != userID {
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
		SenderAgentID: agent.ID,
		ContentXML:    req.ContentXML,
		IntentType:    req.IntentType,
	}

	if err := h.messageRepo.Create(message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store message"})
		return
	}

	c.JSON(http.StatusCreated, message)
}

// GetConversation retrieves message history for a match
func (h *MessageHandler) GetConversation(c *gin.Context) {
	matchID := c.Param("matchId")
	userID := uuid.MustParse(c.GetString("userID"))

	// Verify access
	match, err := h.matchRepo.GetByID(uuid.MustParse(matchID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
		return
	}

	agent, err := h.agentRepo.GetByID(match.SeekerAgentID)
	if err != nil || agent.UserID != userID {
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
type SalaryNegotiator struct {
	minSalary    int
	maxSalary    int
	targetSalary int
	currentOffer int
}

func NewSalaryNegotiator(min, max, target int) *SalaryNegotiator {
	return &SalaryNegotiator{
		minSalary:    min,
		maxSalary:    max,
		targetSalary: target,
		currentOffer: max,
	}
}

type NegotiationResult struct {
	AgreedSalary  int
	FinalOffer    int
	CounterOffers int
	Conceded      bool
	Reached       bool
}

// GenerateCounterOffer creates a counter offer based on strategy and negotiation state
func (n *SalaryNegotiator) GenerateCounterOffer(currentOffer int, rounds int) int {
	if currentOffer < n.minSalary {
		return -1 // Walk away
	}

	gap := n.maxSalary - currentOffer
	concessionRate := 0.15 + (float64(rounds) * 0.05)
	if concessionRate > 0.35 {
		concessionRate = 0.35
	}

	counterOffer := currentOffer + int(float64(gap)*concessionRate)
	if counterOffer < n.targetSalary {
		counterOffer = n.targetSalary
	}

	return counterOffer
}

// EvaluateOffer checks if an offer meets minimum requirements
func (n *SalaryNegotiator) EvaluateOffer(offer int) int {
	if offer < n.minSalary {
		return 0 // Reject
	}
	if offer >= n.targetSalary {
		return 2 // Accept
	}
	return 1 // Counter
}

// SimulateNegotiation runs a complete salary negotiation
func (n *SalaryNegotiator) SimulateNegotiation(initialOffer int) *NegotiationResult {
	result := &NegotiationResult{
		FinalOffer: initialOffer,
	}

	rounds := 0
	currentOffer := initialOffer

	for rounds < 5 {
		decision := n.EvaluateOffer(currentOffer)
		switch decision {
		case 2: // Accept
			result.Reached = true
			result.AgreedSalary = currentOffer
			return result
		case 0: // Reject
			result.Reached = false
			return result
		case 1: // Counter
			rounds++
			result.CounterOffers++
			currentOffer = n.GenerateCounterOffer(currentOffer, rounds)
			if currentOffer < 0 {
				result.Reached = false
				return result
			}
			result.FinalOffer = currentOffer
		}
	}

	// Max rounds reached
	if currentOffer >= n.minSalary {
		result.Reached = true
		result.Conceded = true
		result.AgreedSalary = currentOffer
	} else {
		result.Reached = false
	}
	return result
}