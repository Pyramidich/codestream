package models

import (
	"time"

	"github.com/google/uuid"
)

// ProjectMember represents a user's membership in a project.
type ProjectMember struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null;index" json:"project_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Role      string    `gorm:"not null" json:"role"`
	JoinedAt  time.Time `json:"joined_at"`

	Project Project `gorm:"foreignKey:ProjectID" json:"-"`
	User    User    `gorm:"foreignKey:UserID" json:"-"`
}

// TableName returns the table name for the ProjectMember model.
func (ProjectMember) TableName() string {
	return "project_members"
}
