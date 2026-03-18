package catalog

import "time"

type UpstreamStatus string

const (
	UpstreamStatusActive      UpstreamStatus = "active"
	UpstreamStatusDisabled    UpstreamStatus = "disabled"
	UpstreamStatusMaintenance UpstreamStatus = "maintenance"
)

type ModelStatus string

const (
	ModelStatusActive   ModelStatus = "active"
	ModelStatusDisabled ModelStatus = "disabled"
)

type RouteBindingStatus string

const (
	RouteBindingStatusActive   RouteBindingStatus = "active"
	RouteBindingStatusDisabled RouteBindingStatus = "disabled"
)

type ChannelStatus string

const (
	ChannelStatusActive   ChannelStatus = "active"
	ChannelStatusDisabled ChannelStatus = "disabled"
)

type Upstream struct {
	ID               string         `gorm:"type:uuid;primaryKey"`
	Name             string         `gorm:"column:name"`
	ProviderType     string         `gorm:"column:provider_type"`
	BaseURL          string         `gorm:"column:base_url"`
	AuthType         string         `gorm:"column:auth_type"`
	AuthConfig       map[string]any `gorm:"column:auth_config_encrypted;type:jsonb;serializer:json"`
	Status           UpstreamStatus `gorm:"column:status"`
	TimeoutSeconds   int            `gorm:"column:timeout_seconds"`
	CooldownSeconds  int            `gorm:"column:cooldown_seconds"`
	FailureThreshold int            `gorm:"column:failure_threshold"`
	Metadata         map[string]any `gorm:"column:metadata_json;type:jsonb;serializer:json"`
	CreatedAt        time.Time      `gorm:"column:created_at"`
	UpdatedAt        time.Time      `gorm:"column:updated_at"`
}

func (Upstream) TableName() string {
	return "upstreams"
}

type Channel struct {
	ID            string         `gorm:"type:uuid;primaryKey"`
	Name          string         `gorm:"column:name"`
	Description   *string        `gorm:"column:description"`
	Status        ChannelStatus  `gorm:"column:status"`
	BillingConfig map[string]any `gorm:"column:billing_config_json;type:jsonb;serializer:json"`
	Metadata      map[string]any `gorm:"column:metadata_json;type:jsonb;serializer:json"`
	CreatedAt     time.Time      `gorm:"column:created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at"`
}

func (Channel) TableName() string {
	return "channels"
}

type Model struct {
	ID                  string         `gorm:"type:uuid;primaryKey"`
	ModelKey            string         `gorm:"column:model_key"`
	DisplayName         string         `gorm:"column:display_name"`
	ProviderType        string         `gorm:"column:provider_type"`
	ContextLength       int            `gorm:"column:context_length"`
	MaxOutputTokens     *int           `gorm:"column:max_output_tokens"`
	Pricing             map[string]any `gorm:"column:pricing_json;type:jsonb;serializer:json"`
	Capabilities        map[string]any `gorm:"column:capabilities_json;type:jsonb;serializer:json"`
	VisibleUserGroupIDs []string       `gorm:"column:visible_user_group_ids_json;type:jsonb;serializer:json"`
	Status              ModelStatus    `gorm:"column:status"`
	Metadata            map[string]any `gorm:"column:metadata_json;type:jsonb;serializer:json"`
	CreatedAt           time.Time      `gorm:"column:created_at"`
	UpdatedAt           time.Time      `gorm:"column:updated_at"`
}

func (Model) TableName() string {
	return "models"
}

type ModelRouteBinding struct {
	ID         string             `gorm:"type:uuid;primaryKey"`
	ModelID    string             `gorm:"column:model_id"`
	ChannelID  *string            `gorm:"column:channel_id"`
	UpstreamID string             `gorm:"column:upstream_id"`
	Priority   int                `gorm:"column:priority"`
	Status     RouteBindingStatus `gorm:"column:status"`
	CreatedAt  time.Time          `gorm:"column:created_at"`
	UpdatedAt  time.Time          `gorm:"column:updated_at"`
}

func (ModelRouteBinding) TableName() string {
	return "model_route_bindings"
}

type RouteBindingInput struct {
	ChannelID  *string `json:"channel_id"`
	UpstreamID string  `json:"upstream_id"`
	Priority   int     `json:"priority"`
	Status     string  `json:"status"`
}

type CreateChannelInput struct {
	Name          string
	Description   *string
	Status        string
	BillingConfig map[string]any
	Metadata      map[string]any
}

type UpdateChannelInput struct {
	Name          *string
	Description   *string
	Status        *string
	BillingConfig map[string]any
	Metadata      map[string]any
}

type CreateUpstreamInput struct {
	Name             string
	ProviderType     string
	BaseURL          string
	AuthType         string
	AuthConfig       map[string]any
	Status           string
	TimeoutSeconds   int
	CooldownSeconds  int
	FailureThreshold int
	Metadata         map[string]any
}

type UpdateUpstreamInput struct {
	Name             *string
	ProviderType     *string
	BaseURL          *string
	AuthType         *string
	AuthConfig       map[string]any
	Status           *string
	TimeoutSeconds   *int
	CooldownSeconds  *int
	FailureThreshold *int
	Metadata         map[string]any
}

type CreateModelInput struct {
	ModelKey            string
	DisplayName         string
	ProviderType        string
	ContextLength       int
	MaxOutputTokens     *int
	Pricing             map[string]any
	Capabilities        map[string]any
	VisibleUserGroupIDs []string
	Status              string
	Metadata            map[string]any
	RouteBindings       []RouteBindingInput
}

type UpdateModelInput struct {
	ModelKey            *string
	DisplayName         *string
	ProviderType        *string
	ContextLength       *int
	MaxOutputTokens     *int
	Pricing             map[string]any
	Capabilities        map[string]any
	VisibleUserGroupIDs []string
	Status              *string
	Metadata            map[string]any
	RouteBindings       []RouteBindingInput
}

type ModelWithBindings struct {
	Model         Model
	RouteBindings []ModelRouteBinding
}
