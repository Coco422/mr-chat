package audit

import "time"

type Result string

const (
	ResultSuccess Result = "success"
	ResultFailed  Result = "failed"
)

type Log struct {
	ID           string         `gorm:"type:uuid;primaryKey"`
	ActorUserID  *string        `gorm:"column:actor_user_id"`
	ActorRole    *string        `gorm:"column:actor_role"`
	Action       string         `gorm:"column:action"`
	ResourceType string         `gorm:"column:resource_type"`
	ResourceID   *string        `gorm:"column:resource_id"`
	TargetUserID *string        `gorm:"column:target_user_id"`
	RequestID    *string        `gorm:"column:request_id"`
	IPAddress    *string        `gorm:"column:ip_address"`
	UserAgent    *string        `gorm:"column:user_agent"`
	Result       Result         `gorm:"column:result"`
	Details      map[string]any `gorm:"column:detail_json;type:jsonb;serializer:json"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
}

func (Log) TableName() string {
	return "audit_logs"
}
