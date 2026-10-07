package models

import (
	"time"

	"github.com/google/uuid"
)

// Project represents a collaborative project.
type Project struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	OwnerID   uuid.UUID `gorm:"type:uuid;not null;index" json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Owner   User            `gorm:"foreignKey:OwnerID" json:"-"`
	Members []ProjectMember `gorm:"foreignKey:ProjectID" json:"-"`
	Files   []File          `gorm:"foreignKey:ProjectID" json:"-"`
}

// TableName returns the table name for the Project model.
func (Project) TableName() string {
	return "projects"
}
