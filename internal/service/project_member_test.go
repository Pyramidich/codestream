package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilya/codestream/internal/authz"
	"github.com/ilya/codestream/internal/models"
	"github.com/ilya/codestream/internal/service"
)

type mockUserRepoForMember struct {
	users []*models.User
}

func (m *mockUserRepoForMember) Create(ctx context.Context, email, passwordHash, displayName string) (*models.User, error) {
	user := &models.User{ID: uuid.New(), Email: email, PasswordHash: passwordHash, DisplayName: displayName}
	m.users = append(m.users, user)
	return user, nil
}

func (m *mockUserRepoForMember) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepoForMember) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

type mockMemberRepo struct {
	members []*models.ProjectMember
}

func (m *mockMemberRepo) Create(ctx context.Context, projectID, userID uuid.UUID, role string) (*models.ProjectMember, error) {
	member := &models.ProjectMember{ID: uuid.New(), ProjectID: projectID, UserID: userID, Role: role}
	m.members = append(m.members, member)
	return member, nil
}

func (m *mockMemberRepo) FindByProjectAndUser(ctx context.Context, projectID, userID uuid.UUID) (*models.ProjectMember, error) {
	for _, member := range m.members {
		if member.ProjectID == projectID && member.UserID == userID {
			return member, nil
		}
	}
	return nil, nil
}

func (m *mockMemberRepo) FindByProjectID(ctx context.Context, projectID uuid.UUID) ([]models.ProjectMember, error) {
	var result []models.ProjectMember
	for _, member := range m.members {
		if member.ProjectID == projectID {
			result = append(result, *member)
		}
	}
	return result, nil
}

func (m *mockMemberRepo) Delete(ctx context.Context, projectID, userID uuid.UUID) error {
	for i, member := range m.members {
		if member.ProjectID == projectID && member.UserID == userID {
			m.members = append(m.members[:i], m.members[i+1:]...)
			return nil
		}
	}
	return nil
}

func newTestMemberService() (*service.ProjectMemberService, *mockUserRepoForMember, *mockMemberRepo) {
	userRepo := &mockUserRepoForMember{}
	memberRepo := &mockMemberRepo{}
	authzInstance := authz.NewAuthorization(memberRepo)
	return service.NewProjectMemberService(memberRepo, userRepo, authzInstance), userRepo, memberRepo
}

func TestAddMember(t *testing.T) {
	svc, userRepo, memberRepo := newTestMemberService()
	ownerID := uuid.New()
	projectID := uuid.New()

	// Owner is a member with owner role
	_, err := memberRepo.Create(context.Background(), projectID, ownerID, "owner")
	require.NoError(t, err)

	user, err := userRepo.Create(context.Background(), "member@example.com", "hash", "Member")
	require.NoError(t, err)

	member, err := svc.AddMember(context.Background(), projectID, ownerID, user.ID, "editor")
	require.NoError(t, err)
	assert.Equal(t, "editor", member.Role)
}

func TestRemoveMember(t *testing.T) {
	svc, userRepo, memberRepo := newTestMemberService()
	ownerID := uuid.New()
	projectID := uuid.New()

	_, err := memberRepo.Create(context.Background(), projectID, ownerID, "owner")
	require.NoError(t, err)

	user, err := userRepo.Create(context.Background(), "member@example.com", "hash", "Member")
	require.NoError(t, err)

	_, err = svc.AddMember(context.Background(), projectID, ownerID, user.ID, "editor")
	require.NoError(t, err)

	err = svc.RemoveMember(context.Background(), projectID, ownerID, user.ID)
	require.NoError(t, err)
}
