package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ilya/codestream/internal/models"
)

// RefreshTokenRepository provides data access for refresh tokens.
type RefreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new RefreshTokenRepository.
func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create inserts a new refresh token into the database.
func (r *RefreshTokenRepository) Create(ctx context.Context, token *models.RefreshToken) (*models.RefreshToken, error) {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return nil, err
	}

	return token, nil
}

// FindByHash finds a non-expired, non-revoked refresh token by its hash.
func (r *RefreshTokenRepository) FindByHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	now := time.Now()

	if err := r.db.WithContext(ctx).
		Where("token_hash = ? AND expires_at > ? AND revoked_at IS NULL", hash, now).
		First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &token, nil
}

// Revoke marks a refresh token as revoked.
func (r *RefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("id = ?", id).
		Update("revoked_at", now)

	if result.Error != nil {
		return result.Error
	}

	return nil
}
