package datastore

import (
	"context"
	"fmt"

	"github.com/mrbayss/Go-Commerce/infrastructure/config"
	"github.com/redis/go-redis/v9"
)

// NewRedis creates a new Redis client and verifies the connection.
func NewRedis(cfg *config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Pass,
		DB:       cfg.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Redis.DialTimeout)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		panic(fmt.Errorf("failed to connect redis: %v", err))
	}

	return rdb
}
