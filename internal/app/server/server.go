package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mrchat/internal/app/config"
)

func New(cfg config.HTTPConfig, engine *gin.Engine) *http.Server {
	return &http.Server{
		Addr:         cfg.Address(),
		Handler:      engine,
		ReadTimeout:  cfg.ReadTimeout.Duration(),
		WriteTimeout: cfg.WriteTimeout.Duration(),
	}
}
