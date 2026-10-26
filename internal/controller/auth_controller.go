package controller

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/Go-Commerce/internal/model"
	"github.com/mrbayss/Go-Commerce/internal/model/response"
	"github.com/mrbayss/Go-Commerce/internal/service"
)

type AuthController struct {
	AuthService *service.AuthService
}

func NewAuthController(authService *service.AuthService) *AuthController {
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
