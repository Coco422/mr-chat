// @title MrChat API
// @version 0.1
// @description MrChat backend API for auth, user settings, chat completions, billing, and admin configuration. Swagger UI is exposed at /swagger/index.html.
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter access token as `Bearer <token>`.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"mrchat/internal/app/config"
	"mrchat/internal/app/server"
	"mrchat/internal/http/router"
	_ "mrchat/internal/http/swagger"
	"mrchat/internal/modules/account"
	"mrchat/internal/modules/admin"
	"mrchat/internal/modules/audit"
	authmodule "mrchat/internal/modules/auth"
	"mrchat/internal/modules/billing"
	"mrchat/internal/modules/catalog"
	"mrchat/internal/modules/chat"
	"mrchat/internal/modules/health"
	"mrchat/internal/modules/limits"
	"mrchat/internal/modules/users"
	"mrchat/internal/platform/cache"
	"mrchat/internal/platform/database"
	"mrchat/internal/platform/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	log := logger.New(cfg.App.Environment)

	dbClient, err := database.New(context.Background(), cfg.Postgres)
	if err != nil {
		log.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer dbClient.Close()
	if !dbClient.Enabled() {
		log.Error("postgres must be enabled for the api server")
		os.Exit(1)
	}
	if err := database.RunMigrations(context.Background(), dbClient, cfg.Postgres); err != nil {
		log.Error("failed to run database migrations", "error", err)
		os.Exit(1)
	}

	redisClient, err := cache.New(context.Background(), cfg.Redis)
	if err != nil {
		log.Error("failed to initialize redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	accountRepository := account.NewRepository(dbClient.DB)
	auditRepository := audit.NewRepository(dbClient.DB)
	catalogRepository := catalog.NewRepository(dbClient.DB)
	limitsRepository := limits.NewRepository(dbClient.DB)
	tokenManager := authmodule.NewTokenManager(cfg.Auth)
	authService := authmodule.NewService(accountRepository, tokenManager)
	userService := users.NewService(accountRepository)
	billingService := billing.NewService(accountRepository)
	catalogService := catalog.NewService(catalogRepository, accountRepository)
	limitsService := limits.NewService(limitsRepository, accountRepository)
	chatService := chat.NewService(chat.NewRepository(dbClient.DB), accountRepository, catalogRepository, limitsService)
	adminService := admin.NewService(accountRepository, catalogRepository, limitsService, auditRepository)

	healthHandler := health.NewHandler(cfg, dbClient, redisClient)
	authHandler := authmodule.NewHandler(cfg.Auth, authService)
	userHandler := users.NewHandler(userService)
	billingHandler := billing.NewHandler(billingService)
	catalogHandler := catalog.NewHandler(catalogService)
	chatHandler := chat.NewHandler(chatService)
	adminHandler := admin.NewHandler(adminService)

	engine := router.New(
		cfg,
		log,
		healthHandler,
		authHandler,
		userHandler,
		billingHandler,
		catalogHandler,
		chatHandler,
		adminHandler,
		tokenManager,
	)

	httpServer := server.New(cfg.HTTP, engine)

	go func() {
		log.Info("starting api server", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	waitForShutdown(log, httpServer, cfg.HTTP.ShutdownTimeout)
}

func waitForShutdown(log *slog.Logger, srv *http.Server, shutdownTimeout config.Duration) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	sig := <-stop
	log.Info("shutdown signal received", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout.Duration())
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
		return
	}

	log.Info("server shutdown complete")
}
