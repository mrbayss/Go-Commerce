package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/Go-Commerce/internal/controller"
)

// WithHandlerName wraps a handler and sets its name in context for logging.
func WithHandlerName(name string, handler fiber.Handler) fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Locals("handler", name)
		return handler(c)
	}
}

func SetupAuthRoutes(api fiber.Router, ac *controller.AuthController) {
	auth := api.Group("/auth")
	auth.Post("/signup", WithHandlerName("AuthController.SignUp", ac.SignUp))
}
