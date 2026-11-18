package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/Go-Commerce/internal/controller"
)

func SetupRoutes(
	app *fiber.App,
	authController *controller.AuthController,
	userController *controller.UserController,
	requireAuth fiber.Handler,
) {
	apiV1 := app.Group("/api/v1")
	SetupAuthRoutes(apiV1, authController, requireAuth)
	SetupUserRoutes(apiV1, userController, requireAuth)
}
