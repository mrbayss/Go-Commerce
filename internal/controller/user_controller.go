package controller

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/Go-Commerce/internal/middleware"
	"github.com/mrbayss/Go-Commerce/internal/model"
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
	userID, _, err := middleware.AuthIdentity(c)
	if err != nil {
		return err
	}

	res, err := uc.UserService.GetProfile(c.Context(), userID)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, "profile retrieved successfully", res)
}

func (uc *UserController) UpdateMe(c fiber.Ctx) error {
	var request model.UpdateProfileReq
	if err := c.Bind().Body(&request); err != nil {
		return err
	}

	userID, _, err := middleware.AuthIdentity(c)
	if err != nil {
		return err
	}

	res, err := uc.UserService.UpdateProfile(c.Context(), userID, request)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, "profile updated successfully", res)
}
