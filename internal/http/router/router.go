package router

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"mrchat/internal/app/config"
	"mrchat/internal/http/middleware"
	authmodule "mrchat/internal/modules/auth"
	"mrchat/internal/modules/billing"
	"mrchat/internal/modules/health"
	"mrchat/internal/modules/users"
	"mrchat/internal/shared/httpx"
)

func New(
	cfg config.Config,
	log *slog.Logger,
	healthHandler *health.Handler,
	authHandler *authmodule.Handler,
	userHandler *users.Handler,
	billingHandler *billing.Handler,
	tokenManager *authmodule.TokenManager,
) *gin.Engine {
	setMode(cfg.App.Environment)

	engine := gin.New()
	engine.Use(
		middleware.RequestID(),
		middleware.CORS(cfg.CORS),
		middleware.RequestLogger(log),
		middleware.Recovery(log),
	)

	engine.NoRoute(func(c *gin.Context) {
		httpx.Failure(c, http.StatusNotFound, "NOT_FOUND", "Route not found", nil)
	})

	engine.GET("/healthz", healthHandler.Get)

	api := engine.Group("/api/v1")
	{
		api.GET("/health", healthHandler.Get)
		api.POST("/auth/signup", authHandler.SignUp)
		api.POST("/auth/signin", authHandler.SignIn)
		api.POST("/auth/signout", authHandler.SignOut)
		api.POST("/auth/refresh", authHandler.Refresh)

		authorized := api.Group("")
		authorized.Use(middleware.RequireAuth(tokenManager))
		{
			authorized.GET("/users/me", userHandler.GetMe)
			authorized.PUT("/users/me", userHandler.UpdateMe)
			authorized.GET("/users/me/quota", userHandler.GetQuota)
			authorized.GET("/users/me/usage", userHandler.GetUsage)
			authorized.GET("/users/me/security", userHandler.GetSecurity)
			authorized.PUT("/users/me/password", userHandler.ChangePassword)

			authorized.GET("/billing/summary", billingHandler.GetSummary)
			authorized.GET("/billing/logs", billingHandler.ListLogs)
		}
	}

	return engine
}

func setMode(environment string) {
	switch environment {
	case "production":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
}
