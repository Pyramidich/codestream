package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

// ChangeHistoryEntry represents a change history record.
type ChangeHistoryEntry struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Action    string
	FileID    *uuid.UUID
	ProjectID *uuid.UUID
	Metadata  map[string]interface{}
	CreatedAt interface{}
}

// ChangeHistoryRepository defines the change history repository interface.
type ChangeHistoryRepository interface {
	Create(ctx context.Context, userID uuid.UUID, action string, fileID, projectID *uuid.UUID, metadata map[string]interface{}) (*ChangeHistoryEntry, error)
}

// ChangeHistoryService provides change history logging operations.
type ChangeHistoryService struct {
	repo   ChangeHistoryRepository
	logger *slog.Logger
}

// NewChangeHistoryService creates a new ChangeHistoryService.
func NewChangeHistoryService(repo ChangeHistoryRepository, logger *slog.Logger) *ChangeHistoryService {
	return &ChangeHistoryService{
		repo:   repo,
		logger: logger,
	}
}

// Log records a change history event asynchronously.
func (s *ChangeHistoryService) Log(ctx context.Context, userID uuid.UUID, action string, fileID, projectID *uuid.UUID, metadata map[string]interface{}) {
	go func() {
		if metadata == nil {
			metadata = make(map[string]interface{})
		}

		if _, err := s.repo.Create(ctx, userID, action, fileID, projectID, metadata); err != nil {
			s.logger.Warn(
				"failed to log change history",
				slog.String("error", err.Error()),
				slog.String("action", action),
			)
		}
	}()
}
