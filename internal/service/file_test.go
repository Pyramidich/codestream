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

type mockFileRepo struct {
	files []*models.File
}

func (m *mockFileRepo) Create(ctx context.Context, file *models.File) (*models.File, error) {
	file.ID = uuid.New()
	m.files = append(m.files, file)
	return file, nil
}

func (m *mockFileRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.File, error) {
	for _, f := range m.files {
		if f.ID == id {
			return f, nil
		}
	}
	return nil, nil
}

func (m *mockFileRepo) FindByProjectID(ctx context.Context, projectID uuid.UUID) ([]models.File, error) {
	var files []models.File
	for _, f := range m.files {
		if f.ProjectID == projectID {
			files = append(files, *f)
		}
	}
	return files, nil
}

func (m *mockFileRepo) FindByProjectIDAndPath(ctx context.Context, projectID uuid.UUID, path string) (*models.File, error) {
	for _, f := range m.files {
		if f.ProjectID == projectID && f.Path == path {
			return f, nil
		}
	}
	return nil, nil
}

func (m *mockFileRepo) Update(ctx context.Context, file *models.File) error {
	for i, f := range m.files {
		if f.ID == file.ID {
			m.files[i] = file
			return nil
		}
	}
	return nil
}

func (m *mockFileRepo) Delete(ctx context.Context, id uuid.UUID) error {
	for i, f := range m.files {
		if f.ID == id {
			m.files = append(m.files[:i], m.files[i+1:]...)
			return nil
		}
	}
	return nil
}

type mockMemberRepoForFile struct {
	members []*models.ProjectMember
}

func (m *mockMemberRepoForFile) Create(ctx context.Context, projectID, userID uuid.UUID, role string) (*models.ProjectMember, error) {
	member := &models.ProjectMember{ID: uuid.New(), ProjectID: projectID, UserID: userID, Role: role}
	m.members = append(m.members, member)
	return member, nil
}

func (m *mockMemberRepoForFile) FindByProjectAndUser(ctx context.Context, projectID, userID uuid.UUID) (*models.ProjectMember, error) {
	for _, member := range m.members {
		if member.ProjectID == projectID && member.UserID == userID {
			return member, nil
		}
	}
	return nil, nil
}

func (m *mockMemberRepoForFile) FindByProjectID(ctx context.Context, projectID uuid.UUID) ([]models.ProjectMember, error) {
	return nil, nil
}

func (m *mockMemberRepoForFile) Delete(ctx context.Context, projectID, userID uuid.UUID) error {
	return nil
}

func newTestFileService() (*service.FileService, *mockFileRepo, *mockMemberRepoForFile) {
	fileRepo := &mockFileRepo{}
	memberRepo := &mockMemberRepoForFile{}
	authzInstance := authz.NewAuthorization(memberRepo)
	return service.NewFileService(fileRepo, authzInstance), fileRepo, memberRepo
}

func TestFileCreate(t *testing.T) {
	svc, _, memberRepo := newTestFileService()
	ownerID := uuid.New()
	projectID := uuid.New()

	_, err := memberRepo.Create(context.Background(), projectID, ownerID, "owner")
	require.NoError(t, err)

	file, err := svc.Create(context.Background(), projectID, ownerID, "main.go", "/main.go", "go")
	require.NoError(t, err)
	assert.Equal(t, "main.go", file.Name)
	assert.Equal(t, "/main.go", file.Path)
}

func TestFileCreateViewerForbidden(t *testing.T) {
	svc, _, memberRepo := newTestFileService()
	viewerID := uuid.New()
	projectID := uuid.New()

	_, err := memberRepo.Create(context.Background(), projectID, viewerID, "viewer")
	require.NoError(t, err)

	_, err = svc.Create(context.Background(), projectID, viewerID, "main.go", "/main.go", "go")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient permissions")
}
