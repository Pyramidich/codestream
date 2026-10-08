package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ilya/codestream/internal/models"
)

// ProjectMemberRepository provides data access for project members.
type ProjectMemberRepository struct {
	db *gorm.DB
}

// NewProjectMemberRepository creates a new ProjectMemberRepository.
func NewProjectMemberRepository(db *gorm.DB) *ProjectMemberRepository {
	return &ProjectMemberRepository{db: db}
}

// Create inserts a new project member into the database.
func (r *ProjectMemberRepository) Create(ctx context.Context, projectID, userID uuid.UUID, role string) (*models.ProjectMember, error) {
	member := &models.ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	}

	if err := r.db.WithContext(ctx).Create(member).Error; err != nil {
		return nil, err
	}

	return member, nil
}

// FindByProjectAndUser finds a membership by project and user IDs.
func (r *ProjectMemberRepository) FindByProjectAndUser(ctx context.Context, projectID, userID uuid.UUID) (*models.ProjectMember, error) {
	var member models.ProjectMember
	if err := r.db.WithContext(ctx).Where("project_id = ? AND user_id = ?", projectID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &member, nil
}

// FindByProjectID returns all members of a project.
func (r *ProjectMemberRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID) ([]models.ProjectMember, error) {
	var members []models.ProjectMember
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&members).Error; err != nil {
		return nil, err
	}

	return members, nil
}

// FindByProjectIDWithUsers returns all members of a project with their user data preloaded.
func (r *ProjectMemberRepository) FindByProjectIDWithUser(ctx context.Context, projectID uuid.UUID) ([]models.ProjectMember, error) {
	var members []models.ProjectMember
	if err := r.db.WithContext(ctx).Preload("User").Where("project_id = ?", projectID).Find(&members).Error; err != nil {
		return nil, err
	}

	return members, nil
}

// Delete removes a project member.
func (r *ProjectMemberRepository) Delete(ctx context.Context, projectID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("project_id = ? AND user_id = ?", projectID, userID).Delete(&models.ProjectMember{}).Error
}
