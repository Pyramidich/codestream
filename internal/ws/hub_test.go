package ws_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ilya/codestream/internal/logger"
	"github.com/ilya/codestream/internal/ws"
)

func TestHubJoinLeave(t *testing.T) {
	hub := ws.NewHub()
	log := logger.New("dev", "info")

	conn := ws.NewConnection("conn1", uuid.New().String(), "room1", nil, log)
	hub.Join("room1", conn)

	assert.Equal(t, 1, hub.RoomCount("room1"))

	hub.Leave("conn1")
	assert.Equal(t, 0, hub.RoomCount("room1"))
}

func TestHubBroadcast(t *testing.T) {
	hub := ws.NewHub()
	log := logger.New("dev", "info")

	conn1 := ws.NewConnection("conn1", uuid.New().String(), "room1", nil, log)
	conn2 := ws.NewConnection("conn2", uuid.New().String(), "room1", nil, log)

	hub.Join("room1", conn1)
	hub.Join("room1", conn2)

	msg := []byte(`{"event":"test"}`)
	hub.Broadcast("room1", msg, "conn1")

	// conn2 should receive the message
	select {
	case received := <-conn2.SendChannel():
		assert.Equal(t, msg, received)
	default:
		t.Fatal("expected message to be sent to conn2")
	}

	// conn1 (sender) should not receive
	select {
	case <-conn1.SendChannel():
		t.Fatal("did not expect message to be sent to sender")
	default:
	}
}
