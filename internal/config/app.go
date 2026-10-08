package config

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// BootStrapConfig holds all application dependencies for dependency injection.
type BootStrapConfig struct {
	DB       *gorm.DB
	App      *fiber.App
	Log      *logrus.Logger
	Validate *FiberValidator
	Config   *viper.Viper
	Redis    *redis.Client
}

// Bootstrap initializes the application with the provided dependencies.
func Bootstrap(config *BootStrapConfig) {

}

// HealthCheck returns the health status of the application and its dependencies.
func HealthCheck(c fiber.Ctx, db *gorm.DB, rdb *redis.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	status := "ok"
	components := fiber.Map{}

	// Database check
	if sqlDB, err := db.DB(); err != nil {
		status = "degraded"
		components["database"] = "error"
	} else if err := sqlDB.PingContext(ctx); err != nil {
		status = "degraded"
		components["database"] = "error"
	} else {
		components["database"] = "ok"
	}

	// Redis check
	if err := rdb.Ping(ctx).Err(); err != nil {
		status = "degraded"
		components["redis"] = "error"
	} else {
		components["redis"] = "ok"
	}

	return c.JSON(fiber.Map{
		"status":     status,
		"components": components,
	})
}
