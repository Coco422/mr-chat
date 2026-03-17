package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"mrchat/internal/modules/account"
)

var (
	ErrUsernameTaken      = errors.New("username already exists")
	ErrEmailTaken         = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserDisabled       = errors.New("user disabled")
	ErrWeakPassword       = errors.New("password too short")
	ErrMissingIdentifier  = errors.New("missing identifier")
)

type Service struct {
	repo   *account.Repository
	tokens *TokenManager
}

type SignupInput struct {
	Username string
	Email    string
	Password string
}

type SigninInput struct {
	Identifier string
	Password   string
}

type SessionUser struct {
	ID       string               `json:"id"`
	Username string               `json:"username"`
	Email    string               `json:"email"`
	Role     account.Role         `json:"role"`
	Settings account.UserSettings `json:"settings"`
}

type Session struct {
	AccessToken      string      `json:"access_token"`
	ExpiresIn        int64       `json:"expires_in"`
	RefreshToken     string      `json:"-"`
	RefreshExpiresAt time.Time   `json:"-"`
	User             SessionUser `json:"user"`
}

func NewService(repo *account.Repository, tokens *TokenManager) *Service {
	return &Service{
		repo:   repo,
		tokens: tokens,
	}
}

func (s *Service) Signup(ctx context.Context, input SignupInput) (Session, error) {
	username := normalize(input.Username)
	email := normalize(input.Email)
	password := strings.TrimSpace(input.Password)

	if username == "" || email == "" {
		return Session{}, ErrMissingIdentifier
	}
	if len(password) < 8 {
		return Session{}, ErrWeakPassword
	}

	usernameExists, err := s.repo.UsernameExists(ctx, username)
	if err != nil {
		return Session{}, err
	}
	if usernameExists {
		return Session{}, ErrUsernameTaken
	}

	emailExists, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return Session{}, err
	}
	if emailExists {
		return Session{}, ErrEmailTaken
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return Session{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repo.CreateUserWithPassword(ctx, account.CreateUserInput{
		Username:     username,
		Email:        email,
		DisplayName:  username,
		PasswordHash: string(passwordHash),
		Role:         account.RoleUser,
		Status:       account.UserStatusActive,
		Settings: account.UserSettings{
			Timezone: "Asia/Shanghai",
			Locale:   "zh-CN",
		},
	})
	if err != nil {
		return Session{}, err
	}

	now := time.Now().UTC()
	if err := s.repo.UpdateLastLogin(ctx, user.ID, now); err != nil {
		return Session{}, err
	}
	user.LastLoginAt = timePtr(now)

	return s.issueSession(user)
}

func (s *Service) Signin(ctx context.Context, input SigninInput) (Session, error) {
	identifier := normalize(input.Identifier)
	password := strings.TrimSpace(input.Password)
	if identifier == "" {
		return Session{}, ErrMissingIdentifier
	}

	user, authRecord, err := s.repo.FindPasswordAuthByIdentifier(ctx, identifier)
	if err != nil {
		if errors.Is(err, account.ErrUserNotFound) || errors.Is(err, account.ErrPasswordNotFound) {
			return Session{}, ErrInvalidCredentials
		}

		return Session{}, err
	}

	if user.Status != account.UserStatusActive {
		return Session{}, ErrUserDisabled
	}

	if authRecord.PasswordHash == nil || bcrypt.CompareHashAndPassword([]byte(*authRecord.PasswordHash), []byte(password)) != nil {
		return Session{}, ErrInvalidCredentials
	}

	if err := s.repo.UpdateLastLogin(ctx, user.ID, time.Now().UTC()); err != nil {
		return Session{}, err
	}

	user.LastLoginAt = timePtr(time.Now().UTC())
	return s.issueSession(user)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (Session, error) {
	claims, err := s.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return Session{}, ErrInvalidToken
	}

	user, err := s.repo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return Session{}, err
	}
	if user.Status != account.UserStatusActive {
		return Session{}, ErrUserDisabled
	}

	return s.issueSession(user)
}

func (s *Service) issueSession(user *account.User) (Session, error) {
	tokenPair, err := s.tokens.IssueTokens(user.ID, user.Role)
	if err != nil {
		return Session{}, err
	}

	return Session{
		AccessToken:      tokenPair.AccessToken,
		ExpiresIn:        tokenPair.AccessExpiresIn,
		RefreshToken:     tokenPair.RefreshToken,
		RefreshExpiresAt: tokenPair.RefreshExpiresAt,
		User: SessionUser{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
			Settings: user.Settings,
		},
	}, nil
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func timePtr(value time.Time) *time.Time {
	return &value
}
