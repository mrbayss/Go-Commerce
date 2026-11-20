package model

import (
	"time"

	"github.com/google/uuid"
)

type UpdateProfileReq struct {
	FirstName *string `json:"first_name" validate:"omitempty,notblank,min=1,max=100"`
	LastName  *string `json:"last_name" validate:"omitempty,notblank,min=1,max=100"`
	Phone     *string `json:"phone" validate:"omitempty,id_phone"`
}

type UserProfileRes struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     *string   `json:"phone"`
	AvatarURL *string   `json:"avatar_url"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
