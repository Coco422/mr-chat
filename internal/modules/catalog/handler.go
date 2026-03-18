package catalog

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mrchat/internal/http/middleware"
	"mrchat/internal/modules/account"
	"mrchat/internal/shared/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ListVisibleModels godoc
// @Summary List visible models
// @Description Return models visible to current user according to user group visibility
// @Tags Catalog
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Router /models [get]
func (h *Handler) ListVisibleModels(c *gin.Context) {
	items, err := h.service.ListVisibleModels(c.Request.Context(), middleware.CurrentUserID(c))
	if err != nil {
		if err == account.ErrUserNotFound {
			httpx.Failure(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found", nil)
			return
		}
		httpx.Failure(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		return
	}

	httpx.Success(c, http.StatusOK, items)
}
