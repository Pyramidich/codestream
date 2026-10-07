package models

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a refresh token stored in the database.
type RefreshToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenHash string     `gorm:"uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time  `gorm:"not null;index" json:"-"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"-"`
	UserAgent string     `json:"-"`
	IP        string     `json:"-"`

	// User is the owner of the refresh token.
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// TableName returns the table name for the RefreshToken model.
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
