package websocket

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512 KB
)

// Client represents a WebSocket client connection
type Client struct {
	// Unique client identifier
	ID string

	// The hub
	hub *Hub

	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	Send chan Message

	// Subscribed rooms
	rooms map[string]bool
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		ID:    uuid.New().String(),
		hub:   hub,
		conn:  conn,
		Send:  make(chan Message, 256),
		rooms: make(map[string]bool),
	}
}

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg map[string]interface{}
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Err(err).Str("client_id", c.ID).Msg("WebSocket read error")
			}
			break
		}

		// Handle incoming messages
		c.handleMessage(msg)
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(msg); err != nil {
				log.Error().Err(err).Str("client_id", c.ID).Msg("WebSocket write error")
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Send sends a message to the client
func (c *Client) Send(msg Message) error {
	select {
	case c.Send <- msg:
		return nil
	default:
		return ErrClientSendBufferFull
	}
}

// handleMessage processes incoming messages from the client
func (c *Client) handleMessage(msg map[string]interface{}) {
	msgType, ok := msg["type"].(string)
	if !ok {
		log.Warn().Str("client_id", c.ID).Msg("Message missing type field")
		return
	}

	switch MessageType(msgType) {
	case MessageTypePing:
		c.handlePing()

	case MessageTypeSubscribe:
		c.handleSubscribe(msg)

	case MessageTypeUnsubscribe:
		c.handleUnsubscribe(msg)

	default:
		log.Debug().
			Str("client_id", c.ID).
			Str("type", msgType).
			Msg("Unhandled message type")
	}
}

// handlePing responds to ping messages
func (c *Client) handlePing() {
	pong := NewMessage(MessageTypePong, map[string]interface{}{
		"client_id": c.ID,
	})
	c.Send(pong)
}

// handleSubscribe subscribes the client to a room
func (c *Client) handleSubscribe(msg map[string]interface{}) {
	room, ok := msg["room"].(string)
	if !ok {
		log.Warn().Str("client_id", c.ID).Msg("Subscribe message missing room field")
		return
	}

	c.hub.Subscribe(c, room)
	c.rooms[room] = true

	// Send confirmation
	confirmation := NewMessage(MessageTypeSubscribe, map[string]interface{}{
		"client_id": c.ID,
		"room":      room,
		"status":    "subscribed",
	})
	c.Send(confirmation)
}

// handleUnsubscribe unsubscribes the client from a room
func (c *Client) handleUnsubscribe(msg map[string]interface{}) {
	room, ok := msg["room"].(string)
	if !ok {
		log.Warn().Str("client_id", c.ID).Msg("Unsubscribe message missing room field")
		return
	}

	c.hub.Unsubscribe(c, room)
	delete(c.rooms, room)

	// Send confirmation
	confirmation := NewMessage(MessageTypeUnsubscribe, map[string]interface{}{
		"client_id": c.ID,
		"room":      room,
		"status":    "unsubscribed",
	})
	c.Send(confirmation)
}

// MarshalJSON implements json.Marshaler
func (c *Client) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"id":    c.ID,
		"rooms": c.getRoomList(),
	})
}

func (c *Client) getRoomList() []string {
	rooms := make([]string, 0, len(c.rooms))
	for room := range c.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// Errors
var (
	ErrClientSendBufferFull = fmt.Errorf("client send buffer is full")
)
