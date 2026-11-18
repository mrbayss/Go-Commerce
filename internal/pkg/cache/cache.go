package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrNotFound = errors.New("cache: key not found")

func Set(ctx context.Context, rdb *redis.Client, key string, value any, ttl time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	if err := rdb.Set(ctx, key, bytes, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache key %q: %w", key, err)
	}

	return nil
}

func Get[T any](ctx context.Context, rdb *redis.Client, key string) (*T, error) {
	data, err := rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cache key %q: %w", key, err)
	}

	var dest T
	if err := json.Unmarshal(data, &dest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache key %q: %w", key, err)
	}

	return &dest, nil
}

func Del(ctx context.Context, rdb *redis.Client, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	if err := rdb.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete cache keys: %w", err)
	}

	return nil
}
