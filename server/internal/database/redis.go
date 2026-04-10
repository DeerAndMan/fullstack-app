package database

import (
	"context"
	"fmt"
	"log/slog"

	"fullstack-app/server/internal/config"

	"github.com/redis/go-redis/v9"
)

func NewRedis(cfg *config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("connect redis: %w", err)
	}

	slog.Info("redis connected", "addr", cfg.Addr())
	return rdb, nil
}
