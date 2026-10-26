package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/Go-Commerce/internal/app"
	"github.com/mrbayss/Go-Commerce/internal/config"
	"github.com/mrbayss/Go-Commerce/internal/controller"
	infraDb "github.com/mrbayss/Go-Commerce/internal/infra/db"
	infraMail "github.com/mrbayss/Go-Commerce/internal/infra/mail"
	infraRedis "github.com/mrbayss/Go-Commerce/internal/infra/redis"
	"github.com/mrbayss/Go-Commerce/internal/repository"
	"github.com/mrbayss/Go-Commerce/internal/router"
	"github.com/mrbayss/Go-Commerce/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log := config.NewLogger(cfg)
	validator := app.NewValidation()
	db := infraDb.NewDatabase(cfg, log)
	redisClient := infraRedis.NewRedis(cfg)
	mailer := infraMail.NewClient(&infraMail.SMTPConfig{
		Host:        cfg.SMTP.Host,
		Port:        cfg.SMTP.Port,
		Username:    cfg.SMTP.Username,
		Password:    cfg.SMTP.Password,
		FromAddress: cfg.SMTP.FromAddress,
		FromName:    cfg.SMTP.FromName,
		TLSMode:     cfg.SMTP.TLSMode,
		Timeout:     cfg.SMTP.Timeout,
	})

	// Domain wiring
	authRepo := repository.NewAuthRepository(db)
	authCfg := service.AuthConfig{
		OTPTtl:            cfg.Auth.OTPTtl,
		MaxSignupPerEmail: cfg.Auth.SignupRateLimit.Max,
		SignupRateWindow:  cfg.Auth.SignupRateLimit.Window,
		ResendCooldown:    cfg.Auth.ResendCooldown,
	}
	authService := service.NewAuthService(authRepo, log, redisClient, validator.Validator, mailer, authCfg)
	authController := controller.NewAuthController(authService)

	// HTTP server
	appInstance := app.NewFiber(cfg, log, validator)

	apiV1 := appInstance.Group("/api/v1")
	router.SetupAuthRoutes(apiV1, authController)

	appInstance.Get("/health", func(c fiber.Ctx) error {
		return app.HealthCheck(c, db, redisClient)
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
