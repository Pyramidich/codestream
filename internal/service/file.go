package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/ilya/codestream/internal/authz"
	"github.com/ilya/codestream/internal/models"
)

// FileRepository defines the file repository interface.
type FileRepository interface {
	Create(ctx context.Context, file *models.File) (*models.File, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.File, error)
	FindByProjectID(ctx context.Context, projectID uuid.UUID) ([]models.File, error)
	FindByProjectIDAndPath(ctx context.Context, projectID uuid.UUID, path string) (*models.File, error)
	Update(ctx context.Context, file *models.File) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ChangeHistoryLogger defines the change history logging interface.
type ChangeHistoryLogger interface {
	Log(ctx context.Context, userID uuid.UUID, action string, fileID, projectID *uuid.UUID, metadata map[string]interface{})
}

// FileService provides file management operations.
type FileService struct {
	fileRepo    FileRepository
	authz       *authz.Authorization
	changeHist  ChangeHistoryLogger
}

// NewFileService creates a new FileService.
func NewFileService(fileRepo FileRepository, authz *authz.Authorization, changeHist ChangeHistoryLogger) *FileService {
	return &FileService{
		fileRepo:   fileRepo,
		authz:      authz,
		changeHist: changeHist,
	}
}

// Create creates a new file in a project (owner/editor).
func (s *FileService) Create(ctx context.Context, projectID, userID uuid.UUID, name, path, language string) (*models.File, error) {
	canEdit, err := s.authz.CanEditFile(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	if !canEdit {
		return nil, errors.New("insufficient permissions")
	}

	name = strings.TrimSpace(name)
	path = strings.TrimSpace(path)

	if name == "" || path == "" {
		return nil, errors.New("name and path are required")
	}

	existing, err := s.fileRepo.FindByProjectIDAndPath(ctx, projectID, path)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, errors.New("file with this path already exists")
	}

	emptyText := ""
	file := &models.File{
		ProjectID:   projectID,
		Name:        name,
		Path:        path,
		Language:    language,
		Content:     []byte{},
		ContentText: &emptyText,
		ContentType: "yjs-binary",
	}

	created, err := s.fileRepo.Create(ctx, file)
	if err != nil {
		return nil, err
	}

	s.changeHist.Log(ctx, userID, "file.created", &created.ID, &projectID, map[string]interface{}{
		"name":     created.Name,
		"path":     created.Path,
		"language": created.Language,
	})

	return created, nil
}

// GetByID returns a file if the user can access the project.
func (s *FileService) GetByID(ctx context.Context, fileID, userID uuid.UUID) (*models.File, error) {
	file, err := s.fileRepo.FindByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	if file == nil {
		return nil, errors.New("file not found")
	}

	canAccess, err := s.authz.CanAccessProject(ctx, userID, file.ProjectID)
	if err != nil {
		return nil, err
	}

	if !canAccess {
		return nil, errors.New("access denied")
	}

	return file, nil
}

// ListByProject returns all files in a project if the user is a member.
func (s *FileService) ListByProject(ctx context.Context, projectID, userID uuid.UUID) ([]models.File, error) {
	canAccess, err := s.authz.CanAccessProject(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	if !canAccess {
		return nil, errors.New("access denied")
	}

	return s.fileRepo.FindByProjectID(ctx, projectID)
}

// FileUpdates represents fields that can be updated on a file.
type FileUpdates struct {
	Name        *string
	Path        *string
	Language    *string
	Content     *[]byte
	ContentText *string
}

// Update updates a file (owner/editor).
func (s *FileService) Update(ctx context.Context, fileID, userID uuid.UUID, updates FileUpdates) (*models.File, error) {
	file, err := s.fileRepo.FindByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	if file == nil {
		return nil, errors.New("file not found")
	}

	canEdit, err := s.authz.CanEditFile(ctx, userID, file.ProjectID)
	if err != nil {
		return nil, err
	}

	if !canEdit {
		return nil, errors.New("insufficient permissions")
	}

	if updates.Name != nil {
		name := strings.TrimSpace(*updates.Name)
		if name != "" {
			file.Name = name
		}
	}

	if updates.Path != nil {
		path := strings.TrimSpace(*updates.Path)
		if path != "" && path != file.Path {
			existing, err := s.fileRepo.FindByProjectIDAndPath(ctx, file.ProjectID, path)
			if err != nil {
				return nil, err
			}
			if existing != nil {
				return nil, errors.New("file with this path already exists")
			}
			file.Path = path
		}
	}

	if updates.Language != nil {
		file.Language = *updates.Language
	}

	if updates.Content != nil {
		file.Content = *updates.Content
		file.ContentType = "yjs-binary"
	}

	if updates.ContentText != nil {
		file.ContentText = updates.ContentText
	}

	changedFields := []string{}
	if updates.Name != nil {
		changedFields = append(changedFields, "name")
	}
	if updates.Path != nil {
		changedFields = append(changedFields, "path")
	}
	if updates.Language != nil {
		changedFields = append(changedFields, "language")
	}
	if updates.Content != nil {
		changedFields = append(changedFields, "content")
	}
	if updates.ContentText != nil {
		changedFields = append(changedFields, "content_text")
	}

	if err := s.fileRepo.Update(ctx, file); err != nil {
		return nil, err
	}

	s.changeHist.Log(ctx, userID, "file.updated", &file.ID, &file.ProjectID, map[string]interface{}{
		"changed_fields": changedFields,
	})

	return file, nil
}

// Delete deletes a file (owner/editor).
func (s *FileService) Delete(ctx context.Context, fileID, userID uuid.UUID) error {
	file, err := s.fileRepo.FindByID(ctx, fileID)
	if err != nil {
		return err
	}

	if file == nil {
		return errors.New("file not found")
	}

	canEdit, err := s.authz.CanEditFile(ctx, userID, file.ProjectID)
	if err != nil {
		return err
	}

	if !canEdit {
		return errors.New("insufficient permissions")
	}

	s.changeHist.Log(ctx, userID, "file.deleted", &file.ID, &file.ProjectID, map[string]interface{}{
		"id":   file.ID.String(),
		"name": file.Name,
	})

	return s.fileRepo.Delete(ctx, fileID)
}
