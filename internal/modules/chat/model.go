package chat

import "time"

type ConversationStatus string

const (
	ConversationStatusActive   ConversationStatus = "active"
	ConversationStatusArchived ConversationStatus = "archived"
	ConversationStatusDeleted  ConversationStatus = "deleted"
)

type MessageStatus string

const (
	MessageStatusPending   MessageStatus = "pending"
	MessageStatusStreaming MessageStatus = "streaming"
	MessageStatusCompleted MessageStatus = "completed"
	MessageStatusFailed    MessageStatus = "failed"
	MessageStatusCancelled MessageStatus = "cancelled"
)

type Conversation struct {
	ID            string             `gorm:"type:uuid;primaryKey"`
	UserID        string             `gorm:"column:user_id"`
	Title         string             `gorm:"column:title"`
	ModelID       *string            `gorm:"column:model_id"`
	Status        ConversationStatus `gorm:"column:status"`
	MessageCount  int                `gorm:"column:message_count"`
	LastMessageAt *time.Time         `gorm:"column:last_message_at"`
	Metadata      map[string]any     `gorm:"column:metadata_json;type:jsonb;serializer:json"`
	CreatedAt     time.Time          `gorm:"column:created_at"`
	UpdatedAt     time.Time          `gorm:"column:updated_at"`
	DeletedAt     *time.Time         `gorm:"column:deleted_at"`
}

func (Conversation) TableName() string {
	return "conversations"
}

type Message struct {
	ID               string         `gorm:"type:uuid;primaryKey"`
	ConversationID   string         `gorm:"column:conversation_id"`
	UserID           string         `gorm:"column:user_id"`
	ModelID          *string        `gorm:"column:model_id"`
	UpstreamID       *string        `gorm:"column:upstream_id"`
	RequestID        *string        `gorm:"column:request_id"`
	Role             string         `gorm:"column:role"`
	Content          string         `gorm:"column:content"`
	ReasoningContent *string        `gorm:"column:reasoning_content"`
	Status           MessageStatus  `gorm:"column:status"`
	FinishReason     *string        `gorm:"column:finish_reason"`
	Usage            map[string]any `gorm:"column:usage_json;type:jsonb;serializer:json"`
	ErrorCode        *string        `gorm:"column:error_code"`
	Metadata         map[string]any `gorm:"column:metadata_json;type:jsonb;serializer:json"`
	CreatedAt        time.Time      `gorm:"column:created_at"`
	UpdatedAt        time.Time      `gorm:"column:updated_at"`
	DeletedAt        *time.Time     `gorm:"column:deleted_at"`
}

func (Message) TableName() string {
	return "messages"
}
