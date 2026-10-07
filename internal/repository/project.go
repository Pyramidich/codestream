package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ilya/codestream/internal/models"
)

// ProjectRepository provides data access for projects.
type ProjectRepository struct {
	db *gorm.DB
}

// NewProjectRepository creates a new ProjectRepository.
func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// Create inserts a new project into the database.
func (r *ProjectRepository) Create(ctx context.Context, name string, ownerID uuid.UUID) (*models.Project, error) {
	project := &models.Project{
		Name:    name,
		OwnerID: ownerID,
	}

	if err := r.db.WithContext(ctx).Create(project).Error; err != nil {
		return nil, err
	}

	return project, nil
}

// FindByID finds a project by ID.
func (r *ProjectRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	var project models.Project
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &project, nil
}

// FindByUserID returns all projects where the user is a member.
func (r *ProjectRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Project, error) {
	var projects []*models.Project

	err := r.db.WithContext(ctx).
		Joins("JOIN project_members ON project_members.project_id = projects.id").
		Where("project_members.user_id = ?", userID).
		Find(&projects).Error

	return projects, err
}

// Update updates a project.
func (r *ProjectRepository) Update(ctx context.Context, project *models.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

// Delete deletes a project by ID.
func (r *ProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Project{}, id).Error
}
