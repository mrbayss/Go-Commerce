package app

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func HealthCheck(c fiber.Ctx, db *gorm.DB, rdb *redis.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	status := "ok"
	components := fiber.Map{}

	if sqlDB, err := db.DB(); err != nil {
		status = "degraded"
		components["database"] = "error"
	} else if err := sqlDB.PingContext(ctx); err != nil {
		status = "degraded"
		components["database"] = "error"
	} else {
		components["database"] = "ok"
	}

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
