package entity

import (
	"github.com/google/uuid"
)

// Cart represents a user's persistent shopping cart.
// Each user has at most one active cart.
type Cart struct {
	BaseModel
	UserID uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`

	// Relations
	User  User       `gorm:"foreignKey:UserID"`
	Items []CartItem `gorm:"foreignKey:CartID"`
}

// CartItem represents a single item in a user's cart.
// Links a product variant with a quantity.
type CartItem struct {
	BaseModel
	CartID    uuid.UUID `gorm:"type:uuid;not null;index"`
	VariantID uuid.UUID `gorm:"type:uuid;not null;index"`
	Quantity  int       `gorm:"not null;default:1"`

	// Relations
	Cart    Cart           `gorm:"foreignKey:CartID"`
	Variant ProductVariant `gorm:"foreignKey:VariantID"`
}
