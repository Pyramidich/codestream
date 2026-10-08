package ws

import "encoding/json"

// WSMessage represents a WebSocket message.
type WSMessage struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

// Event types.
const (
	EventJoinFile        = "join:file"
	EventLeaveFile       = "leave:file"
	EventJoinedFile      = "joined:file"
	EventUserJoined      = "user:joined"
	EventUserLeft        = "user:left"
	EventPing            = "ping"
	EventPong            = "pong"
	EventDocUpdate       = "doc:update"
	EventDocSync         = "doc:sync"
	EventAwarenessUpdate = "awareness:update"
	EventPresence        = "presence:list"
)
