package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilya/codestream/internal/handler"
	"github.com/ilya/codestream/internal/models"
	"github.com/ilya/codestream/internal/service"
)

type mockAuthService struct{}

func (m *mockAuthService) Register(ctx context.Context, email, password, displayName string) (*models.User, error) {
	return &models.User{ID: uuid.New(), Email: email, DisplayName: displayName}, nil
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (*service.Tokens, error) {
	return &service.Tokens{AccessToken: "access", RefreshToken: "refresh"}, nil
}

func (m *mockAuthService) Refresh(ctx context.Context, refreshToken string) (*service.Tokens, error) {
	return &service.Tokens{AccessToken: "new-access", RefreshToken: "new-refresh"}, nil
}

func (m *mockAuthService) Logout(ctx context.Context, refreshToken string) error {
	return nil
}

func TestRegisterHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authHandler := handler.NewAuthHandler(&mockAuthService{})
	router := gin.New()
	router.POST("/auth/register", authHandler.Register)

	body, _ := json.Marshal(map[string]string{
		"email":        "test@example.com",
		"password":     "password123",
		"display_name": "Test",
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "test@example.com")
}

func TestLoginHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authHandler := handler.NewAuthHandler(&mockAuthService{})
	router := gin.New()
	router.POST("/auth/login", authHandler.Login)

	body, _ := json.Marshal(map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "access_token")
}
