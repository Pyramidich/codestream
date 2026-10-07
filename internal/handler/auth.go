package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ilya/codestream/internal/models"
	"github.com/ilya/codestream/internal/service"
)

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	authService AuthService
}

// AuthService defines the authentication service interface.
type AuthService interface {
	Register(ctx context.Context, email, password, displayName string) (*models.User, error)
	Login(ctx context.Context, email, password string) (*service.Tokens, error)
	Refresh(ctx context.Context, refreshToken string) (*service.Tokens, error)
	Logout(ctx context.Context, refreshToken string) error
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterRequest represents a registration request.
type RegisterRequest struct {
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	DisplayName string `json:"display_name"`
}

// LoginRequest represents a login request.
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest represents a token refresh request.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LogoutRequest represents a logout request.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Register handles user registration.
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req.Email, req.Password, req.DisplayName)
	if err != nil {
		status := http.StatusUnprocessableEntity
		if err.Error() == "user with this email already exists" {
			status = http.StatusConflict
		}

		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":           user.ID,
		"email":        user.Email,
		"display_name": user.DisplayName,
		"created_at":   user.CreatedAt,
	})
}

// Login handles user login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	tokens, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// Refresh handles token refresh.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	tokens, err := h.authService.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// Logout handles user logout.
func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
