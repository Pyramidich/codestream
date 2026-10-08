package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ilya/codestream/internal/models"
)

// FileRepository provides data access for files.
type FileRepository struct {
	db *gorm.DB
}

// NewFileRepository creates a new FileRepository.
func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

// Create inserts a new file into the database.
func (r *FileRepository) Create(ctx context.Context, file *models.File) (*models.File, error) {
	if err := r.db.WithContext(ctx).Create(file).Error; err != nil {
		return nil, err
	}

	return file, nil
}

// FindByID finds a file by ID.
func (r *FileRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.File, error) {
	var file models.File
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &file, nil
}

// FindByProjectID returns all files in a project.
func (r *FileRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID) ([]models.File, error) {
	var files []models.File
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&files).Error; err != nil {
		return nil, err
	}

	return files, nil
}

// FindByProjectIDAndPath finds a file by project ID and path.
func (r *FileRepository) FindByProjectIDAndPath(ctx context.Context, projectID uuid.UUID, path string) (*models.File, error) {
	var file models.File
	if err := r.db.WithContext(ctx).Where("project_id = ? AND path = ?", projectID, path).First(&file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &file, nil
}

// Update updates a file.
func (r *FileRepository) Update(ctx context.Context, file *models.File) error {
	return r.db.WithContext(ctx).Model(file).
		Select("name", "path", "language", "content", "content_text", "content_type", "updated_at").
		Updates(file).Error
}

// Delete deletes a file by ID.
func (r *FileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.File{}, id).Error
}
