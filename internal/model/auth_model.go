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

type VerifyOTP struct {
	CodeOTP string `json:"code_otp" validate:"required,max=6"`
	Email   string `json:"email" validate:"required,email,max=255"`
}

type VerifyOTPRes struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type ResendOTPReq struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

type ResendOTPRes struct {
	Email     string    `json:"email"`
	ExpiredAt time.Time `json:"expired_at"`
}
