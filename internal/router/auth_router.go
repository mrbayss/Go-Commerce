package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/Go-Commerce/internal/controller"
)

func SetupAuthRoutes(api fiber.Router, ac *controller.AuthController, requireAuth fiber.Handler) {
	auth := api.Group("/auth")
	// Public auth flow.
	auth.Post("/signup", ac.SignUp)
	auth.Post("/verify-otp", ac.VerifyOTP)
	auth.Post("/resend-otp", ac.ResendOTP)
	auth.Post("/login", ac.Login)

	// Auth Middleware protected auth flow
	auth.Post("/refresh", requireAuth, ac.RefreshToken)
	auth.Post("/logout", requireAuth, ac.Logout)
}
