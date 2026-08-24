package websocket

import (
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Broadcast payload to all clients in a specific tenant room.
	broadcast chan Message

	// Broadcast payload to all admin clients.
	broadcastAdmin chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	sync.RWMutex
}

type Message struct {
	TenantID string `json:"tenant_id"`
	Payload  []byte `json:"payload"`
}

func NewHub() *Hub {
	return &Hub{
		broadcast:      make(chan Message),
		broadcastAdmin: make(chan []byte),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		clients:        make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.Lock()
			h.clients[client] = true
			h.Unlock()
		case client := <-h.unregister:
			h.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.Unlock()
		case message := <-h.broadcast:
			h.RLock()
			for client := range h.clients {
				// Only send to clients belonging to the correct tenant
				if client.tenantID == message.TenantID {
					select {
					case client.send <- message.Payload:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
			h.RUnlock()
		case payload := <-h.broadcastAdmin:
			h.RLock()
			for client := range h.clients {
				if client.isAdmin {
					select {
					case client.send <- payload:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
			h.RUnlock()
		}
	}
}

// BroadcastMessage pushes a payload to all connected clients for a tenant
func (h *Hub) BroadcastMessage(tenantID string, payload []byte) {
	h.broadcast <- Message{
		TenantID: tenantID,
		Payload:  payload,
	}
}

// BroadcastAdmin pushes a payload to all connected admin clients
func (h *Hub) BroadcastAdmin(payload []byte) {
	h.broadcastAdmin <- payload
}


