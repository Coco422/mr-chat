package limits

import "time"

type PolicyStatus string

const (
	PolicyStatusActive   PolicyStatus = "active"
	PolicyStatusDisabled PolicyStatus = "disabled"
)

type MetricType string

const (
	MetricTypeRequestCount MetricType = "request_count"
	MetricTypeTotalTokens  MetricType = "total_tokens"
)

type WindowType string

const (
	WindowTypeRollingHour WindowType = "rolling_hour"
	WindowTypeRollingWeek WindowType = "rolling_week"
	WindowTypeLifetime    WindowType = "lifetime"
)

type RequestLogStatus string

const (
	RequestLogStatusPending   RequestLogStatus = "pending"
	RequestLogStatusCompleted RequestLogStatus = "completed"
	RequestLogStatusFailed    RequestLogStatus = "failed"
	RequestLogStatusCancelled RequestLogStatus = "cancelled"
	RequestLogStatusRejected  RequestLogStatus = "rejected"
)

type UserGroupModelLimitPolicy struct {
	ID                   string       `gorm:"type:uuid;primaryKey"`
	UserGroupID          string       `gorm:"column:user_group_id"`
	ModelID              *string      `gorm:"column:model_id"`
	HourRequestLimit     *int64       `gorm:"column:hour_request_limit"`
	WeekRequestLimit     *int64       `gorm:"column:week_request_limit"`
	LifetimeRequestLimit *int64       `gorm:"column:lifetime_request_limit"`
	HourTokenLimit       *int64       `gorm:"column:hour_token_limit"`
	WeekTokenLimit       *int64       `gorm:"column:week_token_limit"`
	LifetimeTokenLimit   *int64       `gorm:"column:lifetime_token_limit"`
	Status               PolicyStatus `gorm:"column:status"`
	CreatedAt            time.Time    `gorm:"column:created_at"`
	UpdatedAt            time.Time    `gorm:"column:updated_at"`
}

func (UserGroupModelLimitPolicy) TableName() string {
	return "user_group_model_limit_policies"
}

type UserLimitAdjustment struct {
	ID          string     `gorm:"type:uuid;primaryKey"`
	UserID      string     `gorm:"column:user_id"`
	ModelID     *string    `gorm:"column:model_id"`
	MetricType  MetricType `gorm:"column:metric_type"`
	WindowType  WindowType `gorm:"column:window_type"`
	Delta       int64      `gorm:"column:delta"`
	ExpiresAt   *time.Time `gorm:"column:expires_at"`
	Reason      *string    `gorm:"column:reason"`
	ActorUserID *string    `gorm:"column:actor_user_id"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
}

func (UserLimitAdjustment) TableName() string {
	return "user_limit_adjustments"
}

type LLMRequestLog struct {
	ID               string           `gorm:"type:uuid;primaryKey"`
	RequestID        string           `gorm:"column:request_id"`
	UserID           string           `gorm:"column:user_id"`
	UserGroupID      *string          `gorm:"column:user_group_id"`
	ConversationID   *string          `gorm:"column:conversation_id"`
	MessageID        *string          `gorm:"column:message_id"`
	ModelID          *string          `gorm:"column:model_id"`
	ChannelID        *string          `gorm:"column:channel_id"`
	PromptTokens     int64            `gorm:"column:prompt_tokens"`
	CompletionTokens int64            `gorm:"column:completion_tokens"`
	TotalTokens      int64            `gorm:"column:total_tokens"`
	BilledQuota      int64            `gorm:"column:billed_quota"`
	Status           RequestLogStatus `gorm:"column:status"`
	ErrorCode        *string          `gorm:"column:error_code"`
	StartedAt        time.Time        `gorm:"column:started_at"`
	CompletedAt      *time.Time       `gorm:"column:completed_at"`
	Metadata         map[string]any   `gorm:"column:metadata_json;type:jsonb;serializer:json"`
}

func (LLMRequestLog) TableName() string {
	return "llm_request_logs"
}

type UsageCounter struct {
	Requests int64 `json:"requests"`
	Tokens   int64 `json:"tokens"`
}

type UsageCounters struct {
	Hour     UsageCounter `json:"hour"`
	Week     UsageCounter `json:"week"`
	Lifetime UsageCounter `json:"lifetime"`
}

type EffectivePolicy struct {
	Source               string  `json:"source"`
	ModelID              *string `json:"model_id"`
	HourRequestLimit     *int64  `json:"hour_request_limit"`
	WeekRequestLimit     *int64  `json:"week_request_limit"`
	LifetimeRequestLimit *int64  `json:"lifetime_request_limit"`
	HourTokenLimit       *int64  `json:"hour_token_limit"`
	WeekTokenLimit       *int64  `json:"week_token_limit"`
	LifetimeTokenLimit   *int64  `json:"lifetime_token_limit"`
}
