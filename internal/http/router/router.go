package router

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"mrchat/internal/app/config"
	"mrchat/internal/http/middleware"
	"mrchat/internal/modules/account"
	"mrchat/internal/modules/admin"
	authmodule "mrchat/internal/modules/auth"
	"mrchat/internal/modules/billing"
	"mrchat/internal/modules/catalog"
	"mrchat/internal/modules/chat"
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
	catalogHandler *catalog.Handler,
	chatHandler *chat.Handler,
	adminHandler *admin.Handler,
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
			authorized.GET("/models", catalogHandler.ListVisibleModels)
			authorized.GET("/conversations", chatHandler.ListConversations)
			authorized.POST("/conversations", chatHandler.CreateConversation)
			authorized.PUT("/conversations/:id", chatHandler.UpdateConversation)
			authorized.DELETE("/conversations/:id", chatHandler.DeleteConversation)
			authorized.GET("/conversations/:id/messages", chatHandler.ListMessages)

			authorized.GET("/users/me", userHandler.GetMe)
			authorized.PUT("/users/me", userHandler.UpdateMe)
			authorized.GET("/users/me/quota", userHandler.GetQuota)
			authorized.GET("/users/me/usage", userHandler.GetUsage)
			authorized.GET("/users/me/security", userHandler.GetSecurity)
			authorized.PUT("/users/me/password", userHandler.ChangePassword)

			authorized.GET("/billing/summary", billingHandler.GetSummary)
			authorized.GET("/billing/logs", billingHandler.ListLogs)
		}

		adminGroup := api.Group("/admin")
		adminGroup.Use(
			middleware.RequireAuth(tokenManager),
			middleware.RequireRoles(account.RoleAdmin, account.RoleRoot),
		)
		{
			adminGroup.GET("/upstreams", adminHandler.ListUpstreams)
			adminGroup.POST("/upstreams", adminHandler.CreateUpstream)
			adminGroup.PUT("/upstreams/:id", adminHandler.UpdateUpstream)

			adminGroup.GET("/channels", adminHandler.ListChannels)
			adminGroup.POST("/channels", adminHandler.CreateChannel)
			adminGroup.PUT("/channels/:id", adminHandler.UpdateChannel)

			adminGroup.GET("/models", adminHandler.ListModels)
			adminGroup.POST("/models", adminHandler.CreateModel)
			adminGroup.PUT("/models/:id", adminHandler.UpdateModel)

			adminGroup.GET("/user-groups", adminHandler.ListUserGroups)
			adminGroup.POST("/user-groups", adminHandler.CreateUserGroup)
			adminGroup.PUT("/user-groups/:id", adminHandler.UpdateUserGroup)
			adminGroup.GET("/user-groups/:id/limits", adminHandler.GetUserGroupLimitPolicies)
			adminGroup.PUT("/user-groups/:id/limits", adminHandler.UpdateUserGroupLimitPolicies)

			adminGroup.GET("/users", adminHandler.ListUsers)
			adminGroup.PUT("/users/:id/group", adminHandler.AssignUserGroup)
			adminGroup.PUT("/users/:id/quota", adminHandler.AdjustUserQuota)
			adminGroup.GET("/users/:id/limit-usage", adminHandler.GetUserLimitUsage)
			adminGroup.GET("/users/:id/limit-adjustments", adminHandler.ListUserLimitAdjustments)
			adminGroup.POST("/users/:id/limit-adjustments", adminHandler.CreateUserLimitAdjustment)

			adminGroup.GET("/audit-logs", adminHandler.ListAuditLogs)
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
