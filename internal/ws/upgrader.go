package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// Upgrader is a permissive WebSocket upgrader for development.
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:    func(r *http.Request) bool { return true },
}
