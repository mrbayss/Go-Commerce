package entity

import (
	"github.com/google/uuid"
)

// Category represents a hierarchical product category.
// Categories support parent-child nesting via self-referencing ParentID.
type Category struct {
	BaseModel
	ParentID    *uuid.UUID `gorm:"type:uuid;index"`
	Name        string     `gorm:"type:varchar(100);not null"`
	Slug        string     `gorm:"type:varchar(100);uniqueIndex;not null"`
	Description *string    `gorm:"type:text"`
	ImageURL    *string    `gorm:"type:varchar(500)"`
	Level       int        `gorm:"not null;default:0"`
	SortOrder   int        `gorm:"not null;default:0"`
	IsActive    bool       `gorm:"not null;default:true"`

	// Relations
	Parent   *Category  `gorm:"foreignKey:ParentID"`
	Children []Category `gorm:"foreignKey:ParentID"`
	Products []Product  `gorm:"foreignKey:CategoryID"`
}
