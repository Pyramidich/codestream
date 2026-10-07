package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilya/codestream/internal/logger"
	"github.com/ilya/codestream/internal/models"
	"github.com/ilya/codestream/internal/service"
	"github.com/ilya/codestream/internal/ws"
)

type mockDocumentVersionRepo struct {
	versions []models.DocumentVersion
}

func (m *mockDocumentVersionRepo) Create(ctx context.Context, fileID, createdBy uuid.UUID, payload []byte) (*models.DocumentVersion, error) {
	v := models.DocumentVersion{
		ID:            uuid.New(),
		FileID:        fileID,
		UpdatePayload: payload,
		CreatedBy:     &createdBy,
		CreatedAt:     time.Now(),
	}
	m.versions = append(m.versions, v)
	return &v, nil
}

func (m *mockDocumentVersionRepo) FindRecentByFileID(ctx context.Context, fileID uuid.UUID, limit int) ([]models.DocumentVersion, error) {
	return m.versions, nil
}

func (m *mockDocumentVersionRepo) DeleteOldByFileID(ctx context.Context, fileID uuid.UUID, keep int) error {
	return nil
}

type mockFileRepoForDocumentState struct {
	files []*models.File
}

func (m *mockFileRepoForDocumentState) FindByID(ctx context.Context, id uuid.UUID) (*models.File, error) {
	for _, f := range m.files {
		if f.ID == id {
			return f, nil
		}
	}
	return nil, nil
}

func (m *mockFileRepoForDocumentState) Update(ctx context.Context, file *models.File) error {
	for i, f := range m.files {
		if f.ID == file.ID {
			m.files[i] = file
			return nil
		}
	}
	return nil
}

func TestApplyUpdate(t *testing.T) {
	hub := ws.NewHub()
	fileRepo := &mockFileRepoForDocumentState{}
	dvRepo := &mockDocumentVersionRepo{}
	log := logger.New("dev", "info")

	fileID := uuid.New()
	userID := uuid.New()
	fileRepo.files = append(fileRepo.files, &models.File{ID: fileID, ProjectID: uuid.New(), Name: "test"})

	dsm := service.NewDocumentStateManager(hub, fileRepo, dvRepo, log)

	update := []byte("fake-update")
	err := dsm.ApplyUpdate(context.Background(), fileID, userID, update)
	require.NoError(t, err)

	assert.Len(t, dvRepo.versions, 1)
	assert.Equal(t, update, dvRepo.versions[0].UpdatePayload)
}

func TestSaveSnapshot(t *testing.T) {
	hub := ws.NewHub()
	fileRepo := &mockFileRepoForDocumentState{}
	dvRepo := &mockDocumentVersionRepo{}
	log := logger.New("dev", "info")

	fileID := uuid.New()
	fileRepo.files = append(fileRepo.files, &models.File{ID: fileID, ProjectID: uuid.New(), Name: "test"})

	dsm := service.NewDocumentStateManager(hub, fileRepo, dvRepo, log)

	update := []byte("fake-update")
	err := dsm.ApplyUpdate(context.Background(), fileID, uuid.New(), update)
	require.NoError(t, err)

	err = dsm.SaveSnapshot(context.Background(), fileID)
	require.NoError(t, err)

	file, err := fileRepo.FindByID(context.Background(), fileID)
	require.NoError(t, err)
	assert.Equal(t, update, file.Content)
	assert.Equal(t, "yjs-binary", file.ContentType)
}
