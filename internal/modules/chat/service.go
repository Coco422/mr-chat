package chat

import (
	"context"
	"strings"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListConversations(ctx context.Context, userID string, page, pageSize int, status string) (ConversationList, error) {
	return s.repo.ListConversations(ctx, ListConversationsFilter{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
		Status:   strings.TrimSpace(status),
	})
}

func (s *Service) CreateConversation(ctx context.Context, userID, title string, modelID *string) (*Conversation, error) {
	return s.repo.CreateConversation(ctx, CreateConversationInput{
		UserID:  userID,
		Title:   title,
		ModelID: modelID,
	})
}

func (s *Service) UpdateConversationTitle(ctx context.Context, userID, conversationID, title string) (*Conversation, error) {
	return s.repo.UpdateConversationTitle(ctx, userID, conversationID, title)
}

func (s *Service) DeleteConversation(ctx context.Context, userID, conversationID string) error {
	return s.repo.DeleteConversation(ctx, userID, conversationID)
}

func (s *Service) ListMessages(ctx context.Context, userID, conversationID string, page, pageSize int) (MessageList, error) {
	return s.repo.ListMessages(ctx, userID, conversationID, page, pageSize)
}
