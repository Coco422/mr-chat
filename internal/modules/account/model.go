package account

import (
	"time"

	"gorm.io/gorm"
)

type Role string

const (
	RoleRoot  Role = "root"
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
	UserStatusPending  UserStatus = "pending"
)

type AuthType string

const (
	AuthTypePassword AuthType = "password"
	AuthTypeOAuth    AuthType = "oauth"
)

type GroupStatus string

const (
	GroupStatusActive   GroupStatus = "active"
	GroupStatusDisabled GroupStatus = "disabled"
)

type GroupMemberRole string

const (
	GroupMemberRoleOwner  GroupMemberRole = "owner"
	GroupMemberRoleAdmin  GroupMemberRole = "admin"
	GroupMemberRoleMember GroupMemberRole = "member"
)

type QuotaLogType string

const (
	QuotaLogTypePreDeduct   QuotaLogType = "pre_deduct"
	QuotaLogTypeFinalCharge QuotaLogType = "final_charge"
	QuotaLogTypeRefund      QuotaLogType = "refund"
	QuotaLogTypeRedeem      QuotaLogType = "redeem"
	QuotaLogTypeAdminAdjust QuotaLogType = "admin_adjust"
)

type UserSettings struct {
	Timezone string `json:"timezone"`
	Locale   string `json:"locale"`
}

type User struct {
	ID             string         `gorm:"type:uuid;primaryKey"`
	Username       string         `gorm:"column:username"`
	Email          string         `gorm:"column:email"`
	DisplayName    string         `gorm:"column:display_name"`
	AvatarURL      *string        `gorm:"column:avatar_url"`
	Role           Role           `gorm:"column:role"`
	Status         UserStatus     `gorm:"column:status"`
	Quota          int64          `gorm:"column:quota"`
	UsedQuota      int64          `gorm:"column:used_quota"`
	PrimaryGroupID *string        `gorm:"column:primary_group_id"`
	Settings       UserSettings   `gorm:"column:settings_json;type:jsonb;serializer:json"`
	LastLoginAt    *time.Time     `gorm:"column:last_login_at"`
	CreatedAt      time.Time      `gorm:"column:created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (User) TableName() string {
	return "users"
}

type Auth struct {
	ID              string     `gorm:"type:uuid;primaryKey"`
	UserID          string     `gorm:"column:user_id"`
	AuthType        AuthType   `gorm:"column:auth_type"`
	Provider        *string    `gorm:"column:provider"`
	ProviderSubject *string    `gorm:"column:provider_subject"`
	PasswordHash    *string    `gorm:"column:password_hash"`
	VerifiedAt      *time.Time `gorm:"column:verified_at"`
	LastLoginAt     *time.Time `gorm:"column:last_login_at"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
}

func (Auth) TableName() string {
	return "auths"
}

type Group struct {
	ID              string      `gorm:"type:uuid;primaryKey"`
	Name            string      `gorm:"column:name"`
	Description     *string     `gorm:"column:description"`
	Status          GroupStatus `gorm:"column:status"`
	PermissionsJSON string      `gorm:"column:permissions_json"`
	CreatedAt       time.Time   `gorm:"column:created_at"`
	UpdatedAt       time.Time   `gorm:"column:updated_at"`
}

func (Group) TableName() string {
	return "groups"
}

type GroupMember struct {
	ID         string          `gorm:"type:uuid;primaryKey"`
	GroupID    string          `gorm:"column:group_id"`
	UserID     string          `gorm:"column:user_id"`
	MemberRole GroupMemberRole `gorm:"column:member_role"`
	CreatedAt  time.Time       `gorm:"column:created_at"`
}

func (GroupMember) TableName() string {
	return "group_members"
}

type QuotaLog struct {
	ID           string       `gorm:"type:uuid;primaryKey"`
	UserID       string       `gorm:"column:user_id"`
	RequestID    *string      `gorm:"column:request_id"`
	LogType      QuotaLogType `gorm:"column:log_type"`
	DeltaQuota   int64        `gorm:"column:delta_quota"`
	BalanceAfter int64        `gorm:"column:balance_after"`
	Reason       *string      `gorm:"column:reason"`
	CreatedAt    time.Time    `gorm:"column:created_at"`
}

func (QuotaLog) TableName() string {
	return "quota_logs"
}
