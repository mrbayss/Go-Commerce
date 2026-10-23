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
	StoreID     uuid.UUID     `json:"store_id" gorm:"type:uuid;not null;index"`
	CategoryID  uuid.UUID     `json:"category_id" gorm:"type:uuid;not null;index"`
	Name        string        `json:"name" gorm:"type:varchar(255);not null"`
	Slug        string        `json:"slug" gorm:"type:varchar(255);uniqueIndex;not null"`
	Description *string       `json:"description,omitempty" gorm:"type:text"`
	MinPrice    int64         `json:"min_price" gorm:"not null;default:0"` // Lowest variant price (cache, in smallest currency unit)
	MaxPrice    int64         `json:"max_price" gorm:"not null;default:0"` // Highest variant price (cache, in smallest currency unit)
	Status      ProductStatus `json:"status" gorm:"type:varchar(20);not null;default:'draft';index"`
	IsActive    bool          `json:"is_active" gorm:"not null;default:true"`
	TotalSold   int64         `json:"total_sold" gorm:"not null;default:0"` // Denormalized sold count
	Metadata    *string       `json:"metadata,omitempty" gorm:"type:jsonb"` // Extensible JSON metadata

	// Relations
	Store    Store            `json:"store,omitempty" gorm:"foreignKey:StoreID"`
	Category Category         `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Variants []ProductVariant `json:"variants,omitempty" gorm:"foreignKey:ProductID"`
	Images   []ProductImage   `json:"images,omitempty" gorm:"foreignKey:ProductID"`
	Reviews  []Review         `json:"reviews,omitempty" gorm:"foreignKey:ProductID"`
}

// ProductVariant represents a specific purchasable SKU of a product.
// For example, "T-Shirt Biru - Size M" is a variant of "T-Shirt Biru".
// Prices are stored as int64 in the smallest currency unit (e.g., 1 = Rp 1).
type ProductVariant struct {
	BaseModel
	ProductID      uuid.UUID `json:"product_id" gorm:"type:uuid;not null;index"`
	SKU            string    `json:"sku" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name           string    `json:"name" gorm:"type:varchar(255);not null"` // e.g., "Biru / M"
	Price          int64     `json:"price" gorm:"not null"`                  // In smallest currency unit
	CompareAtPrice *int64    `json:"compare_at_price,omitempty" gorm:""`     // Original price before discount
	Stock          int       `json:"stock" gorm:"not null;default:0"`
	Weight         *float64  `json:"weight,omitempty" gorm:"type:decimal(10,2)"` // Weight in grams
	IsActive       bool      `json:"is_active" gorm:"not null;default:true"`
	Attributes     *string   `json:"attributes,omitempty" gorm:"type:jsonb"` // e.g., {"color":"Biru","size":"M"}

	// Relations
	Product Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
}

// ProductImage represents an image attached to a product.
// A product can have multiple images with one marked as primary.
type ProductImage struct {
	BaseModel
	ProductID uuid.UUID `json:"product_id" gorm:"type:uuid;not null;index"`
	URL       string    `json:"url" gorm:"type:varchar(500);not null"`
	AltText   *string   `json:"alt_text,omitempty" gorm:"type:varchar(255)"`
	SortOrder int       `json:"sort_order" gorm:"not null;default:0"`
	IsPrimary bool      `json:"is_primary" gorm:"not null;default:false"`

	// Relations
	Product Product `json:"-" gorm:"foreignKey:ProductID"`
}
