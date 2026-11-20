package model

import (
	"time"

	"github.com/google/uuid"
)

type CreateStoreReq struct {
	Name        string  `json:"name" validate:"required,notblank,min=1,max=150"`
	Description *string `json:"description" validate:"omitempty,max=5000"`
}

type UpdateStoreReq struct {
	Name        *string `json:"name" validate:"omitempty,notblank,min=1,max=150"`
	Description *string `json:"description" validate:"omitempty,max=5000"`
}

type StoreRes struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Description    *string   `json:"description"`
	LogoURL        *string   `json:"logo_url"`
	BannerURL      *string   `json:"banner_url"`
	Status         string    `json:"status"`
	CommissionRate float64   `json:"commission_rate"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
