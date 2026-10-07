package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeTimeout = 10 * time.Second
	pingInterval = 30 * time.Second
	pongWait     = 60 * time.Second
)

// Connection represents a single WebSocket connection.
type Connection struct {
	ID     string
	UserID string
	RoomID string

	conn   *websocket.Conn
	sendCh chan []byte
	logger *slog.Logger

	mu       sync.RWMutex
	isClosed bool
}

// NewConnection creates a new WebSocket connection wrapper.
func NewConnection(id, userID, roomID string, conn *websocket.Conn, logger *slog.Logger) *Connection {
	return &Connection{
		ID:     id,
		UserID: userID,
		RoomID: roomID,
		conn:   conn,
		sendCh: make(chan []byte, 256),
		logger: logger,
	}
}

// Send queues a message for sending to the client.
func (c *Connection) Send(message []byte) {
	c.mu.RLock()
	closed := c.isClosed
	c.mu.RUnlock()

	if closed {
		return
	}

	select {
	case c.sendCh <- message:
	default:
		c.logger.Warn("send channel full, closing connection", slog.String("connID", c.ID))
		c.Close()
	}
}

// Close closes the connection and marks it as closed.
func (c *Connection) Close() {
	c.mu.Lock()
	if c.isClosed {
		c.mu.Unlock()
		return
	}
	c.isClosed = true
	c.mu.Unlock()

	close(c.sendCh)
	c.conn.Close()
}

// IsClosed returns true if the connection is closed.
func (c *Connection) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isClosed
}

// SendChannel returns the send channel of the connection (for testing).
func (c *Connection) SendChannel() <-chan []byte {
	return c.sendCh
}

// ReadPump handles reading messages from the client.
func (c *Connection) ReadPump(ctx context.Context, onMessage func(message []byte)) {
	defer c.Close()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Warn("websocket read error", slog.String("error", err.Error()))
			}
			return
		}

		c.conn.SetReadDeadline(time.Now().Add(pongWait))

		if onMessage != nil {
			onMessage(message)
		}
	}
}

// WritePump handles writing messages and ping frames to the client.
func (c *Connection) WritePump(ctx context.Context, onClose func()) {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.Close()
		if onClose != nil {
			onClose()
		}
	}()

	for {
		select {
		case message, ok := <-c.sendCh:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

// write writes a message with the given type and payload.
func (c *Connection) write(mt int, payload []byte) error {
	c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	return c.conn.WriteMessage(mt, payload)
}

// SendJSON sends a JSON message to the client.
func (c *Connection) SendJSON(event string, data interface{}) error {
	message, err := json.Marshal(WSMessage{Event: event, Data: mustJSON(data)})
	if err != nil {
		return err
	}

	c.Send(message)
	return nil
}

func mustJSON(v interface{}) json.RawMessage {
	if v == nil {
		return json.RawMessage("{}")
	}
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage("{}")
	}
	return b
}
