package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/Go-Commerce/internal/controller"
)

func SetupRoutes(app *fiber.App, authController *controller.AuthController) {
	apiV1 := app.Group("/api/v1")
	SetupAuthRoutes(apiV1, authController)
}