package models

import (
	"time"

	"github.com/google/uuid"
)

// ChangeHistory represents an audit log entry.
type ChangeHistory struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FileID    *uuid.UUID `gorm:"type:uuid;index" json:"file_id,omitempty"`
	ProjectID *uuid.UUID `gorm:"type:uuid;index" json:"project_id,omitempty"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Action    string     `gorm:"not null" json:"action"`
	Metadata  string     `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// TableName returns the table name for the ChangeHistory model.
func (ChangeHistory) TableName() string {
	return "change_history"
}
