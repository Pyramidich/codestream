package models

import (
	"time"

	"github.com/google/uuid"
)

// DocumentVersion represents a stored CRDT update for a file.
type DocumentVersion struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FileID       uuid.UUID `gorm:"type:uuid;not null;index" json:"file_id"`
	UpdatePayload []byte   `gorm:"not null" json:"-"`
	CreatedBy    *uuid.UUID `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`

	File File `gorm:"foreignKey:FileID" json:"-"`
}

// TableName returns the table name for the DocumentVersion model.
func (DocumentVersion) TableName() string {
	return "document_versions"
}
