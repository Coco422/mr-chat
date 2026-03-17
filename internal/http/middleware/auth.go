package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"mrchat/internal/modules/account"
	"mrchat/internal/modules/auth"
	"mrchat/internal/shared/httpx"
)

const (
	currentUserIDKey   = "current_user_id"
	currentUserRoleKey = "current_user_role"
)

func RequireAuth(tokens *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			httpx.Failure(c, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Missing access token", nil)
			c.Abort()
			return
		}

		claims, err := tokens.ParseAccessToken(token)
		if err != nil {
			httpx.Failure(c, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Invalid access token", nil)
			c.Abort()
			return
		}

		c.Set(currentUserIDKey, claims.UserID)
		c.Set(currentUserRoleKey, string(claims.Role))
		c.Next()
	}
}

func RequireRoles(roles ...account.Role) gin.HandlerFunc {
	allowed := make(map[account.Role]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		role := CurrentUserRole(c)
		if _, ok := allowed[role]; !ok {
			httpx.Failure(c, http.StatusForbidden, "AUTH_FORBIDDEN", "You do not have permission to access this resource", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

func CurrentUserID(c *gin.Context) string {
	return c.GetString(currentUserIDKey)
}

func CurrentUserRole(c *gin.Context) account.Role {
	return account.Role(c.GetString(currentUserRoleKey))
}

func extractBearerToken(header string) string {
	if header == "" {
		return ""
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
