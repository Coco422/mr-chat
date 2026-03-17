package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

type CreateInput struct {
	ActorUserID  *string
	ActorRole    *string
	Action       string
	ResourceType string
	ResourceID   *string
	TargetUserID *string
	RequestID    *string
	IPAddress    *string
	UserAgent    *string
	Result       Result
	Details      map[string]any
}

type ListFilter struct {
	Page         int
	PageSize     int
	ActorUserID  string
	Action       string
	ResourceType string
	Result       string
}

type ListResult struct {
	Items []Log
	Total int64
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, input CreateInput) error {
	return r.CreateWithDB(ctx, r.db, input)
}

func (r *Repository) CreateWithDB(ctx context.Context, db *gorm.DB, input CreateInput) error {
	if db == nil {
		db = r.db
	}

	entry := Log{
		ID:           uuid.NewString(),
		ActorUserID:  input.ActorUserID,
		ActorRole:    input.ActorRole,
		Action:       input.Action,
		ResourceType: input.ResourceType,
		ResourceID:   input.ResourceID,
		TargetUserID: input.TargetUserID,
		RequestID:    input.RequestID,
		IPAddress:    input.IPAddress,
		UserAgent:    input.UserAgent,
		Result:       input.Result,
		Details:      nonNilMap(input.Details),
		CreatedAt:    time.Now().UTC(),
	}

	if entry.Result == "" {
		entry.Result = ResultSuccess
	}

	if err := db.WithContext(ctx).Create(&entry).Error; err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}

	return nil
}

func (r *Repository) List(ctx context.Context, filter ListFilter) (ListResult, error) {
	query := r.db.WithContext(ctx).Model(&Log{})
	if filter.ActorUserID != "" {
		query = query.Where("actor_user_id = ?", filter.ActorUserID)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.ResourceType != "" {
		query = query.Where("resource_type = ?", filter.ResourceType)
	}
	if filter.Result != "" {
		query = query.Where("result = ?", filter.Result)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return ListResult{}, fmt.Errorf("count audit logs: %w", err)
	}

	var items []Log
	if err := query.Order("created_at DESC").
		Limit(filter.PageSize).
		Offset((filter.Page - 1) * filter.PageSize).
		Find(&items).Error; err != nil {
		return ListResult{}, fmt.Errorf("list audit logs: %w", err)
	}

	return ListResult{
		Items: items,
		Total: total,
	}, nil
}

func nonNilMap(value map[string]any) map[string]any {
	if value == nil {
		return map[string]any{}
	}

	return value
}
