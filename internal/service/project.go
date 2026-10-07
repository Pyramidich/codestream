package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/ilya/codestream/internal/authz"
	"github.com/ilya/codestream/internal/models"
)

// ProjectRepository defines the project repository interface.
type ProjectRepository interface {
	Create(ctx context.Context, name string, ownerID uuid.UUID) (*models.Project, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Project, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Project, error)
	Update(ctx context.Context, project *models.Project) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ProjectMemberRepositoryForProject defines the member repository interface used by project service.
type ProjectMemberRepositoryForProject interface {
	Create(ctx context.Context, projectID, userID uuid.UUID, role string) (*models.ProjectMember, error)
	FindByProjectAndUser(ctx context.Context, projectID, userID uuid.UUID) (*models.ProjectMember, error)
}

// ProjectService provides project management operations.
type ProjectService struct {
	projectRepo ProjectRepository
	memberRepo  ProjectMemberRepositoryForProject
	authz       *authz.Authorization
	changeHist  ChangeHistoryLogger
}

// NewProjectService creates a new ProjectService.
func NewProjectService(projectRepo ProjectRepository, memberRepo ProjectMemberRepositoryForProject, authz *authz.Authorization, changeHist ChangeHistoryLogger) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		memberRepo:  memberRepo,
		authz:       authz,
		changeHist:  changeHist,
	}
}

// Create creates a new project and makes the creator an owner.
func (s *ProjectService) Create(ctx context.Context, name string, ownerID uuid.UUID) (*models.Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("project name is required")
	}

	project, err := s.projectRepo.Create(ctx, name, ownerID)
	if err != nil {
		return nil, err
	}

	_, err = s.memberRepo.Create(ctx, project.ID, ownerID, string(authz.RoleOwner))
	if err != nil {
		return nil, err
	}

	return project, nil
}

// GetByID returns a project if the user is a member.
func (s *ProjectService) GetByID(ctx context.Context, projectID, userID uuid.UUID) (*models.Project, error) {
	canAccess, err := s.authz.CanAccessProject(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	if !canAccess {
		return nil, errors.New("access denied")
	}

	return s.projectRepo.FindByID(ctx, projectID)
}

// ListForUser returns all projects the user is a member of.
func (s *ProjectService) ListForUser(ctx context.Context, userID uuid.UUID) ([]*models.Project, error) {
	return s.projectRepo.FindByUserID(ctx, userID)
}

// Update updates a project name (only owner).
func (s *ProjectService) Update(ctx context.Context, projectID, userID uuid.UUID, name string) (*models.Project, error) {
	canManage, err := s.authz.CanManageProject(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	if !canManage {
		return nil, errors.New("only owner can update project")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("project name is required")
	}

	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if project == nil {
		return nil, errors.New("project not found")
	}

	project.Name = name
	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

// Delete deletes a project (only owner).
func (s *ProjectService) Delete(ctx context.Context, projectID, userID uuid.UUID) error {
	canManage, err := s.authz.CanManageProject(ctx, userID, projectID)
	if err != nil {
		return err
	}

	if !canManage {
		return errors.New("only owner can delete project")
	}

	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return err
	}

	if project == nil {
		return errors.New("project not found")
	}

	s.changeHist.Log(ctx, userID, "project.deleted", nil, &projectID, map[string]interface{}{
		"id":   project.ID.String(),
		"name": project.Name,
	})

	return s.projectRepo.Delete(ctx, projectID)
}
