package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel contains common fields for all entities.
// All entities embed this struct to get UUID primary key, timestamps, and soft delete.
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
