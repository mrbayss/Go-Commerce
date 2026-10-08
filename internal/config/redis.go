package config

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

// NewRedis creates a new Redis client and verifies the connection.
func NewRedis(viper *viper.Viper) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.address"),
		Password: viper.GetString("redis.pass"),
		DB:       viper.GetInt("redis.db"),
	})

	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("redis.dial_timeout"))
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		panic(fmt.Errorf("failed to connect redis: %v", err))
	}

	return rdb
}
