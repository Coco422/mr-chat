package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"mrchat/internal/app/config"
	"mrchat/internal/shared/httpx"
)

type Handler struct {
	service      *Service
	cookieName   string
	cookieDomain string
	cookieSecure bool
}

type signupRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type signinRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

func NewHandler(cfg config.AuthConfig, service *Service) *Handler {
	return &Handler{
		service:      service,
		cookieName:   cfg.RefreshCookieName,
		cookieDomain: cfg.RefreshCookieDomain,
		cookieSecure: cfg.RefreshCookieSecure,
	}
}

// SignUp godoc
// @Summary Sign up
// @Description Create a new user account and issue access/refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body signupRequest true "Signup payload"
// @Success 201 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 409 {object} httpx.Envelope
// @Router /auth/signup [post]
func (h *Handler) SignUp(c *gin.Context) {
	var req signupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid signup payload", gin.H{"error": err.Error()})
		return
	}

	session, err := h.service.Signup(c.Request.Context(), SignupInput(req))
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	h.setRefreshCookie(c, session.RefreshToken, session.RefreshExpiresAt)
	httpx.Success(c, http.StatusCreated, session)
}

// SignIn godoc
// @Summary Sign in
// @Description Authenticate by username or email and issue access/refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body signinRequest true "Signin payload"
// @Success 200 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Router /auth/signin [post]
func (h *Handler) SignIn(c *gin.Context) {
	var req signinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid signin payload", gin.H{"error": err.Error()})
		return
	}

	session, err := h.service.Signin(c.Request.Context(), SigninInput(req))
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	h.setRefreshCookie(c, session.RefreshToken, session.RefreshExpiresAt)
	httpx.Success(c, http.StatusOK, session)
}

// Refresh godoc
// @Summary Refresh access token
// @Description Refresh access token by refresh cookie
// @Tags Auth
// @Produce json
// @Success 200 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Router /auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie(h.cookieName)
	if err != nil || refreshToken == "" {
		httpx.Failure(c, http.StatusUnauthorized, "AUTH_REFRESH_REQUIRED", "Refresh token is missing", nil)
		return
	}

	session, err := h.service.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	h.setRefreshCookie(c, session.RefreshToken, session.RefreshExpiresAt)
	httpx.Success(c, http.StatusOK, gin.H{
		"access_token": session.AccessToken,
		"expires_in":   session.ExpiresIn,
		"user":         session.User,
	})
}

// SignOut godoc
// @Summary Sign out
// @Description Clear refresh cookie for current client
// @Tags Auth
// @Produce json
// @Success 200 {object} httpx.Envelope
// @Router /auth/signout [post]
func (h *Handler) SignOut(c *gin.Context) {
	h.clearRefreshCookie(c)
	httpx.Success(c, http.StatusOK, gin.H{
		"signed_out": true,
	})
}

func (h *Handler) handleAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrUsernameTaken):
		httpx.Failure(c, http.StatusConflict, "AUTH_USERNAME_TAKEN", "Username already exists", nil)
	case errors.Is(err, ErrEmailTaken):
		httpx.Failure(c, http.StatusConflict, "AUTH_EMAIL_TAKEN", "Email already exists", nil)
	case errors.Is(err, ErrWeakPassword):
		httpx.Failure(c, http.StatusBadRequest, "AUTH_WEAK_PASSWORD", "Password must be at least 8 characters", nil)
	case errors.Is(err, ErrMissingIdentifier):
		httpx.Failure(c, http.StatusBadRequest, "AUTH_INVALID_INPUT", "Identifier and password are required", nil)
	case errors.Is(err, ErrInvalidCredentials):
		httpx.Failure(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Invalid credentials", nil)
	case errors.Is(err, ErrUserDisabled):
		httpx.Failure(c, http.StatusForbidden, "AUTH_USER_DISABLED", "User is not active", nil)
	case errors.Is(err, ErrInvalidToken):
		httpx.Failure(c, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "Invalid or expired token", nil)
	default:
		httpx.Failure(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
	}
}

func (h *Handler) setRefreshCookie(c *gin.Context, value string, expiresAt time.Time) {
	cookie := &http.Cookie{
		Name:     h.cookieName,
		Value:    value,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
	}
	if h.cookieDomain != "" {
		cookie.Domain = h.cookieDomain
	}

	http.SetCookie(c.Writer, cookie)
}

func (h *Handler) clearRefreshCookie(c *gin.Context) {
	cookie := &http.Cookie{
		Name:     h.cookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}
	if h.cookieDomain != "" {
		cookie.Domain = h.cookieDomain
	}

	http.SetCookie(c.Writer, cookie)
}
