package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/ilya/codestream/internal/models"
	"github.com/ilya/codestream/internal/service"
)

// ChangeHistoryRepository provides data access for change history.
type ChangeHistoryRepository struct {
	db *gorm.DB
}

// NewChangeHistoryRepository creates a new ChangeHistoryRepository.
func NewChangeHistoryRepository(db *gorm.DB) *ChangeHistoryRepository {
	return &ChangeHistoryRepository{db: db}
}

// Create inserts a new change history entry.
func (r *ChangeHistoryRepository) Create(ctx context.Context, userID uuid.UUID, action string, fileID, projectID *uuid.UUID, metadata map[string]interface{}) (*service.ChangeHistoryEntry, error) {
	metaJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	entry := &models.ChangeHistory{
		UserID:    userID,
		Action:    action,
		FileID:    fileID,
		ProjectID: projectID,
		Metadata:  string(datatypes.JSON(metaJSON)),
	}

	if err := r.db.WithContext(ctx).Create(entry).Error; err != nil {
		return nil, err
	}

	return &service.ChangeHistoryEntry{
		ID:        entry.ID,
		UserID:    entry.UserID,
		Action:    entry.Action,
		FileID:    entry.FileID,
		ProjectID: entry.ProjectID,
		Metadata:  metadata,
		CreatedAt: entry.CreatedAt,
	}, nil
}
