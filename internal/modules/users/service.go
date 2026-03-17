package users

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"mrchat/internal/modules/account"
)

var (
	ErrInvalidUsageRange      = errors.New("invalid usage range")
	ErrCurrentPasswordInvalid = errors.New("current password invalid")
	ErrWeakPassword           = errors.New("password too short")
)

type Service struct {
	repo *account.Repository
}

type Profile struct {
	ID          string               `json:"id"`
	Username    string               `json:"username"`
	Email       string               `json:"email"`
	DisplayName string               `json:"display_name"`
	AvatarURL   *string              `json:"avatar_url"`
	Role        account.Role         `json:"role"`
	Status      account.UserStatus   `json:"status"`
	Quota       int64                `json:"quota"`
	UsedQuota   int64                `json:"used_quota"`
	Settings    account.UserSettings `json:"settings"`
	CreatedAt   string               `json:"created_at"`
	UpdatedAt   string               `json:"updated_at"`
}

type UpdateProfileInput struct {
	DisplayName *string
	AvatarURL   *string
	Settings    *account.UserSettings
}

type ChangePasswordInput struct {
	CurrentPassword string
	NewPassword     string
}

type QuotaSnapshot struct {
	Quota          int64 `json:"quota"`
	UsedQuota      int64 `json:"used_quota"`
	RemainingQuota int64 `json:"remaining_quota"`
}

func NewService(repo *account.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetMe(ctx context.Context, userID string) (Profile, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return Profile{}, err
	}

	return toProfile(user), nil
}

func (s *Service) UpdateMe(ctx context.Context, userID string, input UpdateProfileInput) (Profile, error) {
	user, err := s.repo.UpdateUserProfile(ctx, userID, account.UpdateUserProfileInput{
		DisplayName: input.DisplayName,
		AvatarURL:   input.AvatarURL,
		Settings:    input.Settings,
	})
	if err != nil {
		return Profile{}, err
	}

	return toProfile(user), nil
}

func (s *Service) ChangePassword(ctx context.Context, userID string, input ChangePasswordInput) error {
	authRecord, err := s.repo.FindPasswordAuthByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if authRecord.PasswordHash == nil || bcrypt.CompareHashAndPassword([]byte(*authRecord.PasswordHash), []byte(strings.TrimSpace(input.CurrentPassword))) != nil {
		return ErrCurrentPasswordInvalid
	}
	if len(strings.TrimSpace(input.NewPassword)) < 8 {
		return ErrWeakPassword
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(input.NewPassword)), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.UpdatePasswordHash(ctx, userID, string(passwordHash))
}

func (s *Service) GetSecurity(ctx context.Context, userID string) (account.SecurityInfo, error) {
	return s.repo.GetSecurityInfo(ctx, userID)
}

func (s *Service) GetQuota(ctx context.Context, userID string) (QuotaSnapshot, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return QuotaSnapshot{}, err
	}

	return QuotaSnapshot{
		Quota:          user.Quota,
		UsedQuota:      user.UsedQuota,
		RemainingQuota: user.Quota,
	}, nil
}

func (s *Service) GetUsage(ctx context.Context, userID, rangeKey string) (account.UsageSnapshot, error) {
	if rangeKey == "" {
		rangeKey = "7d"
	}
	if rangeKey != "7d" && rangeKey != "30d" && rangeKey != "month" {
		return account.UsageSnapshot{}, ErrInvalidUsageRange
	}

	return s.repo.GetUsageSnapshot(ctx, userID, rangeKey, nowUTC())
}
