package account

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrPasswordNotFound = errors.New("password auth not found")
)

type Repository struct {
	db *gorm.DB
}

type CreateUserInput struct {
	Username     string
	Email        string
	DisplayName  string
	PasswordHash string
	Role         Role
	Status       UserStatus
	Settings     UserSettings
}

type UpdateUserProfileInput struct {
	DisplayName *string
	AvatarURL   *string
	Settings    *UserSettings
}

type SecurityInfo struct {
	LastLoginAt       *time.Time `json:"last_login_at"`
	PasswordUpdatedAt *time.Time `json:"password_updated_at"`
	HasPassword       bool       `json:"has_password"`
}

type BillingSummary struct {
	RemainingQuota int64 `json:"remaining_quota"`
	ConsumedTotal  int64 `json:"consumed_total"`
	RedeemedTotal  int64 `json:"redeemed_total"`
}

type UsageSummary struct {
	TotalSpentTokens int64 `json:"total_spent_tokens"`
	SpentToday       int64 `json:"spent_today"`
	SpentInRange     int64 `json:"spent_in_range"`
}

type UsageDay struct {
	Date        string `json:"date"`
	SpentTokens int64  `json:"spent_tokens"`
}

type UsageSnapshot struct {
	Summary UsageSummary `json:"summary"`
	Daily   []UsageDay   `json:"daily"`
}

type QuotaLogList struct {
	Items []QuotaLog
	Total int64
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) DB() *gorm.DB {
	return r.db
}

func (r *Repository) UsernameExists(ctx context.Context, username string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&User{}).
		Where("LOWER(username) = ?", normalizeIdentifier(username)).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("count users by username: %w", err)
	}

	return count > 0, nil
}

func (r *Repository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&User{}).
		Where("LOWER(email) = ?", normalizeIdentifier(email)).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("count users by email: %w", err)
	}

	return count > 0, nil
}

func (r *Repository) CreateUserWithPassword(ctx context.Context, input CreateUserInput) (*User, error) {
	now := time.Now().UTC()
	user := &User{
		ID:          uuid.NewString(),
		Username:    normalizeIdentifier(input.Username),
		Email:       normalizeIdentifier(input.Email),
		DisplayName: strings.TrimSpace(input.DisplayName),
		Role:        defaultRole(input.Role),
		Status:      defaultStatus(input.Status),
		Settings:    input.Settings,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}

	passwordHash := input.PasswordHash
	authRecord := &Auth{
		ID:           uuid.NewString(),
		UserID:       user.ID,
		AuthType:     AuthTypePassword,
		PasswordHash: &passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return fmt.Errorf("create user: %w", err)
		}

		if err := tx.Create(authRecord).Error; err != nil {
			return fmt.Errorf("create auth: %w", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *Repository) FindPasswordAuthByIdentifier(ctx context.Context, identifier string) (*User, *Auth, error) {
	var user User
	normalized := normalizeIdentifier(identifier)
	if err := r.db.WithContext(ctx).
		Where("LOWER(username) = ? OR LOWER(email) = ?", normalized, normalized).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrUserNotFound
		}

		return nil, nil, fmt.Errorf("find user by identifier: %w", err)
	}

	authRecord, err := r.FindPasswordAuthByUserID(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}

	return &user, authRecord, nil
}

func (r *Repository) FindPasswordAuthByUserID(ctx context.Context, userID string) (*Auth, error) {
	var authRecord Auth
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND auth_type = ?", userID, AuthTypePassword).
		First(&authRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPasswordNotFound
		}

		return nil, fmt.Errorf("find password auth by user id: %w", err)
	}

	return &authRecord, nil
}

func (r *Repository) GetUserByID(ctx context.Context, userID string) (*User, error) {
	var user User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &user, nil
}

func (r *Repository) UpdateLastLogin(ctx context.Context, userID string, at time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&User{}).
			Where("id = ?", userID).
			Updates(map[string]any{
				"last_login_at": at,
				"updated_at":    at,
			}).Error; err != nil {
			return fmt.Errorf("update user last login: %w", err)
		}

		if err := tx.Model(&Auth{}).
			Where("user_id = ? AND auth_type = ?", userID, AuthTypePassword).
			Updates(map[string]any{
				"last_login_at": at,
				"updated_at":    at,
			}).Error; err != nil {
			return fmt.Errorf("update auth last login: %w", err)
		}

		return nil
	})
}

func (r *Repository) UpdateUserProfile(ctx context.Context, userID string, input UpdateUserProfileInput) (*User, error) {
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if input.DisplayName != nil {
		user.DisplayName = strings.TrimSpace(*input.DisplayName)
	}

	if input.AvatarURL != nil {
		avatarURL := strings.TrimSpace(*input.AvatarURL)
		if avatarURL == "" {
			user.AvatarURL = nil
		} else {
			user.AvatarURL = &avatarURL
		}
	}

	if input.Settings != nil {
		user.Settings = *input.Settings
	}

	user.UpdatedAt = time.Now().UTC()
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return nil, fmt.Errorf("update user profile: %w", err)
	}

	return user, nil
}

func (r *Repository) UpdatePasswordHash(ctx context.Context, userID, passwordHash string) error {
	updates := map[string]any{
		"password_hash": passwordHash,
		"updated_at":    time.Now().UTC(),
	}

	if err := r.db.WithContext(ctx).
		Model(&Auth{}).
		Where("user_id = ? AND auth_type = ?", userID, AuthTypePassword).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("update password hash: %w", err)
	}

	return nil
}

func (r *Repository) GetSecurityInfo(ctx context.Context, userID string) (SecurityInfo, error) {
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return SecurityInfo{}, err
	}

	authRecord, err := r.FindPasswordAuthByUserID(ctx, userID)
	if err != nil && !errors.Is(err, ErrPasswordNotFound) {
		return SecurityInfo{}, err
	}

	info := SecurityInfo{
		LastLoginAt: user.LastLoginAt,
	}
	if authRecord != nil {
		info.HasPassword = authRecord.PasswordHash != nil && *authRecord.PasswordHash != ""
		info.PasswordUpdatedAt = &authRecord.UpdatedAt
	}

	return info, nil
}

func (r *Repository) GetBillingSummary(ctx context.Context, userID string) (BillingSummary, error) {
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return BillingSummary{}, err
	}

	var redeemedTotal int64
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(delta_quota), 0)
		FROM quota_logs
		WHERE user_id = ? AND log_type = ?
	`, userID, QuotaLogTypeRedeem).Scan(&redeemedTotal).Error; err != nil {
		return BillingSummary{}, fmt.Errorf("query redeemed total: %w", err)
	}

	consumedTotal := user.UsedQuota
	return BillingSummary{
		RemainingQuota: user.Quota,
		ConsumedTotal:  consumedTotal,
		RedeemedTotal:  redeemedTotal,
	}, nil
}

func (r *Repository) GetUsageSnapshot(ctx context.Context, userID, rangeKey string, now time.Time) (UsageSnapshot, error) {
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return UsageSnapshot{}, err
	}

	start, end := usageWindow(rangeKey, now.UTC())
	daily := make([]UsageDay, 0)
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			TO_CHAR(DATE(created_at AT TIME ZONE 'UTC'), 'YYYY-MM-DD') AS date,
			COALESCE(SUM(CASE WHEN log_type = 'final_charge' AND delta_quota < 0 THEN -delta_quota ELSE 0 END), 0) AS spent_tokens
		FROM quota_logs
		WHERE user_id = ? AND created_at >= ? AND created_at < ?
		GROUP BY 1
		ORDER BY 1
	`, userID, start, end).Scan(&daily).Error; err != nil {
		return UsageSnapshot{}, fmt.Errorf("query usage daily: %w", err)
	}

	var spentToday int64
	todayStart := now.UTC().Truncate(24 * time.Hour)
	tomorrowStart := todayStart.Add(24 * time.Hour)
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(SUM(CASE WHEN log_type = 'final_charge' AND delta_quota < 0 THEN -delta_quota ELSE 0 END), 0)
		FROM quota_logs
		WHERE user_id = ? AND created_at >= ? AND created_at < ?
	`, userID, todayStart, tomorrowStart).Scan(&spentToday).Error; err != nil {
		return UsageSnapshot{}, fmt.Errorf("query usage today: %w", err)
	}

	var spentInRange int64
	for _, point := range daily {
		spentInRange += point.SpentTokens
	}

	return UsageSnapshot{
		Summary: UsageSummary{
			TotalSpentTokens: user.UsedQuota,
			SpentToday:       spentToday,
			SpentInRange:     spentInRange,
		},
		Daily: daily,
	}, nil
}

func (r *Repository) ListQuotaLogs(ctx context.Context, userID string, page, pageSize int, logType QuotaLogType) (QuotaLogList, error) {
	query := r.db.WithContext(ctx).Model(&QuotaLog{}).Where("user_id = ?", userID)
	if logType != "" {
		query = query.Where("log_type = ?", logType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return QuotaLogList{}, fmt.Errorf("count quota logs: %w", err)
	}

	var items []QuotaLog
	if err := query.Order("created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&items).Error; err != nil {
		return QuotaLogList{}, fmt.Errorf("list quota logs: %w", err)
	}

	return QuotaLogList{
		Items: items,
		Total: total,
	}, nil
}

func defaultRole(role Role) Role {
	if role == "" {
		return RoleUser
	}

	return role
}

func defaultStatus(status UserStatus) UserStatus {
	if status == "" {
		return UserStatusActive
	}

	return status
}

func normalizeIdentifier(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func usageWindow(rangeKey string, now time.Time) (time.Time, time.Time) {
	today := now.Truncate(24 * time.Hour)
	switch rangeKey {
	case "30d":
		return today.AddDate(0, 0, -29), today.Add(24 * time.Hour)
	case "month":
		firstDay := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
		return firstDay, today.Add(24 * time.Hour)
	default:
		return today.AddDate(0, 0, -6), today.Add(24 * time.Hour)
	}
}
