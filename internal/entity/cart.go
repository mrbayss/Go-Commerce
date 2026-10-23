package entity

import (
	"github.com/google/uuid"
)

// Cart represents a user's persistent shopping cart.
// Each user has at most one active cart.
type Cart struct {
	BaseModel
	UserID uuid.UUID `json:"user_id" gorm:"type:uuid;uniqueIndex;not null"`

	// Relations
	User  User       `json:"-" gorm:"foreignKey:UserID"`
	Items []CartItem `json:"items,omitempty" gorm:"foreignKey:CartID"`
}

// CartItem represents a single item in a user's cart.
// Links a product variant with a quantity.
type CartItem struct {
	BaseModel
	CartID    uuid.UUID `json:"cart_id" gorm:"type:uuid;not null;index"`
	VariantID uuid.UUID `json:"variant_id" gorm:"type:uuid;not null;index"`
	Quantity  int       `json:"quantity" gorm:"not null;default:1"`

	// Relations
	Cart    Cart           `json:"-" gorm:"foreignKey:CartID"`
	Variant ProductVariant `json:"variant,omitempty" gorm:"foreignKey:VariantID"`
}
