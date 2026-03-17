package billing

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"mrchat/internal/http/middleware"
	"mrchat/internal/shared/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetSummary(c *gin.Context) {
	summary, err := h.service.GetSummary(c.Request.Context(), middleware.CurrentUserID(c))
	if err != nil {
		h.handleBillingError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, summary)
}

func (h *Handler) ListLogs(c *gin.Context) {
	page := parsePositiveInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parsePositiveInt(c.DefaultQuery("page_size", "20"), 20)

	logs, err := h.service.ListLogs(
		c.Request.Context(),
		middleware.CurrentUserID(c),
		page,
		pageSize,
		c.DefaultQuery("type", ""),
	)
	if err != nil {
		h.handleBillingError(c, err)
		return
	}

	httpx.SuccessWithMeta(c, http.StatusOK, logs.Items, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     logs.Total,
	})
}

func (h *Handler) handleBillingError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidBillingLogType):
		httpx.Failure(c, http.StatusBadRequest, "BILLING_INVALID_LOG_TYPE", "type must be one of consume, refund, redeem, admin_adjust", nil)
	default:
		httpx.Failure(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
	}
}

func parsePositiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	if parsed > 100 {
		return 100
	}

	return parsed
}
