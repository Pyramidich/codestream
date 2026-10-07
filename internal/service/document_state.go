package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/ilya/codestream/internal/models"
	"github.com/ilya/codestream/internal/ws"
)

const (
	defaultSnapshotInterval = 30 * time.Second
	maxKeptVersions         = 50
)

// DocumentVersionRepository defines the document version repository interface.
type DocumentVersionRepository interface {
	Create(ctx context.Context, fileID, createdBy uuid.UUID, payload []byte) (*models.DocumentVersion, error)
	FindRecentByFileID(ctx context.Context, fileID uuid.UUID, limit int) ([]models.DocumentVersion, error)
	DeleteOldByFileID(ctx context.Context, fileID uuid.UUID, keep int) error
}

// FileRepositoryForDocumentState defines the file repository interface used by document state manager.
type FileRepositoryForDocumentState interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.File, error)
	Update(ctx context.Context, file *models.File) error
}

// DocUpdateData represents the data payload of a doc:update message.
type DocUpdateData struct {
	FileID string `json:"fileId"`
	Update string `json:"update"`
}

// DocumentStateManager manages CRDT document state.
type DocumentStateManager struct {
	hub                 *ws.Hub
	fileRepo            FileRepositoryForDocumentState
	documentVersionRepo DocumentVersionRepository
	logger              *slog.Logger

	mu        sync.RWMutex
	snapshots map[uuid.UUID][]byte
}

// NewDocumentStateManager creates a new DocumentStateManager.
func NewDocumentStateManager(hub *ws.Hub, fileRepo FileRepositoryForDocumentState, documentVersionRepo DocumentVersionRepository, logger *slog.Logger) *DocumentStateManager {
	return &DocumentStateManager{
		hub:                 hub,
		fileRepo:            fileRepo,
		documentVersionRepo: documentVersionRepo,
		logger:              logger,
		snapshots:           make(map[uuid.UUID][]byte),
	}
}

// ApplyUpdate saves a CRDT update and broadcasts it to the room.
func (m *DocumentStateManager) ApplyUpdate(ctx context.Context, fileID, userID uuid.UUID, update []byte) error {
	if _, err := m.documentVersionRepo.Create(ctx, fileID, userID, update); err != nil {
		return err
	}

	m.mu.Lock()
	m.snapshots[fileID] = append(m.snapshots[fileID], update...)
	m.mu.Unlock()

	roomID := "file:" + fileID.String()
	message, err := json.Marshal(ws.WSMessage{
		Event: ws.EventDocUpdate,
		Data: mustJSON(map[string]interface{}{
			"fileId": fileID.String(),
			"update": base64.StdEncoding.EncodeToString(update),
			"userId": userID.String(),
		}),
	})
	if err != nil {
		return err
	}

	m.hub.Broadcast(roomID, message, "")

	return nil
}

// GetSnapshot returns the current in-memory snapshot, or loads from DB if not cached.
func (m *DocumentStateManager) GetSnapshot(ctx context.Context, fileID uuid.UUID) ([]byte, error) {
	m.mu.RLock()
	snapshot := m.snapshots[fileID]
	m.mu.RUnlock()

	if len(snapshot) > 0 {
		return snapshot, nil
	}

	file, err := m.fileRepo.FindByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	if file == nil {
		return nil, errors.New("file not found")
	}

	m.mu.Lock()
	m.snapshots[fileID] = file.Content
	m.mu.Unlock()

	return file.Content, nil
}

// SaveSnapshot persists the current snapshot and prunes old versions.
func (m *DocumentStateManager) SaveSnapshot(ctx context.Context, fileID uuid.UUID) error {
	m.mu.RLock()
	snapshot := m.snapshots[fileID]
	m.mu.RUnlock()

	if len(snapshot) == 0 {
		return nil
	}

	file, err := m.fileRepo.FindByID(ctx, fileID)
	if err != nil {
		return err
	}

	if file == nil {
		return errors.New("file not found")
	}

	file.Content = snapshot
	file.ContentType = "yjs-binary"

	if err := m.fileRepo.Update(ctx, file); err != nil {
		return err
	}

	return m.documentVersionRepo.DeleteOldByFileID(ctx, fileID, maxKeptVersions)
}

// LoadSnapshot loads the snapshot from the database into memory.
func (m *DocumentStateManager) LoadSnapshot(ctx context.Context, fileID uuid.UUID) error {
	file, err := m.fileRepo.FindByID(ctx, fileID)
	if err != nil {
		return err
	}

	if file == nil {
		return errors.New("file not found")
	}

	m.mu.Lock()
	m.snapshots[fileID] = file.Content
	m.mu.Unlock()

	return nil
}

// StartSnapshotWorker starts a background worker that periodically saves snapshots.
func (m *DocumentStateManager) StartSnapshotWorker(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = defaultSnapshotInterval
	}

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.saveAllSnapshots(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (m *DocumentStateManager) saveAllSnapshots(ctx context.Context) {
	m.mu.RLock()
	fileIDs := make([]uuid.UUID, 0, len(m.snapshots))
	for id := range m.snapshots {
		fileIDs = append(fileIDs, id)
	}
	m.mu.RUnlock()

	for _, fileID := range fileIDs {
		if err := m.SaveSnapshot(ctx, fileID); err != nil {
			m.logger.Warn("failed to save snapshot", slog.String("error", err.Error()), slog.String("fileID", fileID.String()))
		}
	}
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
