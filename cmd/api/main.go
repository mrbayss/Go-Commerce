package main

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/Go-Commerce/internal/config"
)

// main is the entry point of the application.
func main() {
	cfg := config.NewViper()
	log := config.NewLogger(cfg)
	validator := config.NewValidation()
	app := config.NewFiber(cfg, validator)
	db := config.NewDatabase(cfg, log)
	redis := config.NewRedis(cfg)

	config.Bootstrap(&config.BootStrapConfig{
		DB:       db,
		App:      app,
		Log:      log,
		Validate: validator,
		Config:   cfg,
		Redis:    redis,
	})

	app.Get("/health", func(c fiber.Ctx) error {
		return config.HealthCheck(c, db, redis)
	})

	// Start Fiber HTTP server
	if err := app.Listen(":" + cfg.GetString("app.port")); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
