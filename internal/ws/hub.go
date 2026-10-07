package ws

import (
	"sync"

	"github.com/google/uuid"
)

// Hub manages WebSocket rooms and connections.
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[string]*Connection
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]map[string]*Connection),
	}
}

// Join adds a connection to a room.
func (h *Hub) Join(roomID string, conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[string]*Connection)
	}
	h.rooms[roomID][conn.ID] = conn

	conn.RoomID = roomID
}

// Leave removes a connection from its room.
func (h *Hub) Leave(connID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for roomID, conns := range h.rooms {
		if conn, ok := conns[connID]; ok {
			delete(conns, connID)
			conn.RoomID = ""
			if len(conns) == 0 {
				delete(h.rooms, roomID)
			}
			return
		}
	}
}

// Broadcast sends a message to all connections in a room, optionally excluding one.
func (h *Hub) Broadcast(roomID string, message []byte, excludeConnID string) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns, ok := h.rooms[roomID]
	if !ok {
		return
	}

	for id, conn := range conns {
		if id == excludeConnID {
			continue
		}
		conn.Send(message)
	}
}

// RoomUsers returns the user IDs of all members in a room.
func (h *Hub) RoomUsers(roomID string) []uuid.UUID {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns, ok := h.rooms[roomID]
	if !ok {
		return nil
	}

	seen := make(map[uuid.UUID]struct{})
	for _, conn := range conns {
		if conn.UserID == "" {
			continue
		}
		id, err := uuid.Parse(conn.UserID)
		if err != nil {
			continue
		}
		seen[id] = struct{}{}
	}

	result := make([]uuid.UUID, 0, len(seen))
	for id := range seen {
		result = append(result, id)
	}

	return result
}

// Room returns all connections in a room.
func (h *Hub) Room(roomID string) map[string]*Connection {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make(map[string]*Connection)
	conns, ok := h.rooms[roomID]
	if !ok {
		return result
	}

	for k, v := range conns {
		result[k] = v
	}

	return result
}

// RoomCount returns the number of connections in a room.
func (h *Hub) RoomCount(roomID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms[roomID])
}
