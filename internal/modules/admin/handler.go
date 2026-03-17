package admin

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"mrchat/internal/http/middleware"
	"mrchat/internal/modules/account"
	"mrchat/internal/modules/audit"
	"mrchat/internal/modules/catalog"
	"mrchat/internal/shared/httpx"
)

type Handler struct {
	service *Service
}

type upstreamRequest struct {
	Name             string         `json:"name"`
	ProviderType     string         `json:"provider_type"`
	BaseURL          string         `json:"base_url"`
	AuthType         string         `json:"auth_type"`
	AuthConfig       map[string]any `json:"auth_config"`
	Status           string         `json:"status"`
	TimeoutSeconds   int            `json:"timeout_seconds"`
	CooldownSeconds  int            `json:"cooldown_seconds"`
	FailureThreshold int            `json:"failure_threshold"`
	Metadata         map[string]any `json:"metadata"`
}

type modelRequest struct {
	ModelKey        string                      `json:"model_key"`
	DisplayName     string                      `json:"display_name"`
	ProviderType    string                      `json:"provider_type"`
	ContextLength   int                         `json:"context_length"`
	MaxOutputTokens *int                        `json:"max_output_tokens"`
	Pricing         map[string]any              `json:"pricing"`
	Capabilities    map[string]any              `json:"capabilities"`
	AllowedGroupIDs []string                    `json:"allowed_group_ids"`
	Status          string                      `json:"status"`
	Metadata        map[string]any              `json:"metadata"`
	RouteBindings   []catalog.RouteBindingInput `json:"route_bindings"`
}

type quotaAdjustmentRequest struct {
	Delta  int64  `json:"delta"`
	Reason string `json:"reason"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListUpstreams(c *gin.Context) {
	items, err := h.service.ListUpstreams(c.Request.Context())
	if err != nil {
		h.internalError(c)
		return
	}
	httpx.Success(c, http.StatusOK, toUpstreams(items))
}

func (h *Handler) CreateUpstream(c *gin.Context) {
	var req upstreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid upstream payload", gin.H{"error": err.Error()})
		return
	}

	item, err := h.service.CreateUpstream(c.Request.Context(), actorFromContext(c), CreateUpstreamInput(req))
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusCreated, toUpstream(item))
}

func (h *Handler) UpdateUpstream(c *gin.Context) {
	var req upstreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid upstream payload", gin.H{"error": err.Error()})
		return
	}

	item, err := h.service.UpdateUpstream(c.Request.Context(), actorFromContext(c), c.Param("id"), UpdateUpstreamInput{
		Name:             &req.Name,
		ProviderType:     &req.ProviderType,
		BaseURL:          &req.BaseURL,
		AuthType:         &req.AuthType,
		AuthConfig:       req.AuthConfig,
		Status:           &req.Status,
		TimeoutSeconds:   &req.TimeoutSeconds,
		CooldownSeconds:  &req.CooldownSeconds,
		FailureThreshold: &req.FailureThreshold,
		Metadata:         req.Metadata,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, toUpstream(item))
}

func (h *Handler) ListModels(c *gin.Context) {
	items, err := h.service.ListModels(c.Request.Context())
	if err != nil {
		h.internalError(c)
		return
	}
	httpx.Success(c, http.StatusOK, toAdminModels(items))
}

func (h *Handler) CreateModel(c *gin.Context) {
	var req modelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid model payload", gin.H{"error": err.Error()})
		return
	}

	item, err := h.service.CreateModel(c.Request.Context(), actorFromContext(c), CreateModelInput(req))
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusCreated, toAdminModel(*item))
}

func (h *Handler) UpdateModel(c *gin.Context) {
	var req modelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid model payload", gin.H{"error": err.Error()})
		return
	}

	item, err := h.service.UpdateModel(c.Request.Context(), actorFromContext(c), c.Param("id"), UpdateModelInput{
		ModelKey:        &req.ModelKey,
		DisplayName:     &req.DisplayName,
		ProviderType:    &req.ProviderType,
		ContextLength:   &req.ContextLength,
		MaxOutputTokens: req.MaxOutputTokens,
		Pricing:         req.Pricing,
		Capabilities:    req.Capabilities,
		AllowedGroupIDs: req.AllowedGroupIDs,
		Status:          &req.Status,
		Metadata:        req.Metadata,
		RouteBindings:   req.RouteBindings,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, toAdminModel(*item))
}

func (h *Handler) ListUsers(c *gin.Context) {
	page := parsePositiveInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parsePositiveInt(c.DefaultQuery("page_size", "20"), 20)
	result, err := h.service.ListUsers(c.Request.Context(), ListUsersFilter{
		Page:     page,
		PageSize: pageSize,
		Keyword:  c.DefaultQuery("keyword", ""),
		Status:   c.DefaultQuery("status", ""),
	})
	if err != nil {
		h.internalError(c)
		return
	}

	httpx.SuccessWithMeta(c, http.StatusOK, toAdminUsers(result.Items), gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     result.Total,
	})
}

func (h *Handler) AdjustUserQuota(c *gin.Context) {
	var req quotaAdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid quota payload", gin.H{"error": err.Error()})
		return
	}

	item, err := h.service.AdjustUserQuota(c.Request.Context(), actorFromContext(c), c.Param("id"), req.Delta, req.Reason)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, toAdminUser(*item))
}

func (h *Handler) ListAuditLogs(c *gin.Context) {
	page := parsePositiveInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parsePositiveInt(c.DefaultQuery("page_size", "20"), 20)
	result, err := h.service.ListAuditLogs(c.Request.Context(), audit.ListFilter{
		Page:         page,
		PageSize:     pageSize,
		ActorUserID:  c.DefaultQuery("actor_id", ""),
		Action:       c.DefaultQuery("action", ""),
		ResourceType: c.DefaultQuery("resource_type", ""),
		Result:       c.DefaultQuery("result", ""),
	})
	if err != nil {
		h.internalError(c)
		return
	}

	httpx.SuccessWithMeta(c, http.StatusOK, toAuditLogs(result.Items), gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     result.Total,
	})
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, catalog.ErrUpstreamNotFound):
		httpx.Failure(c, http.StatusNotFound, "UPSTREAM_NOT_FOUND", "Upstream not found", nil)
	case errors.Is(err, catalog.ErrModelNotFound):
		httpx.Failure(c, http.StatusNotFound, "MODEL_NOT_FOUND", "Model not found", nil)
	case errors.Is(err, account.ErrUserNotFound):
		httpx.Failure(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found", nil)
	case errors.Is(err, ErrQuotaWouldBecomeNegative):
		httpx.Failure(c, http.StatusBadRequest, "QUOTA_NEGATIVE_NOT_ALLOWED", "Quota cannot become negative", nil)
	default:
		h.internalError(c)
	}
}

func (h *Handler) internalError(c *gin.Context) {
	httpx.Failure(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
}

func actorFromContext(c *gin.Context) ActorContext {
	return ActorContext{
		ActorUserID: middleware.CurrentUserID(c),
		ActorRole:   middleware.CurrentUserRole(c),
		RequestID:   httpx.RequestIDFromContext(c),
		IPAddress:   c.ClientIP(),
		UserAgent:   c.Request.UserAgent(),
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

func toUpstreams(items []catalog.Upstream) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, toUpstream(&item))
	}
	return result
}

func toUpstream(item *catalog.Upstream) gin.H {
	if item == nil {
		return gin.H{}
	}
	return gin.H{
		"id":                item.ID,
		"name":              item.Name,
		"provider_type":     item.ProviderType,
		"base_url":          item.BaseURL,
		"auth_type":         item.AuthType,
		"auth_config":       item.AuthConfig,
		"status":            item.Status,
		"timeout_seconds":   item.TimeoutSeconds,
		"cooldown_seconds":  item.CooldownSeconds,
		"failure_threshold": item.FailureThreshold,
		"metadata":          item.Metadata,
		"created_at":        item.CreatedAt.UTC().Format(timeLayout),
		"updated_at":        item.UpdatedAt.UTC().Format(timeLayout),
	}
}

func toAdminModels(items []catalog.ModelWithBindings) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, toAdminModel(item))
	}
	return result
}

func toAdminModel(item catalog.ModelWithBindings) gin.H {
	bindings := make([]gin.H, 0, len(item.RouteBindings))
	for _, binding := range item.RouteBindings {
		bindings = append(bindings, gin.H{
			"id":          binding.ID,
			"group_id":    binding.GroupID,
			"upstream_id": binding.UpstreamID,
			"priority":    binding.Priority,
			"status":      binding.Status,
		})
	}

	return gin.H{
		"id":                item.Model.ID,
		"model_key":         item.Model.ModelKey,
		"display_name":      item.Model.DisplayName,
		"provider_type":     item.Model.ProviderType,
		"context_length":    item.Model.ContextLength,
		"max_output_tokens": item.Model.MaxOutputTokens,
		"pricing":           item.Model.Pricing,
		"capabilities":      item.Model.Capabilities,
		"allowed_group_ids": item.Model.AllowedGroupIDs,
		"status":            item.Model.Status,
		"metadata":          item.Model.Metadata,
		"route_bindings":    bindings,
		"created_at":        item.Model.CreatedAt.UTC().Format(timeLayout),
		"updated_at":        item.Model.UpdatedAt.UTC().Format(timeLayout),
	}
}

func toAdminUsers(items []account.User) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, toAdminUser(item))
	}
	return result
}

func toAdminUser(item account.User) gin.H {
	return gin.H{
		"id":            item.ID,
		"username":      item.Username,
		"email":         item.Email,
		"display_name":  item.DisplayName,
		"role":          item.Role,
		"status":        item.Status,
		"quota":         item.Quota,
		"used_quota":    item.UsedQuota,
		"last_login_at": formatOptionalTime(item.LastLoginAt),
		"created_at":    item.CreatedAt.UTC().Format(timeLayout),
		"updated_at":    item.UpdatedAt.UTC().Format(timeLayout),
	}
}

func toAuditLogs(items []audit.Log) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, gin.H{
			"id":             item.ID,
			"actor_user_id":  item.ActorUserID,
			"actor_role":     item.ActorRole,
			"action":         item.Action,
			"resource_type":  item.ResourceType,
			"resource_id":    item.ResourceID,
			"target_user_id": item.TargetUserID,
			"request_id":     item.RequestID,
			"ip_address":     item.IPAddress,
			"user_agent":     item.UserAgent,
			"result":         item.Result,
			"detail":         item.Details,
			"created_at":     item.CreatedAt.UTC().Format(timeLayout),
		})
	}
	return result
}

func formatOptionalTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(timeLayout)
}

const timeLayout = "2006-01-02T15:04:05Z07:00"
