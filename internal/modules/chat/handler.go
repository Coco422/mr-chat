package chat

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"mrchat/internal/http/middleware"
	"mrchat/internal/modules/limits"
	"mrchat/internal/shared/httpx"
)

type Handler struct {
	service *Service
}

type createConversationRequest struct {
	Title   string  `json:"title"`
	ModelID *string `json:"model_id"`
}

type completionMessageRequest struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type createCompletionRequest struct {
	ConversationID *string                    `json:"conversation_id"`
	ModelID        *string                    `json:"model_id"`
	Stream         bool                       `json:"stream"`
	Messages       []completionMessageRequest `json:"messages"`
	MaxTokens      *int                       `json:"max_tokens"`
	Metadata       map[string]any             `json:"metadata"`
}

type updateConversationRequest struct {
	Title string `json:"title" binding:"required"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ListConversations godoc
// @Summary List conversations
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page"
// @Param page_size query int false "Page size"
// @Param status query string false "Conversation status"
// @Success 200 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Router /conversations [get]
func (h *Handler) ListConversations(c *gin.Context) {
	page := parsePositiveInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parsePositiveInt(c.DefaultQuery("page_size", "20"), 20)
	items, err := h.service.ListConversations(
		c.Request.Context(),
		middleware.CurrentUserID(c),
		page,
		pageSize,
		c.DefaultQuery("status", ""),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.SuccessWithMeta(c, http.StatusOK, toConversationSummaries(items.Items), gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     items.Total,
	})
}

// CreateConversation godoc
// @Summary Create conversation
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body createConversationRequest true "Conversation payload"
// @Success 201 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Router /conversations [post]
func (h *Handler) CreateConversation(c *gin.Context) {
	var req createConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid conversation payload", gin.H{"error": err.Error()})
		return
	}

	modelID, ok := normalizeOptionalUUID(req.ModelID)
	if !ok {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "model_id must be a valid UUID", nil)
		return
	}

	item, err := h.service.CreateConversation(c.Request.Context(), middleware.CurrentUserID(c), req.Title, modelID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusCreated, toConversationSummary(item))
}

// UpdateConversation godoc
// @Summary Update conversation title
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation ID"
// @Param request body updateConversationRequest true "Conversation payload"
// @Success 200 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 404 {object} httpx.Envelope
// @Router /conversations/{id} [put]
func (h *Handler) UpdateConversation(c *gin.Context) {
	conversationID, ok := requireUUIDParam(c, "id")
	if !ok {
		return
	}

	var req updateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid conversation payload", gin.H{"error": err.Error()})
		return
	}

	item, err := h.service.UpdateConversationTitle(c.Request.Context(), middleware.CurrentUserID(c), conversationID, req.Title)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, toConversationSummary(item))
}

// DeleteConversation godoc
// @Summary Delete conversation
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation ID"
// @Success 200 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 404 {object} httpx.Envelope
// @Router /conversations/{id} [delete]
func (h *Handler) DeleteConversation(c *gin.Context) {
	conversationID, ok := requireUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteConversation(c.Request.Context(), middleware.CurrentUserID(c), conversationID); err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, gin.H{"deleted": true})
}

// ListMessages godoc
// @Summary List conversation messages
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation ID"
// @Param page query int false "Page"
// @Param page_size query int false "Page size"
// @Success 200 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 404 {object} httpx.Envelope
// @Router /conversations/{id}/messages [get]
func (h *Handler) ListMessages(c *gin.Context) {
	conversationID, ok := requireUUIDParam(c, "id")
	if !ok {
		return
	}

	page := parsePositiveInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parsePositiveInt(c.DefaultQuery("page_size", "50"), 50)
	items, err := h.service.ListMessages(c.Request.Context(), middleware.CurrentUserID(c), conversationID, page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.SuccessWithMeta(c, http.StatusOK, toMessages(items.Items), gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     items.Total,
	})
}

// CreateCompletion godoc
// @Summary Create chat completion
// @Description Main chat entry. `stream=false` returns JSON. `stream=true` returns `text/event-stream` SSE events with `response.start`, `response.delta`, `reasoning.delta`, `response.completed`, `response.failed`, and `[DONE]`.
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body createCompletionRequest true "Chat completion payload"
// @Success 200 {object} httpx.Envelope
// @Failure 400 {object} httpx.Envelope
// @Failure 401 {object} httpx.Envelope
// @Failure 404 {object} httpx.Envelope
// @Failure 429 {object} httpx.Envelope
// @Router /chat/completions [post]
func (h *Handler) CreateCompletion(c *gin.Context) {
	var req createCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid chat completion payload", gin.H{"error": err.Error()})
		return
	}

	conversationID, ok := normalizeOptionalUUID(req.ConversationID)
	if !ok {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "conversation_id must be a valid UUID", nil)
		return
	}
	modelID, ok := normalizeOptionalUUID(req.ModelID)
	if !ok {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "model_id must be a valid UUID", nil)
		return
	}

	messages := make([]CompletionMessageInput, 0, len(req.Messages))
	for _, item := range req.Messages {
		messages = append(messages, CompletionMessageInput{
			Role:    item.Role,
			Content: item.Content,
		})
	}

	if req.Stream {
		started := false
		err := h.service.StreamCompletion(c.Request.Context(), CompletionInput{
			RequestID:      httpx.RequestIDFromContext(c),
			UserID:         middleware.CurrentUserID(c),
			ConversationID: conversationID,
			ModelID:        modelID,
			Stream:         req.Stream,
			Messages:       messages,
			MaxTokens:      req.MaxTokens,
			Metadata:       req.Metadata,
		}, func(payload any) error {
			if !started {
				c.Header("Content-Type", "text/event-stream")
				c.Header("Cache-Control", "no-cache")
				c.Header("Connection", "keep-alive")
				c.Header("X-Accel-Buffering", "no")
				c.Status(http.StatusOK)
				c.Writer.WriteHeaderNow()
				started = true
			}
			return writeSSE(c, payload)
		})
		if err != nil && !started {
			h.handleError(c, err)
		}
		return
	}

	result, err := h.service.CreateCompletion(c.Request.Context(), CompletionInput{
		RequestID:      httpx.RequestIDFromContext(c),
		UserID:         middleware.CurrentUserID(c),
		ConversationID: conversationID,
		ModelID:        modelID,
		Stream:         req.Stream,
		Messages:       messages,
		MaxTokens:      req.MaxTokens,
		Metadata:       req.Metadata,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	httpx.Success(c, http.StatusOK, result)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrConversationNotFound):
		httpx.Failure(c, http.StatusNotFound, "CONVERSATION_NOT_FOUND", "Conversation not found", nil)
	case errors.Is(err, ErrModelNotAvailable):
		httpx.Failure(c, http.StatusBadRequest, "CHAT_MODEL_NOT_AVAILABLE", "Model is not available to the current user", nil)
	case errors.Is(err, ErrInvalidCompletionRequest):
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid chat completion payload", nil)
	case errors.Is(err, ErrStreamNotSupported):
		httpx.Failure(c, http.StatusBadRequest, "CHAT_STREAM_NOT_SUPPORTED", "stream=true is not supported in the first implementation", nil)
	case errors.Is(err, ErrUpstreamUnavailable):
		httpx.Failure(c, http.StatusBadGateway, "CHAT_UPSTREAM_UNAVAILABLE", "No upstream is currently available for this model", nil)
	case errors.Is(err, limits.ErrLimitExceeded):
		httpx.Failure(c, http.StatusTooManyRequests, "CHAT_LIMIT_EXCEEDED", "User model limit exceeded", nil)
	default:
		httpx.Failure(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
	}
}

func parsePositiveInt(value string, fallback int) int {
	var parsed int
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil || parsed <= 0 {
		return fallback
	}
	if parsed > 100 {
		return 100
	}
	return parsed
}

func toConversationSummaries(items []Conversation) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, toConversationSummary(&item))
	}
	return result
}

func toConversationSummary(item *Conversation) gin.H {
	if item == nil {
		return gin.H{}
	}
	return gin.H{
		"id":              item.ID,
		"title":           item.Title,
		"model_id":        item.ModelID,
		"last_message_at": formatTime(item.LastMessageAt),
		"message_count":   item.MessageCount,
		"status":          item.Status,
		"created_at":      item.CreatedAt.UTC().Format(timeLayout),
		"updated_at":      item.UpdatedAt.UTC().Format(timeLayout),
	}
}

func toMessages(items []Message) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, gin.H{
			"id":                item.ID,
			"conversation_id":   item.ConversationID,
			"role":              item.Role,
			"content":           item.Content,
			"reasoning_content": stringOrEmpty(item.ReasoningContent),
			"status":            item.Status,
			"finish_reason":     stringOrNil(item.FinishReason),
			"usage":             item.Usage,
			"created_at":        item.CreatedAt.UTC().Format(timeLayout),
		})
	}
	return result
}

func formatTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(timeLayout)
}

func stringOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func stringOrNil(value *string) *string {
	return value
}

func normalizeOptionalUUID(value *string) (*string, bool) {
	if value == nil {
		return nil, true
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, true
	}

	if _, err := uuid.Parse(trimmed); err != nil {
		return nil, false
	}

	return &trimmed, true
}

func requireUUIDParam(c *gin.Context, name string) (string, bool) {
	value := strings.TrimSpace(c.Param(name))
	if _, err := uuid.Parse(value); err != nil {
		httpx.Failure(c, http.StatusBadRequest, "VALIDATION_ERROR", fmt.Sprintf("%s must be a valid UUID", name), nil)
		return "", false
	}
	return value, true
}

func writeSSE(c *gin.Context, payload any) error {
	var body []byte
	switch value := payload.(type) {
	case string:
		body = []byte(value)
	default:
		encoded, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = encoded
	}
	if _, err := c.Writer.Write([]byte("data: ")); err != nil {
		return err
	}
	if _, err := c.Writer.Write(body); err != nil {
		return err
	}
	if _, err := c.Writer.Write([]byte("\n\n")); err != nil {
		return err
	}
	c.Writer.Flush()
	return nil
}

const timeLayout = "2006-01-02T15:04:05Z07:00"
