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
	FindByProjectIDWithUser(ctx context.Context, projectID uuid.UUID) ([]models.ProjectMember, error)
	Delete(ctx context.Context, projectID, userID uuid.UUID) error
}

// ProjectMemberService provides project membership operations.
type ProjectMemberService struct {
	memberRepo ProjectMemberRepository
	userRepo   UserRepository
	authz      *authz.Authorization
	changeHist ChangeHistoryLogger
}

// MemberInfo represents the public data for a project member.
type MemberInfo struct {
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Role        string    `json:"role"`
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

// AddMember adds a new member to a project (owner or editor can add).
func (s *ProjectMemberService) AddMember(ctx context.Context, projectID, currentUserID uuid.UUID, email, role string) (*models.ProjectMember, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, errors.New("invalid email")
	}

	canAdd, err := s.authz.CanAddMembers(ctx, currentUserID, projectID)
	if err != nil {
		return nil, err
	}

	if !canAdd {
		return nil, errors.New("only owner or editor can add members")
	}

	role = strings.ToLower(role)
	if role != string(authz.RoleOwner) && role != string(authz.RoleEditor) && role != string(authz.RoleViewer) {
		return nil, errors.New("invalid role")
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	if currentUserID == user.ID {
		return nil, errors.New("cannot add yourself")
	}

	existing, err := s.memberRepo.FindByProjectAndUser(ctx, projectID, user.ID)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, errors.New("user is already a member")
	}

	member, err := s.memberRepo.Create(ctx, projectID, user.ID, role)
	if err != nil {
		return nil, err
	}

	s.changeHist.Log(ctx, currentUserID, "project.member_added", nil, &projectID, map[string]interface{}{
		"user_id": user.ID.String(),
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

	members, err := s.memberRepo.FindByProjectID(ctx, projectID)
	if err != nil {
		return err
	}

	var ownerCount int
	var targetMember *models.ProjectMember
	for i := range members {
		if authz.ProjectRole(members[i].Role) == authz.RoleOwner {
			ownerCount++
		}
		if members[i].UserID == userID {
			targetMember = &members[i]
		}
	}

	if targetMember == nil {
		return errors.New("member not found")
	}

	if authz.ProjectRole(targetMember.Role) == authz.RoleOwner && ownerCount <= 1 {
		return errors.New("cannot remove the last owner")
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
func (s *ProjectMemberService) ListMembers(ctx context.Context, projectID, userID uuid.UUID) ([]MemberInfo, error) {
	canAccess, err := s.authz.CanAccessProject(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	if !canAccess {
		return nil, errors.New("access denied")
	}

	members, err := s.memberRepo.FindByProjectIDWithUser(ctx, projectID)
	if err != nil {
		return nil, err
	}

	infos := make([]MemberInfo, 0, len(members))
	for _, member := range members {
		info := MemberInfo{
			UserID:      member.UserID,
			Email:       member.User.Email,
			DisplayName: member.User.DisplayName,
			Role:        member.Role,
		}
		infos = append(infos, info)
	}

	return infos, nil
}
