package authz

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/ilya/codestream/internal/models"
)

// ProjectRole represents a role in a project.
type ProjectRole string

const (
	RoleOwner  ProjectRole = "owner"
	RoleEditor ProjectRole = "editor"
	RoleViewer ProjectRole = "viewer"
)

// MemberProvider provides project membership information.
type MemberProvider interface {
	FindByProjectAndUser(ctx context.Context, projectID, userID uuid.UUID) (*models.ProjectMember, error)
}

// Authorization provides resource-level authorization checks.
type Authorization struct {
	memberProvider MemberProvider
}

// NewAuthorization creates a new Authorization instance.
func NewAuthorization(memberProvider MemberProvider) *Authorization {
	return &Authorization{memberProvider: memberProvider}
}

// CanAccessProject checks if a user can access a project.
func (a *Authorization) CanAccessProject(ctx context.Context, userID, projectID uuid.UUID) (bool, error) {
	member, err := a.memberProvider.FindByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		return false, err
	}

	return member != nil, nil
}

// CanManageProject checks if a user can manage a project (only owner).
func (a *Authorization) CanManageProject(ctx context.Context, userID, projectID uuid.UUID) (bool, error) {
	member, err := a.memberProvider.FindByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		return false, err
	}

	return member != nil && ProjectRole(member.Role) == RoleOwner, nil
}

// CanEditFile checks if a user can edit files in a project (owner/editor).
func (a *Authorization) CanEditFile(ctx context.Context, userID, projectID uuid.UUID) (bool, error) {
	member, err := a.memberProvider.FindByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		return false, err
	}

	if member == nil {
		return false, nil
	}

	role := ProjectRole(member.Role)
	return role == RoleOwner || role == RoleEditor, nil
}

// GetRole returns the user's role in a project or an error if not a member.
func (a *Authorization) GetRole(ctx context.Context, userID, projectID uuid.UUID) (ProjectRole, error) {
	member, err := a.memberProvider.FindByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		return "", err
	}

	if member == nil {
		return "", errors.New("not a project member")
	}

	return ProjectRole(member.Role), nil
}
