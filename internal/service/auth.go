package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/ilya/codestream/internal/models"
)

const (
	bcryptCost       = 10
	refreshTokenSize = 32
)

// Tokens holds access and refresh tokens.
type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// UserRepository defines the user repository interface used by auth service.
type UserRepository interface {
	Create(ctx context.Context, email, passwordHash, displayName string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

// RefreshTokenRepository defines the refresh token repository interface used by auth service.
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *models.RefreshToken) (*models.RefreshToken, error)
	FindByHash(ctx context.Context, hash string) (*models.RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
}

// AuthService provides authentication and authorization operations.
type AuthService struct {
	userRepo         UserRepository
	refreshTokenRepo RefreshTokenRepository
	jwtSecret        string
	accessTTL        time.Duration
	refreshTTL       time.Duration
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo UserRepository, refreshTokenRepo RefreshTokenRepository, jwtSecret string, accessTTL, refreshTTL time.Duration) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtSecret:        jwtSecret,
		accessTTL:        accessTTL,
		refreshTTL:       refreshTTL,
	}
}

// Register creates a new user account.
func (s *AuthService) Register(ctx context.Context, email, password, displayName string) (*models.User, error) {
	if err := validateEmail(email); err != nil {
		return nil, err
	}

	if err := validatePassword(password); err != nil {
		return nil, err
	}

	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, errors.New("user with this email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	return s.userRepo.Create(ctx, email, string(hashedPassword), displayName)
}

// Login authenticates a user and returns tokens.
func (s *AuthService) Login(ctx context.Context, email, password string) (*Tokens, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return s.generateTokenPair(ctx, user.ID)
}

// Refresh validates a refresh token and returns a new token pair.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*Tokens, error) {
	token, err := s.refreshTokenRepo.FindByHash(ctx, HashToken(refreshToken))
	if err != nil {
		return nil, err
	}

	if token == nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	if err := s.refreshTokenRepo.Revoke(ctx, token.ID); err != nil {
		return nil, err
	}

	return s.generateTokenPair(ctx, token.UserID)
}

// Logout revokes a refresh token.
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	token, err := s.refreshTokenRepo.FindByHash(ctx, HashToken(refreshToken))
	if err != nil {
		return err
	}

	if token == nil {
		return errors.New("invalid or expired refresh token")
	}

	return s.refreshTokenRepo.Revoke(ctx, token.ID)
}

// GenerateAccessToken generates a new access token for the given user ID.
func (s *AuthService) GenerateAccessToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"iat": now.Unix(),
		"exp": now.Add(s.accessTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// HashToken returns SHA-256 hash of the given token.
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (s *AuthService) generateTokenPair(ctx context.Context, userID uuid.UUID) (*Tokens, error) {
	accessToken, err := s.GenerateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateRandomToken(refreshTokenSize)
	if err != nil {
		return nil, err
	}

	refreshHash := HashToken(refreshToken)
	expiresAt := time.Now().Add(s.refreshTTL)

	token := &models.RefreshToken{
		UserID:    userID,
		TokenHash: refreshHash,
		ExpiresAt: expiresAt,
	}

	if _, err := s.refreshTokenRepo.Create(ctx, token); err != nil {
		return nil, err
	}

	return &Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func generateRandomToken(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

func validateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}

	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	return nil
}
