package admin

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"mrchat/internal/http/middleware"
	"mrchat/internal/modules/account"
	"mrchat/internal/modules/audit"
	"mrchat/internal/modules/catalog"
	"mrchat/internal/modules/limits"
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

type channelRequest struct {
	Name          string         `json:"name"`
	Description   *string        `json:"description"`
	Status        string         `json:"status"`
	BillingConfig map[string]any `json:"billing_config"`
	Metadata      map[string]any `json:"metadata"`
}

type modelRequest struct {
	ModelKey            string                      `json:"model_key"`
	DisplayName         string                      `json:"display_name"`
	ProviderType        string                      `json:"provider_type"`
	ContextLength       int                         `json:"context_length"`
	MaxOutputTokens     *int                        `json:"max_output_tokens"`
	Pricing             map[string]any              `json:"pricing"`
	Capabilities        map[string]any              `json:"capabilities"`
	VisibleUserGroupIDs []string                    `json:"visible_user_group_ids"`
	Status              string                      `json:"status"`
	Metadata            map[string]any              `json:"metadata"`
	RouteBindings       []catalog.RouteBindingInput `json:"route_bindings"`
}

type importModelsRequest struct {
	UpstreamID string                   `json:"upstream_id"`
	Items      []importModelItemRequest `json:"items"`
}

type importModelItemRequest struct {
	ModelKey            string         `json:"model_key"`
	DisplayName         string         `json:"display_name"`
	ProviderType        string         `json:"provider_type"`
	ContextLength       int            `json:"context_length"`
	MaxOutputTokens     *int           `json:"max_output_tokens"`
	Pricing             map[string]any `json:"pricing"`
	Capabilities        map[string]any `json:"capabilities"`
	VisibleUserGroupIDs []string       `json:"visible_user_group_ids"`
	Status              string         `json:"status"`
	Metadata            map[string]any `json:"metadata"`
	ChannelID           *string        `json:"channel_id"`
	Priority            int            `json:"priority"`
}

type userGroupRequest struct {
	Name        string         `json:"name"`
	Description *string        `json:"description"`
	Status      string         `json:"status"`
	Permissions map[string]any `json:"permissions"`
	Metadata    map[string]any `json:"metadata"`
}

type replacePoliciesRequest struct {
	Policies []policyRequest `json:"policies"`
}

type policyRequest struct {
	ModelID              *string `json:"model_id"`
	HourRequestLimit     *int64  `json:"hour_request_limit"`
	WeekRequestLimit     *int64  `json:"week_request_limit"`
	LifetimeRequestLimit *int64  `json:"lifetime_request_limit"`
	HourTokenLimit       *int64  `json:"hour_token_limit"`
	WeekTokenLimit       *int64  `json:"week_token_limit"`
	LifetimeTokenLimit   *int64  `json:"lifetime_token_limit"`
	Status               string  `json:"status"`
}

type assignUserGroupRequest struct {
	UserGroupID *string `json:"user_group_id"`
}

type quotaAdjustmentRequest struct {
	Delta  int64  `json:"delta"`
	Reason string `json:"reason"`
}

type userLimitAdjustmentRequest struct {
	ModelID    *string `json:"model_id"`
	MetricType string  `json:"metric_type"`
	WindowType string  `json:"window_type"`
	Delta      int64   `json:"delta"`
	Reason     *string `json:"reason"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ListUpstreams godoc
// @Summary List upstreams
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Router /admin/upstreams [get]
func (h *Handler) ListUpstreams(c *gin.Context) {
	items, err := h.service.ListUpstreams(c.Request.Context())
	if err != nil {
		h.internalError(c)
		return
	}
	httpx.Success(c, http.StatusOK, toUpstreams(items))
}

// GetUpstream godoc
// @Summary Get upstream detail
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Upstream ID"
// @Success 200 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Failure 404 {object} httpx.Envelope
// @Router /admin/upstreams/{id} [get]
func (h *Handler) GetUpstream(c *gin.Context) {
	item, err := h.service.GetUpstream(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, toUpstream(item))
}

// DiscoverUpstreamModels godoc
// @Summary Discover upstream models
// @Description Fetch candidate models from the upstream `/v1/models` endpoint and annotate whether they are already imported locally.
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Upstream ID"
// @Success 200 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Failure 404 {object} httpx.Envelope
// @Failure 502 {object} httpx.Envelope
// @Router /admin/upstreams/{id}/discovered-models [get]
func (h *Handler) DiscoverUpstreamModels(c *gin.Context) {
	result, err := h.service.DiscoverUpstreamModels(c.Request.Context(), actorFromContext(c), c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, toUpstreamDiscoveryResult(result))
}

// CreateUpstream godoc
// @Summary Create upstream
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body upstreamRequest true "Upstream payload"
// @Success 201 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Router /admin/upstreams [post]
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

// UpdateUpstream godoc
// @Summary Update upstream
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Upstream ID"
// @Param request body upstreamRequest true "Upstream payload"
// @Success 200 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Failure 404 {object} httpx.Envelope
// @Router /admin/upstreams/{id} [put]
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

// ListChannels godoc
// @Summary List channels
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Router /admin/channels [get]
func (h *Handler) ListChannels(c *gin.Context) {
	items, err := h.service.ListChannels(c.Request.Context())
	if err != nil {
		h.internalError(c)
		return
	}
	httpx.Success(c, http.StatusOK, toChannels(items))
}

// GetChannel godoc
// @Summary Get channel detail
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Channel ID"
// @Success 200 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Failure 404 {object} httpx.Envelope
// @Router /admin/channels/{id} [get]
func (h *Handler) GetChannel(c *gin.Context) {
	item, err := h.service.GetChannel(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, toChannel(item))
}

// CreateChannel godoc
// @Summary Create channel
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body channelRequest true "Channel payload"
// @Success 201 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Router /admin/channels [post]
func (h *Handler) CreateChannel(c *gin.Context) {
	var req channelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid channel payload", gin.H{"error": err.Error()})
		return
	}

	item, err := h.service.CreateChannel(c.Request.Context(), actorFromContext(c), CreateChannelInput{
		Name:          req.Name,
		Description:   req.Description,
		Status:        req.Status,
		BillingConfig: req.BillingConfig,
		Metadata:      req.Metadata,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusCreated, toChannel(item))
}

// UpdateChannel godoc
// @Summary Update channel
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Channel ID"
// @Param request body channelRequest true "Channel payload"
// @Success 200 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Failure 404 {object} httpx.Envelope
// @Router /admin/channels/{id} [put]
func (h *Handler) UpdateChannel(c *gin.Context) {
	var req channelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid channel payload", gin.H{"error": err.Error()})
		return
	}

	item, err := h.service.UpdateChannel(c.Request.Context(), actorFromContext(c), c.Param("id"), UpdateChannelInput{
		Name:          &req.Name,
		Description:   req.Description,
		Status:        &req.Status,
		BillingConfig: req.BillingConfig,
		Metadata:      req.Metadata,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, toChannel(item))
}

// ListModels godoc
// @Summary List admin models
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Router /admin/models [get]
func (h *Handler) ListModels(c *gin.Context) {
	items, err := h.service.ListModels(c.Request.Context())
	if err != nil {
		h.internalError(c)
		return
	}
	httpx.Success(c, http.StatusOK, toAdminModels(items))
}

// GetModel godoc
// @Summary Get model detail
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Model ID"
// @Success 200 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Failure 404 {object} httpx.Envelope
// @Router /admin/models/{id} [get]
func (h *Handler) GetModel(c *gin.Context) {
	item, err := h.service.GetModel(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, toAdminModel(*item))
}

// CreateModel godoc
// @Summary Create model
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body modelRequest true "Model payload"
// @Success 201 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Router /admin/models [post]
func (h *Handler) CreateModel(c *gin.Context) {
	var req modelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid model payload", gin.H{"error": err.Error()})
		return
	}

	item, err := h.service.CreateModel(c.Request.Context(), actorFromContext(c), CreateModelInput{
		ModelKey:            req.ModelKey,
		DisplayName:         req.DisplayName,
		ProviderType:        req.ProviderType,
		ContextLength:       req.ContextLength,
		MaxOutputTokens:     req.MaxOutputTokens,
		Pricing:             req.Pricing,
		Capabilities:        req.Capabilities,
		VisibleUserGroupIDs: req.VisibleUserGroupIDs,
		Status:              req.Status,
		Metadata:            req.Metadata,
		RouteBindings:       req.RouteBindings,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	view, err := h.service.GetModel(c.Request.Context(), item.Model.ID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusCreated, toAdminModel(*view))
}

// UpdateModel godoc
// @Summary Update model
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Model ID"
// @Param request body modelRequest true "Model payload"
// @Success 200 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Failure 404 {object} httpx.Envelope
// @Router /admin/models/{id} [put]
func (h *Handler) UpdateModel(c *gin.Context) {
	var req modelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid model payload", gin.H{"error": err.Error()})
		return
	}

	item, err := h.service.UpdateModel(c.Request.Context(), actorFromContext(c), c.Param("id"), UpdateModelInput{
		ModelKey:            &req.ModelKey,
		DisplayName:         &req.DisplayName,
		ProviderType:        &req.ProviderType,
		ContextLength:       &req.ContextLength,
		MaxOutputTokens:     req.MaxOutputTokens,
		Pricing:             req.Pricing,
		Capabilities:        req.Capabilities,
		VisibleUserGroupIDs: req.VisibleUserGroupIDs,
		Status:              &req.Status,
		Metadata:            req.Metadata,
		RouteBindings:       req.RouteBindings,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	view, err := h.service.GetModel(c.Request.Context(), item.Model.ID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, toAdminModel(*view))
}

// ImportModels godoc
// @Summary Import discovered models
// @Description Import one or more discovered upstream models into the local platform model catalog.
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body importModelsRequest true "Import payload"
// @Success 201 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Failure 404 {object} httpx.Envelope
// @Router /admin/models/import [post]
func (h *Handler) ImportModels(c *gin.Context) {
	var req importModelsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid model import payload", gin.H{"error": err.Error()})
		return
	}

	inputs := make([]ImportModelItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		inputs = append(inputs, ImportModelItemInput{
			ModelKey:            item.ModelKey,
			DisplayName:         item.DisplayName,
			ProviderType:        item.ProviderType,
			ContextLength:       item.ContextLength,
			MaxOutputTokens:     item.MaxOutputTokens,
			Pricing:             item.Pricing,
			Capabilities:        item.Capabilities,
			VisibleUserGroupIDs: item.VisibleUserGroupIDs,
			Status:              item.Status,
			Metadata:            item.Metadata,
			ChannelID:           item.ChannelID,
			Priority:            item.Priority,
		})
	}

	result, err := h.service.ImportModels(c.Request.Context(), actorFromContext(c), ImportModelsInput{
		UpstreamID: req.UpstreamID,
		Items:      inputs,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusCreated, toImportModelsResult(result))
}

// ListUserGroups godoc
// @Summary List user groups
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Router /admin/user-groups [get]
func (h *Handler) ListUserGroups(c *gin.Context) {
	items, err := h.service.ListUserGroups(c.Request.Context())
	if err != nil {
		h.internalError(c)
		return
	}
	httpx.Success(c, http.StatusOK, toUserGroups(items))
}

// GetUserGroup godoc
// @Summary Get user group detail
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "User group ID"
// @Success 200 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Failure 404 {object} httpx.Envelope
// @Router /admin/user-groups/{id} [get]
func (h *Handler) GetUserGroup(c *gin.Context) {
	item, err := h.service.GetUserGroup(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, toUserGroup(item))
}

// CreateUserGroup godoc
// @Summary Create user group
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body userGroupRequest true "User group payload"
// @Success 201 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Router /admin/user-groups [post]
func (h *Handler) CreateUserGroup(c *gin.Context) {
	var req userGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid user group payload", gin.H{"error": err.Error()})
		return
	}

	item, err := h.service.CreateUserGroup(c.Request.Context(), actorFromContext(c), CreateUserGroupInput{
		Name:        req.Name,
		Description: req.Description,
		Status:      account.UserGroupStatus(strings.TrimSpace(req.Status)),
		Permissions: req.Permissions,
		Metadata:    req.Metadata,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusCreated, toUserGroup(item))
}

// UpdateUserGroup godoc
// @Summary Update user group
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User group ID"
// @Param request body userGroupRequest true "User group payload"
// @Success 200 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Failure 404 {object} httpx.Envelope
// @Router /admin/user-groups/{id} [put]
func (h *Handler) UpdateUserGroup(c *gin.Context) {
	var req userGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid user group payload", gin.H{"error": err.Error()})
		return
	}

	status := account.UserGroupStatus(strings.TrimSpace(req.Status))
	item, err := h.service.UpdateUserGroup(c.Request.Context(), actorFromContext(c), c.Param("id"), UpdateUserGroupInput{
		Name:        &req.Name,
		Description: req.Description,
		Status:      &status,
		Permissions: req.Permissions,
		Metadata:    req.Metadata,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, toUserGroup(item))
}

// GetUserGroupLimitPolicies godoc
// @Summary Get user group limit policies
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "User group ID"
// @Success 200 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Failure 404 {object} httpx.Envelope
// @Router /admin/user-groups/{id}/limits [get]
func (h *Handler) GetUserGroupLimitPolicies(c *gin.Context) {
	items, err := h.service.ListUserGroupLimitPolicies(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, toPolicies(items))
}

// UpdateUserGroupLimitPolicies godoc
// @Summary Replace user group limit policies
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User group ID"
// @Param request body replacePoliciesRequest true "Policy payload"
// @Success 200 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Failure 404 {object} httpx.Envelope
// @Router /admin/user-groups/{id}/limits [put]
func (h *Handler) UpdateUserGroupLimitPolicies(c *gin.Context) {
	var req replacePoliciesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid limit policy payload", gin.H{"error": err.Error()})
		return
	}

	inputs := make([]PolicyUpsertInput, 0, len(req.Policies))
	for _, policy := range req.Policies {
		inputs = append(inputs, PolicyUpsertInput{
			ModelID:              policy.ModelID,
			HourRequestLimit:     policy.HourRequestLimit,
			WeekRequestLimit:     policy.WeekRequestLimit,
			LifetimeRequestLimit: policy.LifetimeRequestLimit,
			HourTokenLimit:       policy.HourTokenLimit,
			WeekTokenLimit:       policy.WeekTokenLimit,
			LifetimeTokenLimit:   policy.LifetimeTokenLimit,
			Status:               limits.PolicyStatus(strings.TrimSpace(policy.Status)),
		})
	}

	items, err := h.service.ReplaceUserGroupLimitPolicies(c.Request.Context(), actorFromContext(c), c.Param("id"), inputs)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, toPolicies(items))
}

// ListUsers godoc
// @Summary List users
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page"
// @Param page_size query int false "Page size"
// @Param keyword query string false "Keyword"
// @Param status query string false "User status"
// @Success 200 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Router /admin/users [get]
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

// AssignUserGroup godoc
// @Summary Assign user group
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body assignUserGroupRequest true "Assign user group payload"
// @Success 200 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Failure 404 {object} httpx.Envelope
// @Router /admin/users/{id}/group [put]
func (h *Handler) AssignUserGroup(c *gin.Context) {
	var req assignUserGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid user group assignment payload", gin.H{"error": err.Error()})
		return
	}

	item, err := h.service.AssignUserGroup(c.Request.Context(), actorFromContext(c), c.Param("id"), req.UserGroupID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, toAdminUser(AdminUserRecord{User: *item}))
}

// AdjustUserQuota godoc
// @Summary Adjust user quota
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body quotaAdjustmentRequest true "Quota adjustment payload"
// @Success 200 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Failure 404 {object} httpx.Envelope
// @Router /admin/users/{id}/quota [put]
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

	httpx.Success(c, http.StatusOK, toAdminUser(AdminUserRecord{User: *item}))
}

// GetUserLimitUsage godoc
// @Summary Get user limit usage
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param model_id query string false "Model ID"
// @Success 200 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Failure 404 {object} httpx.Envelope
// @Router /admin/users/{id}/limit-usage [get]
func (h *Handler) GetUserLimitUsage(c *gin.Context) {
	modelID := optionalQueryString(c.Query("model_id"))
	report, err := h.service.GetUserLimitUsage(c.Request.Context(), c.Param("id"), modelID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, report)
}

// ListUserLimitAdjustments godoc
// @Summary List user limit adjustments
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param model_id query string false "Model ID"
// @Param page query int false "Page"
// @Param page_size query int false "Page size"
// @Success 200 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Failure 404 {object} httpx.Envelope
// @Router /admin/users/{id}/limit-adjustments [get]
func (h *Handler) ListUserLimitAdjustments(c *gin.Context) {
	page := parsePositiveInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parsePositiveInt(c.DefaultQuery("page_size", "20"), 20)
	modelID := optionalQueryString(c.Query("model_id"))
	result, err := h.service.ListUserLimitAdjustments(c.Request.Context(), c.Param("id"), modelID, page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.SuccessWithMeta(c, http.StatusOK, toUserLimitAdjustments(result.Items), gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     result.Total,
	})
}

// CreateUserLimitAdjustment godoc
// @Summary Create user limit adjustment
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body userLimitAdjustmentRequest true "User limit adjustment payload"
// @Success 201 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Failure 404 {object} httpx.Envelope
// @Router /admin/users/{id}/limit-adjustments [post]
func (h *Handler) CreateUserLimitAdjustment(c *gin.Context) {
	var req userLimitAdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid user limit adjustment payload", gin.H{"error": err.Error()})
		return
	}

	metricType, ok := parseMetricType(req.MetricType)
	if !ok {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "metric_type must be request_count or total_tokens", nil)
		return
	}
	windowType, ok := parseWindowType(req.WindowType)
	if !ok {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "window_type must be rolling_hour, rolling_week or lifetime", nil)
		return
	}

	item, err := h.service.CreateUserLimitAdjustment(c.Request.Context(), actorFromContext(c), c.Param("id"), CreateUserLimitAdjustmentInput{
		ModelID:    req.ModelID,
		MetricType: metricType,
		WindowType: windowType,
		Delta:      req.Delta,
		Reason:     req.Reason,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusCreated, toUserLimitAdjustment(item))
}

// ListAuditLogs godoc
// @Summary List audit logs
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page"
// @Param page_size query int false "Page size"
// @Param actor_id query string false "Actor user ID"
// @Param action query string false "Action"
// @Param resource_type query string false "Resource type"
// @Param result query string false "Result"
// @Success 200 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 403 {object} httpx.Envelope
// @Router /admin/audit-logs [get]
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
	case errors.Is(err, catalog.ErrChannelNotFound):
		httpx.Failure(c, http.StatusNotFound, "CHANNEL_NOT_FOUND", "Channel not found", nil)
	case errors.Is(err, account.ErrUserNotFound):
		httpx.Failure(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found", nil)
	case errors.Is(err, account.ErrUserGroupNotFound):
		httpx.Failure(c, http.StatusNotFound, "USER_GROUP_NOT_FOUND", "User group not found", nil)
	case errors.Is(err, ErrQuotaWouldBecomeNegative):
		httpx.Failure(c, http.StatusBadRequest, "QUOTA_NEGATIVE_NOT_ALLOWED", "Quota cannot become negative", nil)
	case errors.Is(err, ErrUpstreamDiscoveryUnsupported):
		httpx.Failure(c, http.StatusBadRequest, "UPSTREAM_DISCOVERY_UNSUPPORTED", "The current upstream provider does not support model discovery yet", nil)
	case errors.Is(err, ErrUpstreamDiscoveryFailed):
		httpx.Failure(c, http.StatusBadGateway, "UPSTREAM_DISCOVERY_FAILED", "Failed to fetch model list from upstream", nil)
	case errors.Is(err, ErrModelImportInvalid):
		httpx.Failure(c, http.StatusBadRequest, "MODEL_IMPORT_INVALID", "Invalid model import payload", nil)
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

func parseMetricType(value string) (limits.MetricType, bool) {
	switch strings.TrimSpace(value) {
	case string(limits.MetricTypeRequestCount):
		return limits.MetricTypeRequestCount, true
	case string(limits.MetricTypeTotalTokens):
		return limits.MetricTypeTotalTokens, true
	default:
		return "", false
	}
}

func parseWindowType(value string) (limits.WindowType, bool) {
	switch strings.TrimSpace(value) {
	case string(limits.WindowTypeRollingHour):
		return limits.WindowTypeRollingHour, true
	case string(limits.WindowTypeRollingWeek):
		return limits.WindowTypeRollingWeek, true
	case string(limits.WindowTypeLifetime):
		return limits.WindowTypeLifetime, true
	default:
		return "", false
	}
}

func optionalQueryString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
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
		"auth_config":       sanitizeAuthConfig(item.AuthConfig),
		"status":            item.Status,
		"timeout_seconds":   item.TimeoutSeconds,
		"cooldown_seconds":  item.CooldownSeconds,
		"failure_threshold": item.FailureThreshold,
		"metadata":          item.Metadata,
		"created_at":        item.CreatedAt.UTC().Format(timeLayout),
		"updated_at":        item.UpdatedAt.UTC().Format(timeLayout),
	}
}

func toUpstreamDiscoveryResult(result *UpstreamModelDiscoveryResult) gin.H {
	if result == nil {
		return gin.H{}
	}

	items := make([]gin.H, 0, len(result.Items))
	for _, item := range result.Items {
		entry := gin.H{
			"model_key":                item.ModelKey,
			"display_name":             item.DisplayName,
			"provider_type":            item.ProviderType,
			"object":                   item.Object,
			"owned_by":                 item.OwnedBy,
			"supported_endpoint_types": item.SupportedEndpointTypes,
			"capabilities":             item.Capabilities,
			"raw":                      item.Raw,
			"already_imported":         item.AlreadyImported,
		}
		if item.ExistingModel != nil {
			entry["existing_model"] = gin.H{
				"id":           item.ExistingModel.ID,
				"model_key":    item.ExistingModel.ModelKey,
				"display_name": item.ExistingModel.DisplayName,
				"status":       item.ExistingModel.Status,
			}
		} else {
			entry["existing_model"] = nil
		}
		items = append(items, entry)
	}

	return gin.H{
		"upstream":   toUpstream(result.Upstream),
		"items":      items,
		"fetched_at": result.FetchedAt,
		"summary":    result.Summary,
	}
}

func toChannels(items []catalog.Channel) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, toChannel(&item))
	}
	return result
}

func toChannel(item *catalog.Channel) gin.H {
	if item == nil {
		return gin.H{}
	}
	return gin.H{
		"id":             item.ID,
		"name":           item.Name,
		"description":    item.Description,
		"status":         item.Status,
		"billing_config": item.BillingConfig,
		"metadata":       item.Metadata,
		"created_at":     item.CreatedAt.UTC().Format(timeLayout),
		"updated_at":     item.UpdatedAt.UTC().Format(timeLayout),
	}
}

func toAdminModels(items []AdminModelView) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, toAdminModel(item))
	}
	return result
}

func toAdminModel(item AdminModelView) gin.H {
	bindings := make([]gin.H, 0, len(item.HydratedBindings))
	for _, binding := range item.HydratedBindings {
		var channel any
		if binding.Channel != nil {
			channel = gin.H{
				"id":          binding.Channel.ID,
				"name":        binding.Channel.Name,
				"description": binding.Channel.Description,
				"status":      binding.Channel.Status,
			}
		}

		var upstream any
		if binding.Upstream != nil {
			upstream = gin.H{
				"id":            binding.Upstream.ID,
				"name":          binding.Upstream.Name,
				"provider_type": binding.Upstream.ProviderType,
				"status":        binding.Upstream.Status,
				"base_url":      binding.Upstream.BaseURL,
			}
		}

		bindings = append(bindings, gin.H{
			"id":          binding.Binding.ID,
			"channel_id":  binding.Binding.ChannelID,
			"upstream_id": binding.Binding.UpstreamID,
			"priority":    binding.Binding.Priority,
			"status":      binding.Binding.Status,
			"channel":     channel,
			"upstream":    upstream,
			"summary":     summarizeRouteBinding(binding.Binding, binding.Channel, binding.Upstream),
		})
	}

	visibleUserGroups := make([]gin.H, 0, len(item.VisibleUserGroups))
	for _, group := range item.VisibleUserGroups {
		visibleUserGroups = append(visibleUserGroups, gin.H{
			"id":          group.ID,
			"name":        group.Name,
			"description": group.Description,
			"status":      group.Status,
		})
	}

	return gin.H{
		"id":                     item.Item.Model.ID,
		"model_key":              item.Item.Model.ModelKey,
		"display_name":           item.Item.Model.DisplayName,
		"provider_type":          item.Item.Model.ProviderType,
		"context_length":         item.Item.Model.ContextLength,
		"max_output_tokens":      item.Item.Model.MaxOutputTokens,
		"pricing":                item.Item.Model.Pricing,
		"capabilities":           item.Item.Model.Capabilities,
		"visible_user_group_ids": item.Item.Model.VisibleUserGroupIDs,
		"visible_user_groups":    visibleUserGroups,
		"visibility_summary":     item.VisibilitySummary,
		"status":                 item.Item.Model.Status,
		"metadata":               item.Item.Model.Metadata,
		"route_bindings":         bindings,
		"route_rule_summaries":   item.RouteRuleSummaries,
		"created_at":             item.Item.Model.CreatedAt.UTC().Format(timeLayout),
		"updated_at":             item.Item.Model.UpdatedAt.UTC().Format(timeLayout),
	}
}

func toImportModelsResult(result *ImportModelsResult) gin.H {
	if result == nil {
		return gin.H{}
	}

	items := make([]gin.H, 0, len(result.Items))
	for _, item := range result.Items {
		entry := gin.H{
			"requested_model_key": item.RequestedModelKey,
			"status":              item.Status,
		}
		if item.ExistingModel != nil {
			entry["existing_model"] = gin.H{
				"id":           item.ExistingModel.ID,
				"model_key":    item.ExistingModel.ModelKey,
				"display_name": item.ExistingModel.DisplayName,
				"status":       item.ExistingModel.Status,
			}
		} else {
			entry["existing_model"] = nil
		}
		if item.Model != nil {
			entry["model"] = toAdminModel(*item.Model)
		} else {
			entry["model"] = nil
		}
		items = append(items, entry)
	}

	return gin.H{
		"upstream": toUpstream(result.Upstream),
		"items":    items,
		"summary":  result.Summary,
	}
}

func toUserGroups(items []account.UserGroup) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, toUserGroup(&item))
	}
	return result
}

func toUserGroup(item *account.UserGroup) gin.H {
	if item == nil {
		return gin.H{}
	}
	return gin.H{
		"id":          item.ID,
		"name":        item.Name,
		"description": item.Description,
		"status":      item.Status,
		"permissions": item.Permissions,
		"metadata":    item.Metadata,
		"created_at":  item.CreatedAt.UTC().Format(timeLayout),
		"updated_at":  item.UpdatedAt.UTC().Format(timeLayout),
	}
}

func toPolicies(items []limits.UserGroupModelLimitPolicy) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, gin.H{
			"id":                     item.ID,
			"user_group_id":          item.UserGroupID,
			"model_id":               item.ModelID,
			"hour_request_limit":     item.HourRequestLimit,
			"week_request_limit":     item.WeekRequestLimit,
			"lifetime_request_limit": item.LifetimeRequestLimit,
			"hour_token_limit":       item.HourTokenLimit,
			"week_token_limit":       item.WeekTokenLimit,
			"lifetime_token_limit":   item.LifetimeTokenLimit,
			"status":                 item.Status,
			"created_at":             item.CreatedAt.UTC().Format(timeLayout),
			"updated_at":             item.UpdatedAt.UTC().Format(timeLayout),
		})
	}
	return result
}

func toAdminUsers(items []AdminUserRecord) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, toAdminUser(item))
	}
	return result
}

func toAdminUser(item AdminUserRecord) gin.H {
	return gin.H{
		"id":            item.User.ID,
		"username":      item.User.Username,
		"email":         item.User.Email,
		"display_name":  item.User.DisplayName,
		"role":          item.User.Role,
		"status":        item.User.Status,
		"quota":         item.User.Quota,
		"used_quota":    item.User.UsedQuota,
		"user_group_id": item.User.UserGroupID,
		"user_group":    toUserGroup(item.UserGroup),
		"last_login_at": formatOptionalTime(item.User.LastLoginAt),
		"created_at":    item.User.CreatedAt.UTC().Format(timeLayout),
		"updated_at":    item.User.UpdatedAt.UTC().Format(timeLayout),
	}
}

func toUserLimitAdjustments(items []limits.UserLimitAdjustment) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, toUserLimitAdjustment(&item))
	}
	return result
}

func toUserLimitAdjustment(item *limits.UserLimitAdjustment) gin.H {
	if item == nil {
		return gin.H{}
	}

	return gin.H{
		"id":            item.ID,
		"user_id":       item.UserID,
		"model_id":      item.ModelID,
		"metric_type":   item.MetricType,
		"window_type":   item.WindowType,
		"delta":         item.Delta,
		"expires_at":    formatOptionalTime(item.ExpiresAt),
		"reason":        item.Reason,
		"actor_user_id": item.ActorUserID,
		"created_at":    item.CreatedAt.UTC().Format(timeLayout),
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

func sanitizeAuthConfig(value map[string]any) map[string]any {
	if len(value) == 0 {
		return map[string]any{}
	}

	result := make(map[string]any, len(value))
	for key, rawValue := range value {
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "api_key", "token", "access_token", "authorization", "password":
			result[key] = maskSecretValue(rawValue)
			result[key+"_configured"] = strings.TrimSpace(fmt.Sprint(rawValue)) != ""
		default:
			result[key] = rawValue
		}
	}

	return result
}

func maskSecretValue(value any) any {
	raw := strings.TrimSpace(fmt.Sprint(value))
	if raw == "" {
		return nil
	}
	if len(raw) <= 8 {
		return "***"
	}
	return raw[:4] + strings.Repeat("*", len(raw)-8) + raw[len(raw)-4:]
}
