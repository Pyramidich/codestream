package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/ilya/codestream/internal/models"
	"github.com/ilya/codestream/internal/service"
	"github.com/ilya/codestream/internal/ws"
)

// FileServiceForWS defines the file service interface needed by WSHandler.
type FileServiceForWS interface {
	GetByID(ctx context.Context, fileID, userID uuid.UUID) (*models.File, error)
}

// DocumentStateManagerForWS defines the document state manager interface needed by WSHandler.
type DocumentStateManagerForWS interface {
	ApplyUpdate(ctx context.Context, fileID, userID uuid.UUID, update []byte) error
	GetSnapshot(ctx context.Context, fileID uuid.UUID) ([]byte, error)
	SaveSnapshot(ctx context.Context, fileID uuid.UUID) error
	LoadSnapshot(ctx context.Context, fileID uuid.UUID) error
}

// WSHandler handles WebSocket connections.
type WSHandler struct {
	hub                 *ws.Hub
	fileService         FileServiceForWS
	documentStateManager DocumentStateManagerForWS
	jwtSecret           string
	logger              *slog.Logger
}

// NewWSHandler creates a new WSHandler.
func NewWSHandler(hub *ws.Hub, fileService FileServiceForWS, documentStateManager DocumentStateManagerForWS, jwtSecret string, logger *slog.Logger) *WSHandler {
	return &WSHandler{
		hub:                  hub,
		fileService:          fileService,
		documentStateManager: documentStateManager,
		jwtSecret:            jwtSecret,
		logger:               logger,
	}
}

// Handle upgrades the HTTP connection to WebSocket.
func (h *WSHandler) Handle(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
		return
	}

	userID, err := h.parseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	fileIDStr := c.Query("file_id")
	if fileIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_id required"})
		return
	}

	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file_id"})
		return
	}

	file, err := h.fileService.GetByID(c.Request.Context(), fileID, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if file == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	conn, err := ws.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Warn("websocket upgrade failed", slog.String("error", err.Error()))
		return
	}

	roomID := "file:" + fileID.String()
	connection := ws.NewConnection(uuid.New().String(), userID.String(), roomID, conn, h.logger)

	h.hub.Join(roomID, connection)

	if err := h.documentStateManager.LoadSnapshot(c.Request.Context(), fileID); err != nil {
		h.logger.Warn("failed to load snapshot", slog.String("error", err.Error()))
	}

	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel

	snapshot, err := h.documentStateManager.GetSnapshot(c.Request.Context(), fileID)
	if err != nil {
		h.logger.Warn("failed to get snapshot", slog.String("error", err.Error()))
	}

	connection.SendJSON(ws.EventDocSync, map[string]string{
		"fileId":      fileID.String(),
		"state":       base64.StdEncoding.EncodeToString(snapshot),
		"stateVector": "", // stateVector is accepted but ignored for MVP
	})

	go connection.WritePump(ctx, func() {
		h.handleDisconnect(connection)
	})

	connection.ReadPump(ctx, func(message []byte) {
		h.handleMessage(connection, fileID, userID, message)
	})
}

func (h *WSHandler) handleDisconnect(connection *ws.Connection) {
	h.hub.Leave(connection.ID)

	if connection.RoomID != "" {
		h.hub.Broadcast(connection.RoomID, mustJSON(ws.WSMessage{
			Event: ws.EventUserLeft,
			Data:  mustJSON(map[string]string{"userId": connection.UserID}),
		}), connection.ID)

		// If room is empty, save snapshot
		if h.hub.RoomCount(connection.RoomID) == 0 {
			fileID, err := uuid.Parse(connection.RoomID[5:])
			if err == nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := h.documentStateManager.SaveSnapshot(ctx, fileID); err != nil {
					h.logger.Warn("failed to save snapshot on disconnect", slog.String("error", err.Error()))
				}
			}
		}
	}
}

func (h *WSHandler) handleMessage(connection *ws.Connection, fileID, userID uuid.UUID, message []byte) {
	var msg ws.WSMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		h.logger.Warn("invalid websocket message", slog.String("error", err.Error()))
		return
	}

	switch msg.Event {
	case ws.EventJoinFile:
		connection.SendJSON(ws.EventJoinedFile, map[string]string{"fileId": fileID.String()})

	case ws.EventLeaveFile:
		connection.Close()

	case ws.EventDocUpdate:
		var data service.DocUpdateData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			h.logger.Warn("invalid doc:update data", slog.String("error", err.Error()))
			return
		}

		update, err := base64.StdEncoding.DecodeString(data.Update)
		if err != nil {
			h.logger.Warn("invalid doc:update base64", slog.String("error", err.Error()))
			return
		}

		if err := h.documentStateManager.ApplyUpdate(context.Background(), fileID, userID, update); err != nil {
			h.logger.Warn("failed to apply update", slog.String("error", err.Error()))
		}

	case ws.EventPong:
		// Heartbeat handled by connection pong handler

	default:
		h.logger.Debug("unhandled websocket event", slog.String("event", msg.Event))
	}
}

func (h *WSHandler) parseToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenUnverifiable
		}
		return []byte(h.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return uuid.Nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, err
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, err
	}

	return uuid.Parse(sub)
}

func mustJSON(v interface{}) []byte {
	if v == nil {
		return []byte("{}")
	}
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return b
}
