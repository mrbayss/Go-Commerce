package entity

import (
	"github.com/google/uuid"
)

type ProductStatus string

const (
	ProductStatusDraft    ProductStatus = "draft"
	ProductStatusPending  ProductStatus = "pending"
	ProductStatusActive   ProductStatus = "active"
	ProductStatusRejected ProductStatus = "rejected"
	ProductStatusArchived ProductStatus = "archived"
)

// Product represents a sellable item in the marketplace.
// A product belongs to a store (seller) and a category, and has multiple variants.
type Product struct {
	BaseModel
	StoreID     uuid.UUID     `gorm:"type:uuid;not null;index"`
	CategoryID  uuid.UUID     `gorm:"type:uuid;not null;index"`
	Name        string        `gorm:"type:varchar(255);not null"`
	Slug        string        `gorm:"type:varchar(255);uniqueIndex;not null"`
	Description *string       `gorm:"type:text"`
	MinPrice    int64         `gorm:"not null;default:0"` // Lowest variant price (cache, in smallest currency unit)
	MaxPrice    int64         `gorm:"not null;default:0"` // Highest variant price (cache, in smallest currency unit)
	Status      ProductStatus `gorm:"type:varchar(20);not null;default:'draft';index"`
	IsActive    bool          `gorm:"not null;default:true"`
	TotalSold   int64         `gorm:"not null;default:0"` // Denormalized sold count
	Metadata    *string       `gorm:"type:jsonb"`         // Extensible JSON metadata

	// Relations
	Store    Store            `gorm:"foreignKey:StoreID"`
	Category Category         `gorm:"foreignKey:CategoryID"`
	Variants []ProductVariant `gorm:"foreignKey:ProductID"`
	Images   []ProductImage   `gorm:"foreignKey:ProductID"`
	Reviews  []Review         `gorm:"foreignKey:ProductID"`
}

// ProductVariant represents a specific purchasable SKU of a product.
// For example, "T-Shirt Biru - Size M" is a variant of "T-Shirt Biru".
// Prices are stored as int64 in the smallest currency unit (e.g., 1 = Rp 1).
type ProductVariant struct {
	BaseModel
	ProductID      uuid.UUID `gorm:"type:uuid;not null;index"`
	SKU            string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	Name           string    `gorm:"type:varchar(255);not null"` // e.g., "Biru / M"
	Price          int64     `gorm:"not null"`                   // In smallest currency unit
	CompareAtPrice *int64    // Original price before discount
	Stock          int       `gorm:"not null;default:0"`
	Weight         *float64  `gorm:"type:decimal(10,2)"` // Weight in grams
	IsActive       bool      `gorm:"not null;default:true"`
	Attributes     *string   `gorm:"type:jsonb"` // e.g., {"color":"Biru","size":"M"}

	// Relations
	Product Product `gorm:"foreignKey:ProductID"`
}

// ProductImage represents an image attached to a product.
// A product can have multiple images with one marked as primary.
type ProductImage struct {
	BaseModel
	ProductID uuid.UUID `gorm:"type:uuid;not null;index"`
	URL       string    `gorm:"type:varchar(500);not null"`
	AltText   *string   `gorm:"type:varchar(255)"`
	SortOrder int       `gorm:"not null;default:0"`
	IsPrimary bool      `gorm:"not null;default:false"`

	// Relations
	Product Product `gorm:"foreignKey:ProductID"`
}
