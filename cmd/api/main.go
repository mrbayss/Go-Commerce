package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/Go-Commerce/infrastructure/app"
	"github.com/mrbayss/Go-Commerce/infrastructure/config"
	"github.com/mrbayss/Go-Commerce/infrastructure/datastore"
	"github.com/mrbayss/Go-Commerce/infrastructure/logger"
	"github.com/mrbayss/Go-Commerce/infrastructure/pkg/mailer"
	"github.com/mrbayss/Go-Commerce/infrastructure/validation"
	"github.com/mrbayss/Go-Commerce/internal/controller"
	"github.com/mrbayss/Go-Commerce/internal/repository"
	"github.com/mrbayss/Go-Commerce/internal/router"
	"github.com/mrbayss/Go-Commerce/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log := logger.NewLogger(cfg)
	validator := validation.NewValidation()
	database := datastore.NewDatabase(cfg, log)
	redisClient := datastore.NewRedis(cfg)
	mailer := mailer.NewClient(&mailer.SMTPConfig{
		Host:        cfg.SMTP.Host,
		Port:        cfg.SMTP.Port,
		Username:    cfg.SMTP.Username,
		Password:    cfg.SMTP.Password,
		FromAddress: cfg.SMTP.FromAddress,
		FromName:    cfg.SMTP.FromName,
		TLSMode:     cfg.SMTP.TLSMode,
		Timeout:     cfg.SMTP.Timeout,
	})

	// repository
	authRepo := repository.NewAuthRepository(database)

	// service
	authService := service.NewAuthService(
		authRepo,
		log,
		redisClient,
		validator.Validator,
		mailer,
		cfg.Auth.OTPTtl,
		cfg.Auth.SignupRateLimit.Max,
		cfg.Auth.SignupRateLimit.Window,
		cfg.Auth.ResendCooldown,
	)

	// controller
	authController := controller.NewAuthController(authService)

	// HTTP server
	appInstance := app.NewFiber(cfg, log, validator)

	apiV1 := appInstance.Group("/api/v1")
	router.SetupAuthRoutes(apiV1, authController)

	appInstance.Get("/health", func(c fiber.Ctx) error {
		return app.HealthCheck(c, database, redisClient)
	})

	// Start server in a goroutine so we can wait for signals
	go func() {
		if err := appInstance.Listen(":" + cfg.App.Port); err != nil {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = appInstance.ShutdownWithContext(ctx)
}
