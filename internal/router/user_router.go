package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/Go-Commerce/internal/controller"
)

func SetupUserRoutes(api fiber.Router, uc *controller.UserController, requireAuth fiber.Handler) {
	users := api.Group("/users")
	users.Get("/me", requireAuth, uc.GetMe)
}
