package chat

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mrchat/internal/modules/account"
	"mrchat/internal/modules/catalog"
	"mrchat/internal/modules/limits"
)

var (
	ErrModelNotAvailable        = errors.New("model not available")
	ErrUpstreamUnavailable      = errors.New("upstream unavailable")
	ErrInvalidCompletionRequest = errors.New("invalid chat completion request")
	ErrStreamNotSupported       = errors.New("stream is not supported yet")
)

type Service struct {
	repo          *Repository
	accountRepo   *account.Repository
	catalogRepo   *catalog.Repository
	limitsService *limits.Service
	client        *openAICompatibleClient
}

type CompletionMessageInput struct {
	Role    string
	Content string
}

type CompletionInput struct {
	RequestID      string
	UserID         string
	ConversationID *string
	ModelID        *string
	Stream         bool
	Messages       []CompletionMessageInput
	MaxTokens      *int
	Metadata       map[string]any
}

type CompletionUsage struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
}

type CompletionBilling struct {
	PreDeducted  int64 `json:"pre_deducted"`
	FinalCharged int64 `json:"final_charged"`
	Refunded     int64 `json:"refunded"`
}

type CompletionResult struct {
	ConversationID string            `json:"conversation_id"`
	Message        CompletionMessage `json:"message"`
	Usage          CompletionUsage   `json:"usage"`
	Billing        CompletionBilling `json:"billing"`
}

type CompletionMessage struct {
	ID               string         `json:"id"`
	Role             string         `json:"role"`
	Content          string         `json:"content"`
	ReasoningContent string         `json:"reasoning_content"`
	Status           MessageStatus  `json:"status"`
	FinishReason     *string        `json:"finish_reason"`
	Usage            map[string]any `json:"usage"`
	CreatedAt        string         `json:"created_at"`
}

func NewService(repo *Repository, accountRepo *account.Repository, catalogRepo *catalog.Repository, limitsService *limits.Service) *Service {
	return &Service{
		repo:          repo,
		accountRepo:   accountRepo,
		catalogRepo:   catalogRepo,
		limitsService: limitsService,
		client:        &openAICompatibleClient{},
	}
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

func (s *Service) CreateCompletion(ctx context.Context, input CompletionInput) (*CompletionResult, error) {
	if strings.TrimSpace(input.RequestID) == "" {
		input.RequestID = uuid.NewString()
	}

	if input.Stream {
		return nil, ErrStreamNotSupported
	}

	prepared, err := s.prepareCompletion(ctx, input)
	if err != nil {
		return nil, err
	}

	if err := s.reserveQuota(ctx, input.UserID, input.RequestID, prepared.reservedQuota); err != nil {
		return nil, err
	}

	upstreamResponse, routeBinding, upstream, attempts, err := s.completeWithRoutes(
		ctx,
		prepared.modelWithBindings,
		prepared.combinedMessages,
		prepared.reservedCompletionTokens,
		input.Metadata,
	)
	if err != nil {
		refundErr := s.accountRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if _, refundErr := s.settleReservedQuotaWithDB(ctx, tx, input.UserID, input.RequestID, prepared.reservedQuota, 0); refundErr != nil {
				return refundErr
			}
			_, logErr := s.limitsService.CreateRequestLogWithDB(ctx, tx, limits.RequestLogCreateInput{
				RequestID:      input.RequestID,
				UserID:         input.UserID,
				UserGroupID:    prepared.user.UserGroupID,
				ConversationID: sanitizeOptionalString(input.ConversationID),
				ModelID:        prepared.modelID,
				PromptTokens:   estimatePromptTokens(prepared.combinedMessages),
				BilledQuota:    0,
				Status:         limits.RequestLogStatusFailed,
				ErrorCode:      stringPtr("CHAT_UPSTREAM_UNAVAILABLE"),
				StartedAt:      prepared.startedAt,
				CompletedAt:    timePtr(time.Now().UTC()),
				Metadata: map[string]any{
					"attempts": attempts,
				},
			})
			return logErr
		})
		if refundErr != nil {
			return nil, refundErr
		}
		return nil, err
	}

	usage := normalizeCompletionUsage(
		upstreamResponse.Usage,
		prepared.combinedMessages,
		upstreamResponse.assistantContent(),
		upstreamResponse.reasoningContent(),
	)

	var result *CompletionResult
	err = s.repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		activeConversation := prepared.conversation
		now := time.Now().UTC()
		if activeConversation == nil {
			title := buildConversationTitle(prepared.firstUserMessage)
			item := &Conversation{
				ID:           uuid.NewString(),
				UserID:       input.UserID,
				Title:        title,
				ModelID:      prepared.modelID,
				Status:       ConversationStatusActive,
				MessageCount: 0,
				Metadata:     map[string]any{},
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := tx.Create(item).Error; err != nil {
				return fmt.Errorf("create conversation in completion: %w", err)
			}
			activeConversation = item
		}

		if activeConversation.ModelID == nil || (prepared.modelID != nil && *activeConversation.ModelID != *prepared.modelID) {
			activeConversation.ModelID = prepared.modelID
		}

		createdMessages, err := createRequestMessages(ctx, tx, activeConversation.ID, input.UserID, prepared.modelID, input.RequestID, prepared.normalizedMessages)
		if err != nil {
			return err
		}

		assistantMessage, err := createAssistantMessage(ctx, tx, assistantMessageInput{
			ConversationID:   activeConversation.ID,
			UserID:           input.UserID,
			ModelID:          prepared.modelID,
			UpstreamID:       &upstream.ID,
			RequestID:        input.RequestID,
			Content:          upstreamResponse.assistantContent(),
			ReasoningContent: emptyToNil(upstreamResponse.reasoningContent()),
			FinishReason:     emptyToNil(upstreamResponse.finishReason()),
			Usage: map[string]any{
				"prompt_tokens":     usage.PromptTokens,
				"completion_tokens": usage.CompletionTokens,
				"total_tokens":      usage.TotalTokens,
			},
			Status: MessageStatusCompleted,
			Metadata: map[string]any{
				"provider_model": upstreamResponse.Model,
			},
		})
		if err != nil {
			return err
		}

		activeConversation.MessageCount += len(createdMessages) + 1
		activeConversation.LastMessageAt = timePtr(now)
		activeConversation.UpdatedAt = now
		if err := tx.Save(activeConversation).Error; err != nil {
			return fmt.Errorf("update conversation after completion: %w", err)
		}

		billing, err := s.settleReservedQuotaWithDB(ctx, tx, input.UserID, input.RequestID, prepared.reservedQuota, usage.TotalTokens)
		if err != nil {
			return err
		}

		if _, err := s.limitsService.CreateRequestLogWithDB(ctx, tx, limits.RequestLogCreateInput{
			RequestID:        input.RequestID,
			UserID:           input.UserID,
			UserGroupID:      prepared.user.UserGroupID,
			ConversationID:   &activeConversation.ID,
			MessageID:        &assistantMessage.ID,
			ModelID:          prepared.modelID,
			ChannelID:        routeBinding.ChannelID,
			PromptTokens:     usage.PromptTokens,
			CompletionTokens: usage.CompletionTokens,
			TotalTokens:      usage.TotalTokens,
			BilledQuota:      billing.FinalCharged,
			Status:           limits.RequestLogStatusCompleted,
			StartedAt:        prepared.startedAt,
			CompletedAt:      timePtr(now),
			Metadata: map[string]any{
				"attempts":       attempts,
				"upstream_id":    upstream.ID,
				"provider_model": upstreamResponse.Model,
			},
		}); err != nil {
			return err
		}

		result = &CompletionResult{
			ConversationID: activeConversation.ID,
			Message: CompletionMessage{
				ID:               assistantMessage.ID,
				Role:             assistantMessage.Role,
				Content:          assistantMessage.Content,
				ReasoningContent: stringOrEmpty(assistantMessage.ReasoningContent),
				Status:           assistantMessage.Status,
				FinishReason:     assistantMessage.FinishReason,
				Usage:            assistantMessage.Usage,
				CreatedAt:        assistantMessage.CreatedAt.UTC().Format(timeLayout),
			},
			Usage: CompletionUsage{
				PromptTokens:     usage.PromptTokens,
				CompletionTokens: usage.CompletionTokens,
				TotalTokens:      usage.TotalTokens,
			},
			Billing: billing,
		}
		return nil
	})
	if err != nil {
		if refundErr := s.refundReservedQuota(ctx, input.UserID, input.RequestID, prepared.reservedQuota); refundErr != nil {
			return nil, refundErr
		}
		return nil, err
	}

	return result, nil
}

func (s *Service) loadConversationHistory(ctx context.Context, userID string, conversation *Conversation) ([]openAIChatMessage, error) {
	if conversation == nil {
		return []openAIChatMessage{}, nil
	}

	items, err := s.repo.ListAllMessages(ctx, userID, conversation.ID)
	if err != nil {
		return nil, err
	}

	result := make([]openAIChatMessage, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.Content) == "" {
			continue
		}
		if item.Status == MessageStatusFailed {
			continue
		}
		result = append(result, openAIChatMessage{
			Role:    item.Role,
			Content: item.Content,
		})
	}
	return result, nil
}

func (s *Service) completeWithRoutes(ctx context.Context, model *catalog.ModelWithBindings, messages []openAIChatMessage, maxTokens int, metadata map[string]any) (*openAIChatCompletionResponse, catalog.ModelRouteBinding, *catalog.Upstream, []map[string]any, error) {
	attempts := make([]map[string]any, 0, len(model.RouteBindings))
	for _, binding := range model.RouteBindings {
		if binding.Status != catalog.RouteBindingStatusActive {
			continue
		}

		upstream, err := s.loadUpstream(ctx, binding.UpstreamID)
		if err != nil {
			attempts = append(attempts, map[string]any{
				"upstream_id": binding.UpstreamID,
				"error":       err.Error(),
			})
			continue
		}
		if upstream.Status != catalog.UpstreamStatusActive {
			attempts = append(attempts, map[string]any{
				"upstream_id": binding.UpstreamID,
				"error":       "upstream_not_active",
			})
			continue
		}

		response, err := s.client.ChatCompletion(ctx, upstream, openAIChatCompletionRequest{
			Model:    model.Model.ModelKey,
			Messages: messages,
			Stream:   false,
			MaxTokens: func() *int {
				if maxTokens <= 0 {
					return nil
				}
				return &maxTokens
			}(),
			Metadata: metadata,
		})
		if err == nil {
			attempts = append(attempts, map[string]any{
				"upstream_id": binding.UpstreamID,
				"channel_id":  binding.ChannelID,
				"result":      "success",
			})
			return response, binding, upstream, attempts, nil
		}

		attempts = append(attempts, map[string]any{
			"upstream_id": binding.UpstreamID,
			"channel_id":  binding.ChannelID,
			"error":       err.Error(),
		})
	}

	return nil, catalog.ModelRouteBinding{}, nil, attempts, ErrUpstreamUnavailable
}

func (s *Service) loadUpstream(ctx context.Context, upstreamID string) (*catalog.Upstream, error) {
	var upstream catalog.Upstream
	if err := s.catalogRepo.DB().WithContext(ctx).First(&upstream, "id = ?", upstreamID).Error; err != nil {
		return nil, err
	}
	return &upstream, nil
}

func canUserAccessModel(user *account.User, visibleUserGroupIDs []string) bool {
	if user.Role == account.RoleAdmin || user.Role == account.RoleRoot {
		return true
	}
	if len(visibleUserGroupIDs) == 0 {
		return true
	}
	if user.UserGroupID == nil || *user.UserGroupID == "" {
		return false
	}
	return slices.Contains(visibleUserGroupIDs, *user.UserGroupID)
}

func normalizeCompletionMessages(items []CompletionMessageInput) ([]openAIChatMessage, string, error) {
	result := make([]openAIChatMessage, 0, len(items))
	firstUser := ""
	for _, item := range items {
		role := strings.TrimSpace(item.Role)
		content := strings.TrimSpace(item.Content)
		if role == "" || content == "" {
			continue
		}
		if role != "system" && role != "user" && role != "assistant" {
			return nil, "", ErrInvalidCompletionRequest
		}
		if role == "user" && firstUser == "" {
			firstUser = content
		}
		result = append(result, openAIChatMessage{
			Role:    role,
			Content: content,
		})
	}
	if len(result) == 0 || firstUser == "" {
		return nil, "", ErrInvalidCompletionRequest
	}
	return result, firstUser, nil
}

func resolveReservedCompletionTokens(requestMaxTokens *int, modelMaxOutputTokens *int) int {
	if requestMaxTokens != nil && *requestMaxTokens > 0 {
		return *requestMaxTokens
	}
	if modelMaxOutputTokens != nil && *modelMaxOutputTokens > 0 {
		return *modelMaxOutputTokens
	}
	return 1024
}

func estimatePromptTokens(messages []openAIChatMessage) int64 {
	var total int64
	for _, item := range messages {
		total += 4 + estimateTextTokens(item.Role) + estimateTextTokens(item.Content)
	}
	if total == 0 {
		return 1
	}
	return total
}

func estimateTextTokens(value string) int64 {
	runes := len([]rune(strings.TrimSpace(value)))
	if runes == 0 {
		return 0
	}
	return int64((runes + 3) / 4)
}

func normalizeCompletionUsage(usage openAIUsage, promptMessages []openAIChatMessage, assistantContent, reasoningContent string) CompletionUsage {
	promptTokens := int64(usage.PromptTokens)
	completionTokens := int64(usage.CompletionTokens)
	totalTokens := int64(usage.TotalTokens)

	if promptTokens == 0 {
		promptTokens = estimatePromptTokens(promptMessages)
	}
	if completionTokens == 0 {
		completionTokens = estimateTextTokens(assistantContent) + estimateTextTokens(reasoningContent)
	}
	if totalTokens == 0 {
		totalTokens = promptTokens + completionTokens
	}

	return CompletionUsage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
	}
}

func buildConversationTitle(message string) string {
	normalized := strings.Join(strings.Fields(message), " ")
	if normalized == "" {
		return "New conversation"
	}

	runes := []rune(normalized)
	if len(runes) <= 32 {
		return normalized
	}
	return string(runes[:32])
}

type assistantMessageInput struct {
	ConversationID   string
	UserID           string
	ModelID          *string
	UpstreamID       *string
	RequestID        string
	Content          string
	ReasoningContent *string
	FinishReason     *string
	Usage            map[string]any
	Status           MessageStatus
	Metadata         map[string]any
}

func createRequestMessages(ctx context.Context, tx *gorm.DB, conversationID, userID string, modelID *string, requestID string, messages []openAIChatMessage) ([]Message, error) {
	now := time.Now().UTC()
	rows := make([]Message, 0, len(messages))
	for _, item := range messages {
		row := Message{
			ID:             uuid.NewString(),
			ConversationID: conversationID,
			UserID:         userID,
			ModelID:        modelID,
			RequestID:      stringPtr(requestID),
			Role:           item.Role,
			Content:        item.Content,
			Status:         MessageStatusCompleted,
			Usage:          map[string]any{},
			Metadata:       map[string]any{},
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return []Message{}, nil
	}
	if err := tx.WithContext(ctx).Create(&rows).Error; err != nil {
		return nil, fmt.Errorf("create request messages: %w", err)
	}
	return rows, nil
}

func createAssistantMessage(ctx context.Context, tx *gorm.DB, input assistantMessageInput) (*Message, error) {
	now := time.Now().UTC()
	item := &Message{
		ID:               uuid.NewString(),
		ConversationID:   input.ConversationID,
		UserID:           input.UserID,
		ModelID:          input.ModelID,
		UpstreamID:       input.UpstreamID,
		RequestID:        stringPtr(input.RequestID),
		Role:             "assistant",
		Content:          input.Content,
		ReasoningContent: input.ReasoningContent,
		Status:           input.Status,
		FinishReason:     input.FinishReason,
		Usage:            nonNilMap(input.Usage),
		Metadata:         nonNilMap(input.Metadata),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := tx.WithContext(ctx).Create(item).Error; err != nil {
		return nil, fmt.Errorf("create assistant message: %w", err)
	}
	return item, nil
}

func nonNilMap(value map[string]any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	return value
}

func emptyToNil(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func stringPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func timePtr(value time.Time) *time.Time {
	return &value
}
