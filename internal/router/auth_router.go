package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/Go-Commerce/internal/controller"
)

func SetupAuthRoutes(api fiber.Router, ac *controller.AuthController) {
	auth := api.Group("/auth")
	auth.Post("/signup", ac.SignUp)
	auth.Post("/verify-otp", ac.VerifyOTP)
	auth.Post("/resend-otp", ac.ResendOTP)
}
