package limits

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrPolicyNotFound = errors.New("limit policy not found")

type Repository struct {
	db *gorm.DB
}

type PolicyUpsertInput struct {
	ModelID              *string
	HourRequestLimit     *int64
	WeekRequestLimit     *int64
	LifetimeRequestLimit *int64
	HourTokenLimit       *int64
	WeekTokenLimit       *int64
	LifetimeTokenLimit   *int64
	Status               PolicyStatus
}

type AdjustmentCreateInput struct {
	UserID      string
	ModelID     *string
	MetricType  MetricType
	WindowType  WindowType
	Delta       int64
	ExpiresAt   *time.Time
	Reason      *string
	ActorUserID *string
}

type AdjustmentListFilter struct {
	UserID   string
	ModelID  *string
	Page     int
	PageSize int
}

type AdjustmentListResult struct {
	Items []UserLimitAdjustment
	Total int64
}

type RequestUsageFilter struct {
	UserID  string
	ModelID *string
	Now     time.Time
}

type RequestLogCreateInput struct {
	RequestID        string
	UserID           string
	UserGroupID      *string
	ConversationID   *string
	MessageID        *string
	ModelID          *string
	ChannelID        *string
	PromptTokens     int64
	CompletionTokens int64
	TotalTokens      int64
	BilledQuota      int64
	Status           RequestLogStatus
	ErrorCode        *string
	StartedAt        time.Time
	CompletedAt      *time.Time
	Metadata         map[string]any
}

type RequestLogUpdateInput struct {
	PromptTokens     *int64
	CompletionTokens *int64
	TotalTokens      *int64
	BilledQuota      *int64
	Status           *RequestLogStatus
	ErrorCode        *string
	CompletedAt      *time.Time
	Metadata         map[string]any
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) DB() *gorm.DB {
	return r.db
}

func (r *Repository) ListPoliciesByUserGroup(ctx context.Context, userGroupID string) ([]UserGroupModelLimitPolicy, error) {
	var items []UserGroupModelLimitPolicy
	if err := r.db.WithContext(ctx).
		Where("user_group_id = ?", userGroupID).
		Order("CASE WHEN model_id IS NULL THEN 0 ELSE 1 END ASC, created_at ASC").
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list limit policies: %w", err)
	}

	return items, nil
}

func (r *Repository) ReplacePoliciesByUserGroup(ctx context.Context, userGroupID string, inputs []PolicyUpsertInput) ([]UserGroupModelLimitPolicy, error) {
	return r.ReplacePoliciesByUserGroupWithDB(ctx, r.db, userGroupID, inputs)
}

func (r *Repository) ReplacePoliciesByUserGroupWithDB(ctx context.Context, db *gorm.DB, userGroupID string, inputs []PolicyUpsertInput) ([]UserGroupModelLimitPolicy, error) {
	if db == nil {
		db = r.db
	}

	now := time.Now().UTC()
	items := make([]UserGroupModelLimitPolicy, 0, len(inputs))
	for _, input := range inputs {
		items = append(items, UserGroupModelLimitPolicy{
			ID:                   uuid.NewString(),
			UserGroupID:          userGroupID,
			ModelID:              sanitizeOptionalString(input.ModelID),
			HourRequestLimit:     input.HourRequestLimit,
			WeekRequestLimit:     input.WeekRequestLimit,
			LifetimeRequestLimit: input.LifetimeRequestLimit,
			HourTokenLimit:       input.HourTokenLimit,
			WeekTokenLimit:       input.WeekTokenLimit,
			LifetimeTokenLimit:   input.LifetimeTokenLimit,
			Status:               defaultPolicyStatus(input.Status),
			CreatedAt:            now,
			UpdatedAt:            now,
		})
	}

	if err := db.WithContext(ctx).Where("user_group_id = ?", userGroupID).Delete(&UserGroupModelLimitPolicy{}).Error; err != nil {
		return nil, fmt.Errorf("clear limit policies: %w", err)
	}

	if len(items) == 0 {
		return []UserGroupModelLimitPolicy{}, nil
	}

	if err := db.WithContext(ctx).Create(&items).Error; err != nil {
		return nil, fmt.Errorf("create limit policies: %w", err)
	}

	return items, nil
}

func (r *Repository) ResolvePolicy(ctx context.Context, userGroupID string, modelID *string) (*UserGroupModelLimitPolicy, error) {
	query := r.db.WithContext(ctx).
		Model(&UserGroupModelLimitPolicy{}).
		Where("user_group_id = ? AND status = ?", userGroupID, PolicyStatusActive)

	var item UserGroupModelLimitPolicy
	if modelID != nil && strings.TrimSpace(*modelID) != "" {
		if err := query.
			Where("model_id = ?", strings.TrimSpace(*modelID)).
			First(&item).Error; err == nil {
			return &item, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("resolve model override policy: %w", err)
		}
	}

	if err := query.Where("model_id IS NULL").First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPolicyNotFound
		}
		return nil, fmt.Errorf("resolve default policy: %w", err)
	}

	return &item, nil
}

func (r *Repository) CreateAdjustment(ctx context.Context, input AdjustmentCreateInput) (*UserLimitAdjustment, error) {
	return r.CreateAdjustmentWithDB(ctx, r.db, input)
}

func (r *Repository) CreateAdjustmentWithDB(ctx context.Context, db *gorm.DB, input AdjustmentCreateInput) (*UserLimitAdjustment, error) {
	if db == nil {
		db = r.db
	}

	item := &UserLimitAdjustment{
		ID:          uuid.NewString(),
		UserID:      input.UserID,
		ModelID:     sanitizeOptionalString(input.ModelID),
		MetricType:  input.MetricType,
		WindowType:  input.WindowType,
		Delta:       input.Delta,
		ExpiresAt:   input.ExpiresAt,
		Reason:      sanitizeOptionalString(input.Reason),
		ActorUserID: sanitizeOptionalString(input.ActorUserID),
		CreatedAt:   time.Now().UTC(),
	}

	if err := db.WithContext(ctx).Create(item).Error; err != nil {
		return nil, fmt.Errorf("create user limit adjustment: %w", err)
	}

	return item, nil
}

func (r *Repository) ListAdjustments(ctx context.Context, filter AdjustmentListFilter) (AdjustmentListResult, error) {
	query := r.db.WithContext(ctx).Model(&UserLimitAdjustment{}).Where("user_id = ?", filter.UserID)
	if filter.ModelID != nil && strings.TrimSpace(*filter.ModelID) != "" {
		query = query.Where("model_id = ?", strings.TrimSpace(*filter.ModelID))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return AdjustmentListResult{}, fmt.Errorf("count user limit adjustments: %w", err)
	}

	var items []UserLimitAdjustment
	if err := query.Order("created_at DESC").
		Limit(filter.PageSize).
		Offset((filter.Page - 1) * filter.PageSize).
		Find(&items).Error; err != nil {
		return AdjustmentListResult{}, fmt.Errorf("list user limit adjustments: %w", err)
	}

	return AdjustmentListResult{
		Items: items,
		Total: total,
	}, nil
}

func (r *Repository) GetRequestUsage(ctx context.Context, filter RequestUsageFilter) (UsageCounters, error) {
	now := filter.Now.UTC()
	hourStart := now.Add(-1 * time.Hour)
	weekStart := now.Add(-7 * 24 * time.Hour)

	lifetime, err := r.aggregateUsage(ctx, filter.UserID, filter.ModelID, nil, now)
	if err != nil {
		return UsageCounters{}, err
	}
	hour, err := r.aggregateUsage(ctx, filter.UserID, filter.ModelID, &hourStart, now)
	if err != nil {
		return UsageCounters{}, err
	}
	week, err := r.aggregateUsage(ctx, filter.UserID, filter.ModelID, &weekStart, now)
	if err != nil {
		return UsageCounters{}, err
	}

	return UsageCounters{
		Hour:     hour,
		Week:     week,
		Lifetime: lifetime,
	}, nil
}

func (r *Repository) GetAdjustmentUsage(ctx context.Context, userID string, modelID *string, now time.Time) (UsageCounters, error) {
	return UsageCounters{
		Hour: UsageCounter{
			Requests: r.sumAdjustments(ctx, userID, modelID, MetricTypeRequestCount, WindowTypeRollingHour, now),
			Tokens:   r.sumAdjustments(ctx, userID, modelID, MetricTypeTotalTokens, WindowTypeRollingHour, now),
		},
		Week: UsageCounter{
			Requests: r.sumAdjustments(ctx, userID, modelID, MetricTypeRequestCount, WindowTypeRollingWeek, now),
			Tokens:   r.sumAdjustments(ctx, userID, modelID, MetricTypeTotalTokens, WindowTypeRollingWeek, now),
		},
		Lifetime: UsageCounter{
			Requests: r.sumAdjustments(ctx, userID, modelID, MetricTypeRequestCount, WindowTypeLifetime, now),
			Tokens:   r.sumAdjustments(ctx, userID, modelID, MetricTypeTotalTokens, WindowTypeLifetime, now),
		},
	}, nil
}

func (r *Repository) CreateRequestLog(ctx context.Context, input RequestLogCreateInput) (*LLMRequestLog, error) {
	return r.CreateRequestLogWithDB(ctx, r.db, input)
}

func (r *Repository) CreateRequestLogWithDB(ctx context.Context, db *gorm.DB, input RequestLogCreateInput) (*LLMRequestLog, error) {
	if db == nil {
		db = r.db
	}

	item := &LLMRequestLog{
		ID:               uuid.NewString(),
		RequestID:        strings.TrimSpace(input.RequestID),
		UserID:           input.UserID,
		UserGroupID:      sanitizeOptionalString(input.UserGroupID),
		ConversationID:   sanitizeOptionalString(input.ConversationID),
		MessageID:        sanitizeOptionalString(input.MessageID),
		ModelID:          sanitizeOptionalString(input.ModelID),
		ChannelID:        sanitizeOptionalString(input.ChannelID),
		PromptTokens:     input.PromptTokens,
		CompletionTokens: input.CompletionTokens,
		TotalTokens:      input.TotalTokens,
		BilledQuota:      input.BilledQuota,
		Status:           defaultRequestLogStatus(input.Status),
		ErrorCode:        sanitizeOptionalString(input.ErrorCode),
		StartedAt:        input.StartedAt.UTC(),
		CompletedAt:      input.CompletedAt,
		Metadata:         nonNilMap(input.Metadata),
	}

	if item.StartedAt.IsZero() {
		item.StartedAt = time.Now().UTC()
	}

	if err := db.WithContext(ctx).Create(item).Error; err != nil {
		return nil, fmt.Errorf("create llm request log: %w", err)
	}

	return item, nil
}

func (r *Repository) UpdateRequestLogByRequestID(ctx context.Context, requestID string, input RequestLogUpdateInput) (*LLMRequestLog, error) {
	return r.UpdateRequestLogByRequestIDWithDB(ctx, r.db, requestID, input)
}

func (r *Repository) UpdateRequestLogByRequestIDWithDB(ctx context.Context, db *gorm.DB, requestID string, input RequestLogUpdateInput) (*LLMRequestLog, error) {
	if db == nil {
		db = r.db
	}

	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return nil, fmt.Errorf("request_id is required")
	}

	var item LLMRequestLog
	if err := db.WithContext(ctx).Where("request_id = ?", requestID).First(&item).Error; err != nil {
		return nil, fmt.Errorf("get llm request log: %w", err)
	}

	if input.PromptTokens != nil {
		item.PromptTokens = *input.PromptTokens
	}
	if input.CompletionTokens != nil {
		item.CompletionTokens = *input.CompletionTokens
	}
	if input.TotalTokens != nil {
		item.TotalTokens = *input.TotalTokens
	}
	if input.BilledQuota != nil {
		item.BilledQuota = *input.BilledQuota
	}
	if input.Status != nil {
		item.Status = *input.Status
	}
	if input.ErrorCode != nil {
		item.ErrorCode = sanitizeOptionalString(input.ErrorCode)
	}
	if input.CompletedAt != nil {
		completedAt := input.CompletedAt.UTC()
		item.CompletedAt = &completedAt
	}
	if input.Metadata != nil {
		item.Metadata = nonNilMap(input.Metadata)
	}

	if err := db.WithContext(ctx).Save(&item).Error; err != nil {
		return nil, fmt.Errorf("update llm request log: %w", err)
	}

	return &item, nil
}

func (r *Repository) aggregateUsage(ctx context.Context, userID string, modelID *string, start *time.Time, end time.Time) (UsageCounter, error) {
	query := r.db.WithContext(ctx).
		Model(&LLMRequestLog{}).
		Where("user_id = ?", userID).
		Where("status IN ?", []RequestLogStatus{
			RequestLogStatusCompleted,
			RequestLogStatusFailed,
			RequestLogStatusCancelled,
		}).
		Where("started_at <= ?", end)

	if modelID != nil && strings.TrimSpace(*modelID) != "" {
		query = query.Where("model_id = ?", strings.TrimSpace(*modelID))
	}
	if start != nil {
		query = query.Where("started_at >= ?", start.UTC())
	}

	var row struct {
		Requests int64 `gorm:"column:requests"`
		Tokens   int64 `gorm:"column:tokens"`
	}
	if err := query.Select("COUNT(*) AS requests, COALESCE(SUM(total_tokens), 0) AS tokens").Scan(&row).Error; err != nil {
		return UsageCounter{}, fmt.Errorf("aggregate llm request logs: %w", err)
	}

	return UsageCounter{
		Requests: row.Requests,
		Tokens:   row.Tokens,
	}, nil
}

func (r *Repository) sumAdjustments(ctx context.Context, userID string, modelID *string, metricType MetricType, windowType WindowType, now time.Time) int64 {
	query := r.db.WithContext(ctx).
		Model(&UserLimitAdjustment{}).
		Where("user_id = ? AND metric_type = ? AND window_type = ?", userID, metricType, windowType).
		Where("expires_at IS NULL OR expires_at > ?", now.UTC())

	if modelID != nil && strings.TrimSpace(*modelID) != "" {
		query = query.Where("(model_id = ? OR model_id IS NULL)", strings.TrimSpace(*modelID))
	} else {
		query = query.Where("model_id IS NULL")
	}

	var total int64
	if err := query.Select("COALESCE(SUM(delta), 0)").Scan(&total).Error; err != nil {
		return 0
	}

	return total
}

func defaultPolicyStatus(status PolicyStatus) PolicyStatus {
	if status == "" {
		return PolicyStatusActive
	}
	return status
}

func defaultRequestLogStatus(status RequestLogStatus) RequestLogStatus {
	if status == "" {
		return RequestLogStatusPending
	}
	return status
}

func sanitizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func nonNilMap(value map[string]any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	return value
}
