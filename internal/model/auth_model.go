package model

import "time"

type SignUpReq struct {
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate:"required,min=8,max=72,password"`
	FirstName string `json:"first_name" validate:"required,notblank,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,notblank,min=1,max=100"`
	Phone     string `json:"phone" validate:"omitempty,id_phone"`
}

type SignUpRes struct {
	Email     string    `json:"email"`
	ExpiredAt time.Time `json:"expired_at"`
}
