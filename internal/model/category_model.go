package model

import (
	"time"

	"github.com/google/uuid"
)

type CreateCategoryReq struct {
	Name        string     `json:"name" validate:"required,notblank,min=1,max=100"`
	ParentID    *uuid.UUID `json:"parent_id" validate:"omitempty,uuid"`
	Description *string    `json:"description" validate:"omitempty,max=5000"`
	ImageURL    *string    `json:"image_url" validate:"omitempty,url,max=500"`
}

// UpdateCategoryReq.ParentID is *string (not *uuid.UUID) to distinguish
// "not sent" (nil) from "re-root" (empty string), following UpdateProfileReq.Phone.
// No `uuid` tag: validator's omitempty does not skip a non-nil pointer to "",
// so the service parses the UUID manually ("" is a legitimate re-root request).
type UpdateCategoryReq struct {
	Name        *string `json:"name" validate:"omitempty,notblank,min=1,max=100"`
	ParentID    *string `json:"parent_id"`
	Description *string `json:"description" validate:"omitempty,max=5000"`
	IsActive    *bool   `json:"is_active"`
	SortOrder   *int    `json:"sort_order" validate:"omitempty,min=0"`
}

type CategoryRes struct {
	ID          uuid.UUID  `json:"id"`
	ParentID    *uuid.UUID `json:"parent_id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description *string    `json:"description"`
	ImageURL    *string    `json:"image_url"`
	Level       int        `json:"level"`
	SortOrder   int        `json:"sort_order"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
