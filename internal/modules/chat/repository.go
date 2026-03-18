package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrConversationNotFound = errors.New("conversation not found")

type Repository struct {
	db *gorm.DB
}

type ListConversationsFilter struct {
	UserID   string
	Page     int
	PageSize int
	Status   string
}

type ConversationList struct {
	Items []Conversation
	Total int64
}

type MessageList struct {
	Items []Message
	Total int64
}

type CreateConversationInput struct {
	UserID  string
	Title   string
	ModelID *string
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) DB() *gorm.DB {
	return r.db
}

func (r *Repository) ListConversations(ctx context.Context, filter ListConversationsFilter) (ConversationList, error) {
	query := r.db.WithContext(ctx).Model(&Conversation{}).
		Where("user_id = ?", filter.UserID).
		Where("deleted_at IS NULL")

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	} else {
		query = query.Where("status = ?", ConversationStatusActive)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return ConversationList{}, fmt.Errorf("count conversations: %w", err)
	}

	var items []Conversation
	if err := query.Order("COALESCE(last_message_at, created_at) DESC").
		Limit(filter.PageSize).
		Offset((filter.Page - 1) * filter.PageSize).
		Find(&items).Error; err != nil {
		return ConversationList{}, fmt.Errorf("list conversations: %w", err)
	}

	return ConversationList{
		Items: items,
		Total: total,
	}, nil
}

func (r *Repository) CreateConversation(ctx context.Context, input CreateConversationInput) (*Conversation, error) {
	now := time.Now().UTC()
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "New conversation"
	}
	item := &Conversation{
		ID:           uuid.NewString(),
		UserID:       input.UserID,
		Title:        title,
		ModelID:      sanitizeOptionalString(input.ModelID),
		Status:       ConversationStatusActive,
		MessageCount: 0,
		Metadata:     map[string]any{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return nil, fmt.Errorf("create conversation: %w", err)
	}

	return item, nil
}

func (r *Repository) UpdateConversationTitle(ctx context.Context, userID, conversationID, title string) (*Conversation, error) {
	item, err := r.getConversation(ctx, userID, conversationID)
	if err != nil {
		return nil, err
	}

	item.Title = strings.TrimSpace(title)
	item.UpdatedAt = time.Now().UTC()
	if err := r.db.WithContext(ctx).Save(item).Error; err != nil {
		return nil, fmt.Errorf("update conversation title: %w", err)
	}

	return item, nil
}

func (r *Repository) DeleteConversation(ctx context.Context, userID, conversationID string) error {
	item, err := r.getConversation(ctx, userID, conversationID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	item.Status = ConversationStatusDeleted
	item.DeletedAt = &now
	item.UpdatedAt = now
	if err := r.db.WithContext(ctx).Save(item).Error; err != nil {
		return fmt.Errorf("delete conversation: %w", err)
	}

	return nil
}

func (r *Repository) ListMessages(ctx context.Context, userID, conversationID string, page, pageSize int) (MessageList, error) {
	if _, err := r.getConversation(ctx, userID, conversationID); err != nil {
		return MessageList{}, err
	}

	query := r.db.WithContext(ctx).Model(&Message{}).
		Where("conversation_id = ?", conversationID).
		Where("deleted_at IS NULL")

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return MessageList{}, fmt.Errorf("count messages: %w", err)
	}

	var items []Message
	if err := query.Order("created_at ASC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&items).Error; err != nil {
		return MessageList{}, fmt.Errorf("list messages: %w", err)
	}

	return MessageList{
		Items: items,
		Total: total,
	}, nil
}

func (r *Repository) ListAllMessages(ctx context.Context, userID, conversationID string) ([]Message, error) {
	if _, err := r.getConversation(ctx, userID, conversationID); err != nil {
		return nil, err
	}

	var items []Message
	if err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Where("deleted_at IS NULL").
		Order("created_at ASC").
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list all messages: %w", err)
	}

	return items, nil
}

func (r *Repository) getConversation(ctx context.Context, userID, conversationID string) (*Conversation, error) {
	var item Conversation
	if err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", conversationID, userID).
		First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrConversationNotFound
		}
		return nil, fmt.Errorf("get conversation: %w", err)
	}

	return &item, nil
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
