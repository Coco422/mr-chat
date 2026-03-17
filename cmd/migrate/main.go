package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"mrchat/internal/app/config"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	if !cfg.Postgres.Enabled {
		fmt.Fprintln(os.Stderr, "postgres is disabled")
		os.Exit(1)
	}

	db, err := sql.Open("pgx", cfg.Postgres.DSN())
	if err != nil {
		fmt.Fprintf(os.Stderr, "open postgres: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "ping postgres: %v\n", err)
		os.Exit(1)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		fmt.Fprintf(os.Stderr, "set goose dialect: %v\n", err)
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]
	if err := goose.RunContext(context.Background(), command, db, cfg.Postgres.MigrationsDir, args...); err != nil {
		fmt.Fprintf(os.Stderr, "run migration command %q: %v\n", command, err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage: go run ./cmd/migrate <status|up|down|redo|reset|version>")
}
