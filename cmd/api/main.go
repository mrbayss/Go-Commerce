package main

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/Go-Commerce/internal/config"
	"github.com/mrbayss/Go-Commerce/internal/pkg/mail"
	"github.com/mrbayss/Go-Commerce/internal/pkg/mail/templates"
)

// main is the entry point of the application.
func main() {
	cfg := config.NewViper()
	log := config.NewLogger(cfg)
	validator := config.NewValidation()
	app := config.NewFiber(cfg, validator, log)
	db := config.NewDatabase(cfg, log)
	redis := config.NewRedis(cfg)

	mailCfg, err := mail.NewClient(cfg, templates.FS)
	if err != nil {
		log.Fatalf("failed to connect email: %v", err)
	}

	config.Bootstrap(&config.BootStrapConfig{
		DB:       db,
		App:      app,
		Log:      log,
		Validate: validator,
		Config:   cfg,
		Redis:    redis,
		Mailer:   *mailCfg,
	})

	app.Get("/health", func(c fiber.Ctx) error {
		return config.HealthCheck(c, db, redis)
	})

	// Start Fiber HTTP server
	if err := app.Listen(":" + cfg.GetString("app.port")); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
