package database

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func (cfg *Config) ConnectRedis() (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPwd,
		DB:       cfg.RedisDB, // 0 = default DB
	})

	// Test the connection
	ctx := context.Background()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	cfg.logger.Info("Connected to Redis successfully")

	return rdb, nil
}
