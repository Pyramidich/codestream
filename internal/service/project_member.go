package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/ilya/codestream/internal/authz"
	"github.com/ilya/codestream/internal/models"
)

// ProjectMemberRepository defines the project member repository interface.
type ProjectMemberRepository interface {
	Create(ctx context.Context, projectID, userID uuid.UUID, role string) (*models.ProjectMember, error)
	FindByProjectAndUser(ctx context.Context, projectID, userID uuid.UUID) (*models.ProjectMember, error)
	FindByProjectID(ctx context.Context, projectID uuid.UUID) ([]models.ProjectMember, error)
	Delete(ctx context.Context, projectID, userID uuid.UUID) error
}

// ProjectMemberService provides project membership operations.
type ProjectMemberService struct {
	memberRepo   ProjectMemberRepository
	userRepo     UserRepository
	authz        *authz.Authorization
	changeHist   ChangeHistoryLogger
}

// NewProjectMemberService creates a new ProjectMemberService.
func NewProjectMemberService(memberRepo ProjectMemberRepository, userRepo UserRepository, authz *authz.Authorization, changeHist ChangeHistoryLogger) *ProjectMemberService {
	return &ProjectMemberService{
		memberRepo: memberRepo,
		userRepo:   userRepo,
		authz:      authz,
		changeHist: changeHist,
	}
}

// AddMember adds a new member to a project (only owner).
func (s *ProjectMemberService) AddMember(ctx context.Context, projectID, ownerID, userID uuid.UUID, role string) (*models.ProjectMember, error) {
	if ownerID == userID {
		return nil, errors.New("cannot add yourself")
	}

	canManage, err := s.authz.CanManageProject(ctx, ownerID, projectID)
	if err != nil {
		return nil, err
	}

	if !canManage {
		return nil, errors.New("only owner can add members")
	}

	role = strings.ToLower(role)
	if role != string(authz.RoleOwner) && role != string(authz.RoleEditor) && role != string(authz.RoleViewer) {
		return nil, errors.New("invalid role")
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	existing, err := s.memberRepo.FindByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, errors.New("user is already a member")
	}

	member, err := s.memberRepo.Create(ctx, projectID, userID, role)
	if err != nil {
		return nil, err
	}

	s.changeHist.Log(ctx, ownerID, "project.member_added", nil, &projectID, map[string]interface{}{
		"user_id": userID.String(),
		"role":    role,
	})

	return member, nil
}

// RemoveMember removes a member from a project (only owner).
func (s *ProjectMemberService) RemoveMember(ctx context.Context, projectID, ownerID, userID uuid.UUID) error {
	if ownerID == userID {
		return errors.New("cannot remove yourself")
	}

	canManage, err := s.authz.CanManageProject(ctx, ownerID, projectID)
	if err != nil {
		return err
	}

	if !canManage {
		return errors.New("only owner can remove members")
	}

	if err := s.memberRepo.Delete(ctx, projectID, userID); err != nil {
		return err
	}

	s.changeHist.Log(ctx, ownerID, "project.member_removed", nil, &projectID, map[string]interface{}{
		"user_id": userID.String(),
	})

	return nil
}

// ListMembers returns all members of a project (any member).
func (s *ProjectMemberService) ListMembers(ctx context.Context, projectID, userID uuid.UUID) ([]models.ProjectMember, error) {
	canAccess, err := s.authz.CanAccessProject(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	if !canAccess {
		return nil, errors.New("access denied")
	}

	return s.memberRepo.FindByProjectID(ctx, projectID)
}
