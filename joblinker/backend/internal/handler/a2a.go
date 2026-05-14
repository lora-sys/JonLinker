package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"joblinker/internal/repository"
	"joblinker/pkg/rabbitmq"
)

var a2aUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type A2AHandler struct {
	matchRepo   *repository.MatchRepository
	agentRepo   *repository.AgentRepository
	messageRepo *repository.MessageRepository
	rmq         *rabbitmq.RabbitMQ
	clients     map[string]map[string]*websocket.Conn
	mu          sync.RWMutex
}

func NewA2AHandler(
	matchRepo *repository.MatchRepository,
	agentRepo *repository.AgentRepository,
	messageRepo *repository.MessageRepository,
	rmq *rabbitmq.RabbitMQ,
) *A2AHandler {
	return &A2AHandler{
		matchRepo:   matchRepo,
		agentRepo:   agentRepo,
		messageRepo: messageRepo,
		rmq:         rmq,
		clients:     make(map[string]map[string]*websocket.Conn),
	}
}

// A2A auth uses HMAC-SHA256 signature with a shared internal key
func (h *A2AHandler) verifyA2AToken(tokenStr string) (string, bool) {
	secret := os.Getenv("A2A_INTERNAL_SECRET")
	if secret == "" {
		secret = "a2a-internal-dev-key"
	}
	parts := strings.SplitN(tokenStr, ":", 2)
	if len(parts) != 2 {
		return "", false
	}
	agentID := parts[0]
	sig := parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(agentID))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return "", false
	}
	return agentID, true
}

func (h *A2AHandler) HandleA2AWebSocket(c *gin.Context) {
	matchID := c.Param("matchId")

	tokenStr := c.GetHeader("Sec-WebSocket-Protocol")
	if tokenStr == "" {
		tokenStr = c.Query("token")
	}
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "A2A token required"})
		return
	}

	agentID, ok := h.verifyA2AToken(tokenStr)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid A2A token"})
		return
	}

	log.Printf("A2A WebSocket authenticated: agentID=%s, matchId=%s", agentID, matchID)

	conn, err := a2aUpgrader.Upgrade(c.Writer, c.Request, http.Header{
		"Sec-WebSocket-Protocol": {tokenStr},
	})
	if err != nil {
		log.Printf("A2A WebSocket upgrade failed: %v", err)
		return
	}

	clientID := uuid.New().String()
	h.mu.Lock()
	if h.clients[matchID] == nil {
		h.clients[matchID] = make(map[string]*websocket.Conn)
	}
	h.clients[matchID][clientID] = conn
	h.mu.Unlock()

	log.Printf("A2A client connected: clientID=%s, matchID=%s", clientID, matchID)
	conn.WriteJSON(map[string]string{"type": "connected", "client_id": clientID})

	defer func() {
		h.mu.Lock()
		delete(h.clients[matchID], clientID)
		if len(h.clients[matchID]) == 0 {
			delete(h.clients, matchID)
		}
		h.mu.Unlock()
		conn.Close()
		log.Printf("A2A client disconnected: clientID=%s", clientID)
	}()

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			continue
		}

		msg["sender_agent_id"] = agentID
		msg["match_id"] = matchID

		h.broadcastA2AMessage(matchID, msg)

		if h.rmq != nil {
			agentMsg := &rabbitmq.AgentMessage{
				MessageID:  uuid.New().String(),
				SenderID:   agentID,
				MatchID:    matchID,
				Intent:     msg["intent"].(string),
				Payload:    msg,
				Timestamp:  time.Now(),
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			h.rmq.PublishAgentMessage(ctx, agentMsg)
			cancel()
		}
	}
}

func (h *A2AHandler) broadcastA2AMessage(matchID string, msg map[string]interface{}) {
	h.mu.RLock()
	clients := h.clients[matchID]
	h.mu.RUnlock()

	for clientID, conn := range clients {
		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("A2A broadcast error to client %s: %v", clientID, err)
		}
	}
}
