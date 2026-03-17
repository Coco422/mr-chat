package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"mrchat/internal/app/config"
)

type Client struct {
	Redis *redis.Client
}

func New(ctx context.Context, cfg config.RedisConfig) (*Client, error) {
	if !cfg.Enabled {
		return &Client{}, nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Address(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.DialTimeout.Duration(),
		ReadTimeout:  cfg.ReadTimeout.Duration(),
		WriteTimeout: cfg.WriteTimeout.Duration(),
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Client{Redis: client}, nil
}

func (c *Client) Close() error {
	if c == nil || c.Redis == nil {
		return nil
	}

	return c.Redis.Close()
}

func (c *Client) Enabled() bool {
	return c != nil && c.Redis != nil
}
