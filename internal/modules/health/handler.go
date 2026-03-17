package health

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mrchat/internal/app/config"
	"mrchat/internal/platform/cache"
	"mrchat/internal/platform/database"
	"mrchat/internal/shared/httpx"
)

type Handler struct {
	cfg   config.Config
	db    *database.Client
	cache *cache.Client
}

func NewHandler(cfg config.Config, db *database.Client, cache *cache.Client) *Handler {
	return &Handler{
		cfg:   cfg,
		db:    db,
		cache: cache,
	}
}

func (h *Handler) Get(c *gin.Context) {
	httpx.Success(c, http.StatusOK, gin.H{
		"status":      "ok",
		"service":     h.cfg.App.Name,
		"version":     h.cfg.App.Version,
		"environment": h.cfg.App.Environment,
		"postgres": gin.H{
			"enabled": h.db.Enabled(),
		},
		"redis": gin.H{
			"enabled": h.cache.Enabled(),
		},
	})
}
