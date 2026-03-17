package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"

	"mrchat/internal/app/config"
)

var (
	defaultAllowedHeaders = "Authorization, Content-Type, X-Request-ID"
	defaultAllowedMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	defaultExposeHeaders  = "X-Request-ID"
)

func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	allowedOrigins := cfg.AllowedOrigins

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}

		if len(allowedOrigins) > 0 && !slices.Contains(allowedOrigins, origin) {
			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}

			c.Next()
			return
		}

		headers := c.Writer.Header()
		headers.Set("Access-Control-Allow-Origin", origin)
		headers.Set("Access-Control-Allow-Headers", defaultAllowedHeaders)
		headers.Set("Access-Control-Allow-Methods", defaultAllowedMethods)
		headers.Set("Access-Control-Expose-Headers", defaultExposeHeaders)
		headers.Add("Vary", "Origin")
		headers.Add("Vary", "Access-Control-Request-Method")
		headers.Add("Vary", "Access-Control-Request-Headers")
		if cfg.AllowCredentials {
			headers.Set("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
