package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilya/codestream/internal/models"
	"github.com/ilya/codestream/internal/service"
)

type mockUserRepo struct {
	users []*models.User
}

func (m *mockUserRepo) Create(ctx context.Context, email, passwordHash, displayName string) (*models.User, error) {
	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		DisplayName:  displayName,
	}
	m.users = append(m.users, user)
	return user, nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

type mockRefreshTokenRepo struct {
	tokens []*models.RefreshToken
}

func (m *mockRefreshTokenRepo) Create(ctx context.Context, token *models.RefreshToken) (*models.RefreshToken, error) {
	token.ID = uuid.New()
	m.tokens = append(m.tokens, token)
	return token, nil
}

func (m *mockRefreshTokenRepo) FindByHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	for _, t := range m.tokens {
		if t.TokenHash == hash && t.RevokedAt == nil && t.ExpiresAt.After(time.Now()) {
			return t, nil
		}
	}
	return nil, nil
}

func (m *mockRefreshTokenRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	for _, t := range m.tokens {
		if t.ID == id {
			now := time.Now()
			t.RevokedAt = &now
			return nil
		}
	}
	return nil
}

func newTestAuthService() (*service.AuthService, *mockUserRepo, *mockRefreshTokenRepo) {
	userRepo := &mockUserRepo{}
	refreshRepo := &mockRefreshTokenRepo{}
	authService := service.NewAuthService(userRepo, refreshRepo, "secret", 15*time.Minute, 7*24*time.Hour)
	return authService, userRepo, refreshRepo
}

func TestRegister(t *testing.T) {
	authService, _, _ := newTestAuthService()

	user, err := authService.Register(context.Background(), "test@example.com", "password123", "Test")
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "Test", user.DisplayName)

	_, err = authService.Register(context.Background(), "test@example.com", "password123", "Test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestLogin(t *testing.T) {
	authService, _, _ := newTestAuthService()

	_, err := authService.Register(context.Background(), "test@example.com", "password123", "Test")
	require.NoError(t, err)

	tokens, err := authService.Login(context.Background(), "test@example.com", "password123")
	require.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)

	_, err = authService.Login(context.Background(), "test@example.com", "wrongpassword")
	assert.Error(t, err)

	_, err = authService.Login(context.Background(), "missing@example.com", "password123")
	assert.Error(t, err)
}

func TestRefresh(t *testing.T) {
	authService, _, _ := newTestAuthService()

	_, err := authService.Register(context.Background(), "test@example.com", "password123", "Test")
	require.NoError(t, err)

	tokens, err := authService.Login(context.Background(), "test@example.com", "password123")
	require.NoError(t, err)

	newTokens, err := authService.Refresh(context.Background(), tokens.RefreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, newTokens.AccessToken)
	assert.NotEmpty(t, newTokens.RefreshToken)
	assert.NotEqual(t, tokens.RefreshToken, newTokens.RefreshToken)
}

func TestLogout(t *testing.T) {
	authService, _, _ := newTestAuthService()

	_, err := authService.Register(context.Background(), "test@example.com", "password123", "Test")
	require.NoError(t, err)

	tokens, err := authService.Login(context.Background(), "test@example.com", "password123")
	require.NoError(t, err)

	err = authService.Logout(context.Background(), tokens.RefreshToken)
	require.NoError(t, err)

	_, err = authService.Refresh(context.Background(), tokens.RefreshToken)
	assert.Error(t, err)
}
