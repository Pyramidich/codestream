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

type mockProjectRepo struct {
	projects []*models.Project
}

func (m *mockProjectRepo) Create(ctx context.Context, name string, ownerID uuid.UUID) (*models.Project, error) {
	p := &models.Project{ID: uuid.New(), Name: name, OwnerID: ownerID}
	m.projects = append(m.projects, p)
	return p, nil
}

func (m *mockProjectRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	for _, p := range m.projects {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockProjectRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Project, error) {
	return m.projects, nil
}

func (m *mockProjectRepo) Update(ctx context.Context, project *models.Project) error {
	for i, p := range m.projects {
		if p.ID == project.ID {
			m.projects[i] = project
			return nil
		}
	}
	return nil
}

func (m *mockProjectRepo) Delete(ctx context.Context, id uuid.UUID) error {
	for i, p := range m.projects {
		if p.ID == id {
			m.projects = append(m.projects[:i], m.projects[i+1:]...)
			return nil
		}
	}
	return nil
}

type mockMemberRepoForProject struct {
	members []*models.ProjectMember
}

func (m *mockMemberRepoForProject) Create(ctx context.Context, projectID, userID uuid.UUID, role string) (*models.ProjectMember, error) {
	member := &models.ProjectMember{ID: uuid.New(), ProjectID: projectID, UserID: userID, Role: role}
	m.members = append(m.members, member)
	return member, nil
}

func (m *mockMemberRepoForProject) FindByProjectAndUser(ctx context.Context, projectID, userID uuid.UUID) (*models.ProjectMember, error) {
	for _, member := range m.members {
		if member.ProjectID == projectID && member.UserID == userID {
			return member, nil
		}
	}
	return nil, nil
}

func (m *mockMemberRepoForProject) FindByProjectID(ctx context.Context, projectID uuid.UUID) ([]models.ProjectMember, error) {
	return nil, nil
}

func (m *mockMemberRepoForProject) Delete(ctx context.Context, projectID, userID uuid.UUID) error {
	return nil
}

func newTestProjectService() (*service.ProjectService, *mockProjectRepo, *mockMemberRepoForProject) {
	projectRepo := &mockProjectRepo{}
	memberRepo := &mockMemberRepoForProject{}
	authzInstance := authz.NewAuthorization(memberRepo)
	return service.NewProjectService(projectRepo, memberRepo, authzInstance, &noopChangeHistoryLogger{}), projectRepo, memberRepo
}

func TestProjectCreate(t *testing.T) {
	svc, _, _ := newTestProjectService()
	ownerID := uuid.New()

	project, err := svc.Create(context.Background(), "My Project", ownerID)
	require.NoError(t, err)
	assert.Equal(t, "My Project", project.Name)
	assert.Equal(t, ownerID, project.OwnerID)
}

func TestProjectUpdateOnlyOwner(t *testing.T) {
	svc, _, memberRepo := newTestProjectService()
	ownerID := uuid.New()
	otherID := uuid.New()

	project, err := svc.Create(context.Background(), "My Project", ownerID)
	require.NoError(t, err)

	// Add other user as viewer
	_, err = memberRepo.Create(context.Background(), project.ID, otherID, "viewer")
	require.NoError(t, err)

	updated, err := svc.Update(context.Background(), project.ID, ownerID, "New Name")
	require.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)

	_, err = svc.Update(context.Background(), project.ID, otherID, "Hacked")
	assert.Error(t, err)
}

func TestProjectDeleteOnlyOwner(t *testing.T) {
	svc, _, memberRepo := newTestProjectService()
	ownerID := uuid.New()
	otherID := uuid.New()

	project, err := svc.Create(context.Background(), "My Project", ownerID)
	require.NoError(t, err)

	_, err = memberRepo.Create(context.Background(), project.ID, otherID, "editor")
	require.NoError(t, err)

	err = svc.Delete(context.Background(), project.ID, otherID)
	assert.Error(t, err)

	err = svc.Delete(context.Background(), project.ID, ownerID)
	assert.NoError(t, err)
}
