package users

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"mrchat/internal/http/middleware"
	"mrchat/internal/modules/account"
	"mrchat/internal/shared/httpx"
)

type Handler struct {
	service *Service
}

type updateProfileRequest struct {
	DisplayName *string               `json:"display_name"`
	AvatarURL   *string               `json:"avatar_url"`
	Settings    *account.UserSettings `json:"settings"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetMe godoc
// @Summary Get current user profile
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Router /users/me [get]
func (h *Handler) GetMe(c *gin.Context) {
	profile, err := h.service.GetMe(c.Request.Context(), middleware.CurrentUserID(c))
	if err != nil {
		h.handleUserError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, profile)
}

// UpdateMe godoc
// @Summary Update current user profile
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body updateProfileRequest true "Profile payload"
// @Success 200 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Router /users/me [put]
func (h *Handler) UpdateMe(c *gin.Context) {
	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid profile payload", gin.H{"error": err.Error()})
		return
	}

	profile, err := h.service.UpdateMe(c.Request.Context(), middleware.CurrentUserID(c), UpdateProfileInput(req))
	if err != nil {
		h.handleUserError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, profile)
}

// GetQuota godoc
// @Summary Get current user quota
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Router /users/me/quota [get]
func (h *Handler) GetQuota(c *gin.Context) {
	quota, err := h.service.GetQuota(c.Request.Context(), middleware.CurrentUserID(c))
	if err != nil {
		h.handleUserError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, quota)
}

// GetUsage godoc
// @Summary Get current user usage
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param range query string false "Usage range" Enums(7d,30d,month)
// @Success 200 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Router /users/me/usage [get]
func (h *Handler) GetUsage(c *gin.Context) {
	usage, err := h.service.GetUsage(c.Request.Context(), middleware.CurrentUserID(c), c.DefaultQuery("range", "7d"))
	if err != nil {
		h.handleUserError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, usage)
}

// GetSecurity godoc
// @Summary Get current user security info
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Router /users/me/security [get]
func (h *Handler) GetSecurity(c *gin.Context) {
	security, err := h.service.GetSecurity(c.Request.Context(), middleware.CurrentUserID(c))
	if err != nil {
		h.handleUserError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, security)
}

// ChangePassword godoc
// @Summary Change current user password
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body changePasswordRequest true "Password payload"
// @Success 200 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Router /users/me/password [put]
func (h *Handler) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid password payload", gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), middleware.CurrentUserID(c), ChangePasswordInput(req)); err != nil {
		h.handleUserError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, gin.H{"password_updated": true})
}

func (h *Handler) handleUserError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, account.ErrUserNotFound):
		httpx.Failure(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found", nil)
	case errors.Is(err, ErrInvalidUsageRange):
		httpx.Failure(c, http.StatusBadRequest, "USAGE_INVALID_RANGE", "range must be one of 7d, 30d, month", nil)
	case errors.Is(err, ErrCurrentPasswordInvalid):
		httpx.Failure(c, http.StatusBadRequest, "AUTH_CURRENT_PASSWORD_INVALID", "Current password is incorrect", nil)
	case errors.Is(err, ErrWeakPassword):
		httpx.Failure(c, http.StatusBadRequest, "AUTH_WEAK_PASSWORD", "Password must be at least 8 characters", nil)
	default:
		httpx.Failure(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
	}
}

func toProfile(user *account.User) Profile {
	return Profile{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Role:        user.Role,
		Status:      user.Status,
		Quota:       user.Quota,
		UsedQuota:   user.UsedQuota,
		Settings:    user.Settings,
		CreatedAt:   user.CreatedAt.UTC().Format(timeLayout),
		UpdatedAt:   user.UpdatedAt.UTC().Format(timeLayout),
	}
}

const timeLayout = "2006-01-02T15:04:05Z07:00"
