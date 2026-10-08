package models

import (
	"time"

	"github.com/google/uuid"
)

// File represents a file in a project.
type File struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID   uuid.UUID `gorm:"type:uuid;not null;index" json:"project_id"`
	Name        string    `gorm:"not null" json:"name"`
	Path        string    `gorm:"not null" json:"path"`
	Language    string    `json:"language"`
	Content      []byte  `gorm:"type:bytea" json:"-"`
	ContentText  *string `gorm:"type:text" json:"content_text,omitempty"`
	ContentType  string  `gorm:"column:content_type;default:'yjs-binary'" json:"content_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Project Project `gorm:"foreignKey:ProjectID" json:"-"`
}

// TableName returns the table name for the File model.
func (File) TableName() string {
	return "files"
}
