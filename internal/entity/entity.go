package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel contains common fields for all entities.
// All entities embed this struct to get UUID primary key, timestamps, and soft delete.
type BaseModel struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}
