package controller

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/Go-Commerce/internal/model"
	"github.com/mrbayss/Go-Commerce/internal/model/response"
	"github.com/mrbayss/Go-Commerce/internal/service"
)

type AuthController struct {
	AuthService service.AuthService
}

func NewAuthController(authService service.AuthService) *AuthController {
	return &AuthController{
		AuthService: authService,
	}
}

func (ac *AuthController) SignUp(c fiber.Ctx) error {
	var request model.SignUpReq
	if err := c.Bind().Body(&request); err != nil {
		return err // Let middleware handle bind errors
	}

	res, err := ac.AuthService.SignUp(c, request)
	if err != nil {
		return err // Let middleware handle all errors
	}

	return response.SuccessResponse(c, "success signup, verify otp send to your mail", res)
}

func (ac *AuthController) VerifyOTP(c fiber.Ctx) error {
	var request model.VerifyOTP
	if err := c.Bind().Body(&request); err != nil {
		return err
	}

	res, err := ac.AuthService.VerifyOTP(c, request)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, "OTP verified successfully", res)
}

func (ac *AuthController) ResendOTP(c fiber.Ctx) error {
	var request model.ResendOTPReq
	if err := c.Bind().Body(&request); err != nil {
		return err
	}

	res, err := ac.AuthService.ResendOTP(c, request)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, "OTP resent successfully", res)
}
