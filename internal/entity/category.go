package entity

import (
	"github.com/google/uuid"
)

// Category represents a hierarchical product category.
// Categories support parent-child nesting via self-referencing ParentID.
type Category struct {
	BaseModel
	ParentID    *uuid.UUID `json:"parent_id,omitempty" gorm:"type:uuid;index"`
	Name        string     `json:"name" gorm:"type:varchar(100);not null"`
	Slug        string     `json:"slug" gorm:"type:varchar(100);uniqueIndex;not null"`
	Description *string    `json:"description,omitempty" gorm:"type:text"`
	ImageURL    *string    `json:"image_url,omitempty" gorm:"type:varchar(500)"`
	Level       int        `json:"level" gorm:"not null;default:0"`
	SortOrder   int        `json:"sort_order" gorm:"not null;default:0"`
	IsActive    bool       `json:"is_active" gorm:"not null;default:true"`

	// Relations
	Parent   *Category  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []Category `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Products []Product  `json:"products,omitempty" gorm:"foreignKey:CategoryID"`
}
