package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// Server manages WebSocket connections and message broadcasting
type Server struct {
	hub     *Hub
	clients map[string]*Client
	mu      sync.RWMutex
}

// NewServer creates a new WebSocket server
func NewServer() *Server {
	return &Server{
		hub:     NewHub(),
		clients: make(map[string]*Client),
	}
}

// Start starts the WebSocket server
func Start(ctx context.Context, s *Server) error {
	go s.hub.Run(ctx)
	log.Info().Msg("WebSocket server started")
	return nil
}

// Handler creates a Fiber WebSocket handler
func (s *Server) Handler() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		// Create new client
		client := NewClient(s.hub, c)

		// Register client
		s.mu.Lock()
		s.clients[client.ID] = client
		s.mu.Unlock()

		// Send welcome message
		welcome := Message{
			Type: MessageTypeWelcome,
			Data: map[string]interface{}{
				"client_id": client.ID,
				"message":   "Connected to Kite WebSocket API",
			},
			Timestamp: time.Now(),
		}
		if err := client.Send(welcome); err != nil {
			log.Error().Err(err).Str("client_id", client.ID).Msg("Failed to send welcome message")
		}

		// Register with hub
		s.hub.Register <- client

		// Start client read/write loops
		go client.WritePump()
		client.ReadPump()

		// Cleanup on disconnect
		s.mu.Lock()
		delete(s.clients, client.ID)
		s.mu.Unlock()
	})
}

// Broadcast sends a message to all connected clients
func (s *Server) Broadcast(msg Message) {
	s.hub.Broadcast <- msg
}

// BroadcastToRoom sends a message to all clients in a specific room
func (s *Server) BroadcastToRoom(room string, msg Message) {
	s.hub.BroadcastToRoom(room, msg)
}

// SendToClient sends a message to a specific client
func (s *Server) SendToClient(clientID string, msg Message) error {
	s.mu.RLock()
	client, exists := s.clients[clientID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("client not found: %s", clientID)
	}

	return client.Send(msg)
}

// GetClientCount returns the number of connected clients
func (s *Server) GetClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

// GetRoomClients returns the number of clients in a room
func (s *Server) GetRoomClients(room string) int {
	return s.hub.GetRoomSize(room)
}

// Message represents a WebSocket message
type Message struct {
	Type      MessageType            `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// MessageType represents the type of WebSocket message
type MessageType string

const (
	// Connection messages
	MessageTypeWelcome    MessageType = "welcome"
	MessageTypePing       MessageType = "ping"
	MessageTypePong       MessageType = "pong"
	MessageTypeSubscribe  MessageType = "subscribe"
	MessageTypeUnsubscribe MessageType = "unsubscribe"

	// Scraping events
	MessageTypeScrapeStarted  MessageType = "scrape.started"
	MessageTypeScrapeProgress MessageType = "scrape.progress"
	MessageTypeScrapeComplete MessageType = "scrape.complete"
	MessageTypeScrapeError    MessageType = "scrape.error"

	// Search events
	MessageTypeSearchResults MessageType = "search.results"
	MessageTypeSearchUpdate  MessageType = "search.update"

	// Case events
	MessageTypeCaseCreated MessageType = "case.created"
	MessageTypeCaseUpdated MessageType = "case.updated"
	MessageTypeCaseDeleted MessageType = "case.deleted"

	// Validation events
	MessageTypeValidationComplete MessageType = "validation.complete"
	MessageTypeQualityAlert       MessageType = "quality.alert"

	// Worker events
	MessageTypeWorkerStatus MessageType = "worker.status"
	MessageTypeQueueUpdate  MessageType = "queue.update"

	// System events
	MessageTypeSystemAlert MessageType = "system.alert"
	MessageTypeMetricUpdate MessageType = "metric.update"
)

// ToJSON converts a message to JSON
func (m Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// NewMessage creates a new message with current timestamp
func NewMessage(msgType MessageType, data map[string]interface{}) Message {
	return Message{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// Middleware to upgrade HTTP to WebSocket
func UpgradeMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if it's a WebSocket upgrade request
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}
