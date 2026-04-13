package database

import (
	"context"
	"fmt"

	"fullstack-app/server/internal/config"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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

	zap.L().Info("redis connected", zap.String("addr", cfg.Addr()))
	return rdb, nil
}
