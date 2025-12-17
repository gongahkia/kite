package websocket

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Rooms and their clients
	rooms map[string]map[*Client]bool

	// Inbound messages from clients
	Broadcast chan Message

	// Register requests from clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Room subscription requests
	subscribe   chan *subscription
	unsubscribe chan *subscription

	mu sync.RWMutex
}

type subscription struct {
	client *Client
	room   string
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		rooms:       make(map[string]map[*Client]bool),
		Broadcast:   make(chan Message, 256),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		subscribe:   make(chan *subscription),
		unsubscribe: make(chan *subscription),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run(ctx context.Context) {
	log.Info().Msg("WebSocket hub started")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("WebSocket hub shutting down")
			return

		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Info().Str("client_id", client.ID).Msg("Client registered")

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)

				// Remove client from all rooms
				for room, clients := range h.rooms {
					if clients[client] {
						delete(clients, client)
						if len(clients) == 0 {
							delete(h.rooms, room)
						}
					}
				}
			}
			h.mu.Unlock()
			log.Info().Str("client_id", client.ID).Msg("Client unregistered")

		case msg := <-h.Broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- msg:
				default:
					// Client's send channel is full, close and remove
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()

		case sub := <-h.subscribe:
			h.mu.Lock()
			if h.rooms[sub.room] == nil {
				h.rooms[sub.room] = make(map[*Client]bool)
			}
			h.rooms[sub.room][sub.client] = true
			h.mu.Unlock()
			log.Info().
				Str("client_id", sub.client.ID).
				Str("room", sub.room).
				Msg("Client subscribed to room")

		case sub := <-h.unsubscribe:
			h.mu.Lock()
			if clients, ok := h.rooms[sub.room]; ok {
				if clients[sub.client] {
					delete(clients, sub.client)
					if len(clients) == 0 {
						delete(h.rooms, sub.room)
					}
				}
			}
			h.mu.Unlock()
			log.Info().
				Str("client_id", sub.client.ID).
				Str("room", sub.room).
				Msg("Client unsubscribed from room")
		}
	}
}

// BroadcastToRoom sends a message to all clients in a specific room
func (h *Hub) BroadcastToRoom(room string, msg Message) {
	h.mu.RLock()
	clients, exists := h.rooms[room]
	h.mu.RUnlock()

	if !exists {
		return
	}

	h.mu.RLock()
	for client := range clients {
		select {
		case client.Send <- msg:
		default:
			close(client.Send)
			delete(h.clients, client)
		}
	}
	h.mu.RUnlock()
}

// Subscribe adds a client to a room
func (h *Hub) Subscribe(client *Client, room string) {
	h.subscribe <- &subscription{
		client: client,
		room:   room,
	}
}

// Unsubscribe removes a client from a room
func (h *Hub) Unsubscribe(client *Client, room string) {
	h.unsubscribe <- &subscription{
		client: client,
		room:   room,
	}
}

// GetRoomSize returns the number of clients in a room
func (h *Hub) GetRoomSize(room string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.rooms[room]; ok {
		return len(clients)
	}
	return 0
}

// GetClientCount returns the total number of connected clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetRooms returns a list of all active rooms
func (h *Hub) GetRooms() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	rooms := make([]string, 0, len(h.rooms))
	for room := range h.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}
