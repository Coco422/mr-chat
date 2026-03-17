package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"mrchat/internal/shared/httpx"
)

func Recovery(log *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Error(
			"panic recovered",
			"error", recovered,
			"request_id", c.GetString("request_id"),
			"path", c.FullPath(),
			"method", c.Request.Method,
		)

		httpx.Failure(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
	})
}
