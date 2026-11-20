package entity

import "github.com/google/uuid"

type ProductStatus string

const (
	ProductStatusDraft    ProductStatus = "draft"
	ProductStatusPending  ProductStatus = "pending"
	ProductStatusActive   ProductStatus = "active"
	ProductStatusRejected ProductStatus = "rejected"
	ProductStatusArchived ProductStatus = "archived"
)

// Product belongs to a store and a category. MinPrice/MaxPrice are a
// server-computed cache of the cheapest/dearest active variant.
type Product struct {
	BaseModel
	StoreID     uuid.UUID     `gorm:"type:uuid;not null;index"`
	CategoryID  uuid.UUID     `gorm:"type:uuid;not null;index"`
	Name        string        `gorm:"type:varchar(255);not null"`
	Slug        string        `gorm:"type:varchar(255);not null;uniqueIndex:idx_products_slug,where:deleted_at IS NULL"`
	Description *string       `gorm:"type:text"`
	MinPrice    int64         `gorm:"not null;default:0"`
	MaxPrice    int64         `gorm:"not null;default:0"`
	Status      ProductStatus `gorm:"type:varchar(20);not null;default:'active';index"`
	IsActive    bool          `gorm:"not null;default:true"`
	TotalSold   int64         `gorm:"not null;default:0"`
	Metadata    *string       `gorm:"type:jsonb"`

	Store    Store            `gorm:"foreignKey:StoreID"`
	Category Category         `gorm:"foreignKey:CategoryID"`
	Variants []ProductVariant `gorm:"foreignKey:ProductID"`
	Images   []ProductImage   `gorm:"foreignKey:ProductID"`
}

// ProductVariant carries SKU, price and stock. StoreID is denormalized
// so the composite unique index (store_id, sku) can enforce SKU
// uniqueness per store.
type ProductVariant struct {
	BaseModel
	ProductID      uuid.UUID `gorm:"type:uuid;not null;index"`
	StoreID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_product_variants_store_sku,priority:1,where:deleted_at IS NULL"`
	SKU            string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_product_variants_store_sku,priority:2,where:deleted_at IS NULL"`
	Name           string    `gorm:"type:varchar(255);not null"`
	Price          int64     `gorm:"not null"`
	CompareAtPrice *int64
	Stock          int      `gorm:"not null;default:0"`
	Weight         *float64 `gorm:"type:decimal(10,2)"`
	IsActive       bool     `gorm:"not null;default:true"`
	Attributes     *string  `gorm:"type:jsonb"`

	Product Product `gorm:"foreignKey:ProductID"`
}

// ProductImage is a locally uploaded product photo.
// Exactly one image per product may have IsPrimary set.
type ProductImage struct {
	BaseModel
	ProductID uuid.UUID `gorm:"type:uuid;not null;index"`
	URL       string    `gorm:"type:varchar(500);not null"`
	AltText   *string   `gorm:"type:varchar(255)"`
	SortOrder int       `gorm:"not null;default:0"`
	IsPrimary bool      `gorm:"not null;default:false"`

	Product Product `gorm:"foreignKey:ProductID"`
}
