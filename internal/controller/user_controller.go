package controller

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	apperrors "github.com/mrbayss/Go-Commerce/internal/domain"
	"github.com/mrbayss/Go-Commerce/internal/middleware"
	"github.com/mrbayss/Go-Commerce/internal/model/response"
	"github.com/mrbayss/Go-Commerce/internal/service"
)

type UserController struct {
	UserService service.UserService
}

func NewUserController(userService service.UserService) *UserController {
	return &UserController{
		UserService: userService,
	}
}

func (uc *UserController) GetMe(c fiber.Ctx) error {
	claims, ok := middleware.AuthClaimsFrom(c.Context())
	if !ok {
		return &apperrors.AppError{Code: http.StatusUnauthorized, Message: "unauthorized"}
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return &apperrors.AppError{Code: http.StatusUnauthorized, Message: "invalid user identity"}
	}

	res, err := uc.UserService.GetProfile(c.Context(), userID)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, "profile retrieved successfully", res)
}
