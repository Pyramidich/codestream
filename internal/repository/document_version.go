package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ilya/codestream/internal/models"
)

// DocumentVersionRepository provides data access for document versions.
type DocumentVersionRepository struct {
	db *gorm.DB
}

// NewDocumentVersionRepository creates a new DocumentVersionRepository.
func NewDocumentVersionRepository(db *gorm.DB) *DocumentVersionRepository {
	return &DocumentVersionRepository{db: db}
}

// Create inserts a new document version into the database.
func (r *DocumentVersionRepository) Create(ctx context.Context, fileID, createdBy uuid.UUID, payload []byte) (*models.DocumentVersion, error) {
	version := &models.DocumentVersion{
		FileID:        fileID,
		UpdatePayload: payload,
		CreatedBy:     &createdBy,
	}

	if err := r.db.WithContext(ctx).Create(version).Error; err != nil {
		return nil, err
	}

	return version, nil
}

// FindRecentByFileID returns the most recent document versions for a file.
func (r *DocumentVersionRepository) FindRecentByFileID(ctx context.Context, fileID uuid.UUID, limit int) ([]models.DocumentVersion, error) {
	var versions []models.DocumentVersion
	if err := r.db.WithContext(ctx).
		Where("file_id = ?", fileID).
		Order("created_at DESC").
		Limit(limit).
		Find(&versions).Error; err != nil {
		return nil, err
	}

	return versions, nil
}

// DeleteOldByFileID removes old document versions for a file, keeping the most recent ones.
func (r *DocumentVersionRepository) DeleteOldByFileID(ctx context.Context, fileID uuid.UUID, keep int) error {
	subQuery := r.db.WithContext(ctx).
		Model(&models.DocumentVersion{}).
		Select("id").
		Where("file_id = ?", fileID).
		Order("created_at DESC").
		Limit(1000000)

	return r.db.WithContext(ctx).
		Where("file_id = ?", fileID).
		Where("id NOT IN (?)", subQuery).
		Delete(&models.DocumentVersion{}).Error
}
