package database

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/pressly/goose/v3"

	"mrchat/internal/app/config"
)

func RunMigrations(ctx context.Context, client *Client, cfg config.PostgresConfig) error {
	if client == nil || !client.Enabled() || !cfg.AutoMigrate {
		return nil
	}

	sqlDB, err := client.DB.DB()
	if err != nil {
		return fmt.Errorf("get sql db for migrations: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	migrationsDir := filepath.Clean(cfg.MigrationsDir)
	if err := goose.UpContext(ctx, sqlDB, migrationsDir); err != nil {
		return fmt.Errorf("run goose migrations: %w", err)
	}

	return nil
}
