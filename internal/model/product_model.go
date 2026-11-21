package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type VariantReq struct {
	SKU            string          `json:"sku" validate:"required,notblank,min=1,max=100"`
	Name           string          `json:"name" validate:"required,notblank,min=1,max=255"`
	Price          int64           `json:"price" validate:"required,min=1"`
	CompareAtPrice *int64          `json:"compare_at_price" validate:"omitempty,min=1"`
	Stock          int             `json:"stock" validate:"min=0"`
	Weight         *float64        `json:"weight" validate:"omitempty,min=0"`
	Attributes     json.RawMessage `json:"attributes"`
}

type CreateProductReq struct {
	Name        string          `json:"name" validate:"required,notblank,min=1,max=255"`
	CategoryID  uuid.UUID       `json:"category_id" validate:"required,uuid"`
	Description *string         `json:"description" validate:"omitempty,max=10000"`
	Metadata    json.RawMessage `json:"metadata"`
	Variants    []VariantReq    `json:"variants" validate:"required,min=1,max=100,dive"`
}

type UpdateProductReq struct {
	Name        *string         `json:"name" validate:"omitempty,notblank,min=1,max=255"`
	CategoryID  *uuid.UUID      `json:"category_id" validate:"omitempty,uuid"`
	Description *string         `json:"description" validate:"omitempty,max=10000"`
	Metadata    json.RawMessage `json:"metadata"`
	IsActive    *bool           `json:"is_active"`
}

type UpdateVariantReq struct {
	SKU            *string         `json:"sku" validate:"omitempty,notblank,min=1,max=100"`
	Name           *string         `json:"name" validate:"omitempty,notblank,min=1,max=255"`
	Price          *int64          `json:"price" validate:"omitempty,min=1"`
	CompareAtPrice *int64          `json:"compare_at_price" validate:"omitempty,min=1"`
	Stock          *int            `json:"stock" validate:"omitempty,min=0"`
	Weight         *float64        `json:"weight" validate:"omitempty,min=0"`
	Attributes     json.RawMessage `json:"attributes"`
	IsActive       *bool           `json:"is_active"`
}

type VariantRes struct {
	ID             uuid.UUID       `json:"id"`
	ProductID      uuid.UUID       `json:"product_id"`
	SKU            string          `json:"sku"`
	Name           string          `json:"name"`
	Price          int64           `json:"price"`
	CompareAtPrice *int64          `json:"compare_at_price"`
	Stock          int             `json:"stock"`
	Weight         *float64        `json:"weight"`
	IsActive       bool            `json:"is_active"`
	Attributes     json.RawMessage `json:"attributes"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type ImageRes struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	URL       string    `json:"url"`
	AltText   *string   `json:"alt_text"`
	SortOrder int       `json:"sort_order"`
	IsPrimary bool      `json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProductRes struct {
	ID          uuid.UUID       `json:"id"`
	StoreID     uuid.UUID       `json:"store_id"`
	CategoryID  uuid.UUID       `json:"category_id"`
	Name        string          `json:"name"`
	Slug        string          `json:"slug"`
	Description *string         `json:"description"`
	MinPrice    int64           `json:"min_price"`
	MaxPrice    int64           `json:"max_price"`
	Status      string          `json:"status"`
	IsActive    bool            `json:"is_active"`
	TotalSold   int64           `json:"total_sold"`
	Metadata    json.RawMessage `json:"metadata"`
	Variants    []VariantRes    `json:"variants"`
	Images      []ImageRes      `json:"images"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
