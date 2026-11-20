package entity

import "github.com/google/uuid"

// Category is a global taxonomy node. Managed by admins only,
// used by sellers when creating products. Hierarchical via ParentID.
type Category struct {
	BaseModel
	ParentID    *uuid.UUID `gorm:"type:uuid;index"`
	Name        string     `gorm:"type:varchar(100);not null"`
	Slug        string     `gorm:"type:varchar(100);not null;uniqueIndex:idx_categories_slug,where:deleted_at IS NULL"`
	Description *string    `gorm:"type:text"`
	ImageURL    *string    `gorm:"type:varchar(500)"`
	Level       int        `gorm:"not null;default:0"`
	SortOrder   int        `gorm:"not null;default:0"`
	IsActive    bool       `gorm:"not null;default:true"`

	Parent   *Category  `gorm:"foreignKey:ParentID"`
	Children []Category `gorm:"foreignKey:ParentID"`
	Products []Product  `gorm:"foreignKey:CategoryID"`
}
