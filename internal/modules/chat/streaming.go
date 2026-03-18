package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mrchat/internal/modules/account"
	"mrchat/internal/modules/catalog"
	"mrchat/internal/modules/limits"
)

type StreamEmitFunc func(payload any) error

type completionPreparedState struct {
	startedAt                time.Time
	user                     *account.User
	conversation             *Conversation
	modelWithBindings        *catalog.ModelWithBindings
	modelID                  *string
	normalizedMessages       []openAIChatMessage
	firstUserMessage         string
	combinedMessages         []openAIChatMessage
	reservedCompletionTokens int
}

type streamTurnState struct {
	conversation  *Conversation
	assistant     *Message
	routeBinding  catalog.ModelRouteBinding
	upstream      *catalog.Upstream
	providerModel string
	attempts      []map[string]any
	createdAt     time.Time
}

func (s *Service) StreamCompletion(ctx context.Context, input CompletionInput, emit StreamEmitFunc) error {
	if strings.TrimSpace(input.RequestID) == "" {
		input.RequestID = uuid.NewString()
	}

	prepared, err := s.prepareCompletion(ctx, input)
	if err != nil {
		return err
	}

	stream, routeBinding, upstream, attempts, err := s.openStreamWithRoutes(ctx, prepared.modelWithBindings, prepared.combinedMessages, prepared.reservedCompletionTokens, input.Metadata)
	if err != nil {
		_, _ = s.limitsService.CreateRequestLog(ctx, limits.RequestLogCreateInput{
			RequestID:      input.RequestID,
			UserID:         input.UserID,
			UserGroupID:    prepared.user.UserGroupID,
			ConversationID: sanitizeOptionalString(input.ConversationID),
			ModelID:        prepared.modelID,
			PromptTokens:   estimatePromptTokens(prepared.combinedMessages),
			Status:         limits.RequestLogStatusFailed,
			ErrorCode:      stringPtr("CHAT_UPSTREAM_UNAVAILABLE"),
			StartedAt:      prepared.startedAt,
			CompletedAt:    timePtr(time.Now().UTC()),
			Metadata: map[string]any{
				"attempts": attempts,
			},
		})
		return err
	}
	defer stream.Close()

	turnState, err := s.createStreamingTurn(ctx, input, prepared, routeBinding, upstream, attempts)
	if err != nil {
		return err
	}

	if err := emit(map[string]any{
		"type":                 "response.start",
		"request_id":           input.RequestID,
		"conversation_id":      turnState.conversation.ID,
		"assistant_message_id": turnState.assistant.ID,
	}); err != nil {
		_ = s.failStreamingTurn(ctx, turnState, prepared, "CHAT_STREAM_CLIENT_DISCONNECTED", limits.RequestLogStatusCancelled, MessageStatusCancelled, "", "")
		return nil
	}

	var contentBuilder strings.Builder
	var reasoningBuilder strings.Builder
	finishReason := ""
	usage := openAIUsage{}

	for {
		chunk, done, err := stream.Next()
		if err != nil {
			status := limits.RequestLogStatusFailed
			messageStatus := MessageStatusFailed
			errorCode := "CHAT_STREAM_READ_FAILED"
			if errors.Is(ctx.Err(), context.Canceled) || errors.Is(err, context.Canceled) {
				status = limits.RequestLogStatusCancelled
				messageStatus = MessageStatusCancelled
				errorCode = "CHAT_STREAM_CLIENT_DISCONNECTED"
			}
			_ = s.failStreamingTurn(ctx, turnState, prepared, errorCode, status, messageStatus, contentBuilder.String(), reasoningBuilder.String())
			if status == limits.RequestLogStatusCancelled {
				return nil
			}
			_ = emit(map[string]any{
				"type":                 "response.failed",
				"request_id":           stringOrEmpty(turnState.assistant.RequestID),
				"conversation_id":      turnState.conversation.ID,
				"assistant_message_id": turnState.assistant.ID,
				"error": map[string]any{
					"code":    errorCode,
					"message": "streaming failed",
				},
			})
			_ = emit("[DONE]")
			return nil
		}
		if done {
			break
		}

		if chunkUsage := chunk.usage(); chunkUsage != (openAIUsage{}) {
			usage = chunkUsage
		}

		if delta := chunk.reasoningDelta(); delta != "" {
			reasoningBuilder.WriteString(delta)
			if err := emit(map[string]any{
				"type": "reasoning.delta",
				"delta": map[string]any{
					"reasoning_content": delta,
				},
			}); err != nil {
				_ = s.failStreamingTurn(ctx, turnState, prepared, "CHAT_STREAM_CLIENT_DISCONNECTED", limits.RequestLogStatusCancelled, MessageStatusCancelled, contentBuilder.String(), reasoningBuilder.String())
				return nil
			}
		}

		if delta := chunk.contentDelta(); delta != "" {
			contentBuilder.WriteString(delta)
			if err := emit(map[string]any{
				"type": "response.delta",
				"delta": map[string]any{
					"content": delta,
				},
			}); err != nil {
				_ = s.failStreamingTurn(ctx, turnState, prepared, "CHAT_STREAM_CLIENT_DISCONNECTED", limits.RequestLogStatusCancelled, MessageStatusCancelled, contentBuilder.String(), reasoningBuilder.String())
				return nil
			}
		}

		if currentFinishReason := chunk.finishReason(); currentFinishReason != "" {
			finishReason = currentFinishReason
		}
	}

	finalUsage := normalizeCompletionUsage(usage, prepared.combinedMessages, contentBuilder.String(), reasoningBuilder.String())
	if finishReason == "" {
		finishReason = "stop"
	}

	if err := s.completeStreamingTurn(ctx, turnState, finalUsage, contentBuilder.String(), reasoningBuilder.String(), finishReason); err != nil {
		return err
	}

	_ = emit(map[string]any{
		"type":                 "response.completed",
		"request_id":           stringOrEmpty(turnState.assistant.RequestID),
		"conversation_id":      turnState.conversation.ID,
		"assistant_message_id": turnState.assistant.ID,
		"usage": map[string]any{
			"prompt_tokens":     finalUsage.PromptTokens,
			"completion_tokens": finalUsage.CompletionTokens,
			"total_tokens":      finalUsage.TotalTokens,
		},
		"billing": map[string]any{
			"pre_deducted":  0,
			"final_charged": 0,
			"refunded":      0,
		},
		"finish_reason": finishReason,
	})
	_ = emit("[DONE]")

	return nil
}

func (s *Service) prepareCompletion(ctx context.Context, input CompletionInput) (*completionPreparedState, error) {
	startedAt := time.Now().UTC()
	if strings.TrimSpace(input.RequestID) == "" {
		input.RequestID = uuid.NewString()
	}

	messages, firstUserMessage, err := normalizeCompletionMessages(input.Messages)
	if err != nil {
		return nil, err
	}

	user, err := s.accountRepo.GetUserByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	conversationID := sanitizeOptionalString(input.ConversationID)
	var conversation *Conversation
	if conversationID != nil {
		conversation, err = s.repo.getConversation(ctx, input.UserID, *conversationID)
		if err != nil {
			return nil, err
		}
	}

	modelID := sanitizeOptionalString(input.ModelID)
	if modelID == nil && conversation != nil {
		modelID = sanitizeOptionalString(conversation.ModelID)
	}
	if modelID == nil {
		return nil, ErrInvalidCompletionRequest
	}

	modelWithBindings, err := s.catalogRepo.GetModelByID(ctx, *modelID)
	if err != nil {
		return nil, ErrModelNotAvailable
	}
	if modelWithBindings.Model.Status != catalog.ModelStatusActive || !canUserAccessModel(user, modelWithBindings.Model.VisibleUserGroupIDs) {
		return nil, ErrModelNotAvailable
	}

	history, err := s.loadConversationHistory(ctx, input.UserID, conversation)
	if err != nil {
		return nil, err
	}
	combinedMessages := append(history, messages...)

	reservedCompletionTokens := resolveReservedCompletionTokens(input.MaxTokens, modelWithBindings.Model.MaxOutputTokens)
	limitResult, err := s.limitsService.CheckUserModelLimit(ctx, limits.LimitCheckInput{
		UserID:                   input.UserID,
		ModelID:                  modelID,
		PromptTokens:             estimatePromptTokens(combinedMessages),
		ReservedCompletionTokens: int64(reservedCompletionTokens),
		Now:                      startedAt,
	})
	if err != nil {
		if errors.Is(err, limits.ErrLimitExceeded) {
			_ = s.limitsService.RecordRejectedRequest(ctx, input.RequestID, limitResult.Report, "CHAT_LIMIT_EXCEEDED", map[string]any{
				"conversation_id": conversationID,
				"model_id":        modelID,
				"reason":          "user_model_limit",
			})
		}
		return nil, err
	}

	return &completionPreparedState{
		startedAt:                startedAt,
		user:                     user,
		conversation:             conversation,
		modelWithBindings:        modelWithBindings,
		modelID:                  modelID,
		normalizedMessages:       messages,
		firstUserMessage:         firstUserMessage,
		combinedMessages:         combinedMessages,
		reservedCompletionTokens: reservedCompletionTokens,
	}, nil
}

func (s *Service) openStreamWithRoutes(ctx context.Context, model *catalog.ModelWithBindings, messages []openAIChatMessage, maxTokens int, metadata map[string]any) (*openAIChatCompletionStream, catalog.ModelRouteBinding, *catalog.Upstream, []map[string]any, error) {
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

		stream, err := s.client.OpenChatCompletionStream(ctx, upstream, openAIChatCompletionRequest{
			Model:    model.Model.ModelKey,
			Messages: messages,
			Stream:   true,
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
			return stream, binding, upstream, attempts, nil
		}

		attempts = append(attempts, map[string]any{
			"upstream_id": binding.UpstreamID,
			"channel_id":  binding.ChannelID,
			"error":       err.Error(),
		})
	}

	return nil, catalog.ModelRouteBinding{}, nil, attempts, ErrUpstreamUnavailable
}

func (s *Service) createStreamingTurn(ctx context.Context, input CompletionInput, prepared *completionPreparedState, routeBinding catalog.ModelRouteBinding, upstream *catalog.Upstream, attempts []map[string]any) (*streamTurnState, error) {
	state := &streamTurnState{
		conversation: prepared.conversation,
		routeBinding: routeBinding,
		upstream:     upstream,
		providerModel: func() string {
			if prepared == nil || prepared.modelWithBindings == nil {
				return ""
			}
			return prepared.modelWithBindings.Model.ModelKey
		}(),
		attempts:  attempts,
		createdAt: time.Now().UTC(),
	}

	err := s.repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		if state.conversation == nil {
			item := &Conversation{
				ID:           uuid.NewString(),
				UserID:       input.UserID,
				Title:        buildConversationTitle(prepared.firstUserMessage),
				ModelID:      prepared.modelID,
				Status:       ConversationStatusActive,
				MessageCount: 0,
				Metadata:     map[string]any{},
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := tx.Create(item).Error; err != nil {
				return fmt.Errorf("create conversation for stream: %w", err)
			}
			state.conversation = item
		}

		if state.conversation.ModelID == nil || (prepared.modelID != nil && *state.conversation.ModelID != *prepared.modelID) {
			state.conversation.ModelID = prepared.modelID
		}

		createdMessages, err := createRequestMessages(ctx, tx, state.conversation.ID, input.UserID, prepared.modelID, input.RequestID, prepared.normalizedMessages)
		if err != nil {
			return err
		}

		assistant, err := createAssistantMessage(ctx, tx, assistantMessageInput{
			ConversationID: state.conversation.ID,
			UserID:         input.UserID,
			ModelID:        prepared.modelID,
			UpstreamID:     &upstream.ID,
			RequestID:      input.RequestID,
			Status:         MessageStatusStreaming,
			Usage:          map[string]any{},
			Metadata: map[string]any{
				"provider_model": state.providerModel,
			},
		})
		if err != nil {
			return err
		}
		state.assistant = assistant

		state.conversation.MessageCount += len(createdMessages) + 1
		state.conversation.LastMessageAt = timePtr(now)
		state.conversation.UpdatedAt = now
		if err := tx.Save(state.conversation).Error; err != nil {
			return fmt.Errorf("update conversation for stream: %w", err)
		}

		if _, err := s.limitsService.CreateRequestLogWithDB(ctx, tx, limits.RequestLogCreateInput{
			RequestID:      input.RequestID,
			UserID:         input.UserID,
			UserGroupID:    prepared.user.UserGroupID,
			ConversationID: &state.conversation.ID,
			MessageID:      &state.assistant.ID,
			ModelID:        prepared.modelID,
			ChannelID:      routeBinding.ChannelID,
			Status:         limits.RequestLogStatusPending,
			StartedAt:      prepared.startedAt,
			Metadata: map[string]any{
				"attempts":       attempts,
				"upstream_id":    upstream.ID,
				"provider_model": state.providerModel,
			},
		}); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (s *Service) completeStreamingTurn(ctx context.Context, state *streamTurnState, usage CompletionUsage, content, reasoningContent, finishReason string) error {
	now := time.Now().UTC()
	status := limits.RequestLogStatusCompleted

	return s.repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Message{}).
			Where("id = ?", state.assistant.ID).
			Updates(map[string]any{
				"content":           content,
				"reasoning_content": emptyToNil(reasoningContent),
				"status":            MessageStatusCompleted,
				"finish_reason":     emptyToNil(finishReason),
				"usage_json": map[string]any{
					"prompt_tokens":     usage.PromptTokens,
					"completion_tokens": usage.CompletionTokens,
					"total_tokens":      usage.TotalTokens,
				},
				"updated_at": now,
				"metadata_json": map[string]any{
					"provider_model": state.providerModel,
				},
			}).Error; err != nil {
			return fmt.Errorf("complete assistant message: %w", err)
		}

		if _, err := s.limitsService.UpdateRequestLogByRequestIDWithDB(ctx, tx, stringOrEmpty(state.assistant.RequestID), limits.RequestLogUpdateInput{
			PromptTokens:     int64Ptr(usage.PromptTokens),
			CompletionTokens: int64Ptr(usage.CompletionTokens),
			TotalTokens:      int64Ptr(usage.TotalTokens),
			BilledQuota:      int64Ptr(0),
			Status:           &status,
			CompletedAt:      timePtr(now),
			Metadata: map[string]any{
				"attempts":       state.attempts,
				"upstream_id":    state.upstream.ID,
				"provider_model": state.providerModel,
			},
		}); err != nil {
			return err
		}

		return nil
	})
}

func (s *Service) failStreamingTurn(ctx context.Context, state *streamTurnState, prepared *completionPreparedState, errorCode string, requestStatus limits.RequestLogStatus, messageStatus MessageStatus, content, reasoningContent string) error {
	now := time.Now().UTC()
	usage := normalizeCompletionUsage(openAIUsage{}, prepared.combinedMessages, content, reasoningContent)

	return s.repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Message{}).
			Where("id = ?", state.assistant.ID).
			Updates(map[string]any{
				"content":           content,
				"reasoning_content": emptyToNil(reasoningContent),
				"status":            messageStatus,
				"error_code":        emptyToNil(errorCode),
				"usage_json": map[string]any{
					"prompt_tokens":     usage.PromptTokens,
					"completion_tokens": usage.CompletionTokens,
					"total_tokens":      usage.TotalTokens,
				},
				"updated_at": now,
				"metadata_json": map[string]any{
					"provider_model": state.providerModel,
				},
			}).Error; err != nil {
			return fmt.Errorf("update failed stream message: %w", err)
		}

		if _, err := s.limitsService.UpdateRequestLogByRequestIDWithDB(ctx, tx, stringOrEmpty(state.assistant.RequestID), limits.RequestLogUpdateInput{
			PromptTokens:     int64Ptr(usage.PromptTokens),
			CompletionTokens: int64Ptr(usage.CompletionTokens),
			TotalTokens:      int64Ptr(usage.TotalTokens),
			Status:           &requestStatus,
			ErrorCode:        &errorCode,
			CompletedAt:      timePtr(now),
			Metadata: map[string]any{
				"attempts":       state.attempts,
				"upstream_id":    state.upstream.ID,
				"provider_model": state.providerModel,
			},
		}); err != nil {
			return err
		}

		return nil
	})
}

func int64Ptr(value int64) *int64 {
	return &value
}
