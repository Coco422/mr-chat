package chat

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"mrchat/internal/http/middleware"
	"mrchat/internal/shared/httpx"
)

type Handler struct {
	service *Service
}

type createConversationRequest struct {
	Title   string  `json:"title"`
	ModelID *string `json:"model_id"`
}

type updateConversationRequest struct {
	Title string `json:"title" binding:"required"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

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

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrConversationNotFound):
		httpx.Failure(c, http.StatusNotFound, "CONVERSATION_NOT_FOUND", "Conversation not found", nil)
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

const timeLayout = "2006-01-02T15:04:05Z07:00"
