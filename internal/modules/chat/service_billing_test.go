package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"mrchat/internal/modules/account"
	"mrchat/internal/modules/catalog"
	"mrchat/internal/modules/limits"
)

func TestCreateCompletionSettlesQuotaAndRequestLog(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openChatBillingTestDB(t)
	service, user, modelID := seedChatBillingFixture(t, ctx, db, newCompletionServer(t))

	inputMessages := []CompletionMessageInput{
		{Role: "user", Content: "Explain reserve and settlement"},
	}

	normalizedMessages, _, err := normalizeCompletionMessages(inputMessages)
	if err != nil {
		t.Fatalf("normalize messages: %v", err)
	}

	result, err := service.CreateCompletion(ctx, CompletionInput{
		RequestID: "req-non-stream",
		UserID:    user.ID,
		ModelID:   &modelID,
		Messages:  inputMessages,
	})
	if err != nil {
		t.Fatalf("create completion: %v", err)
	}

	expectedReserve := estimatePromptTokens(normalizedMessages) + 20
	if result.Billing.PreDeducted != expectedReserve {
		t.Fatalf("expected pre_deducted %d, got %d", expectedReserve, result.Billing.PreDeducted)
	}
	if result.Billing.FinalCharged != 11 {
		t.Fatalf("expected final_charged 11, got %d", result.Billing.FinalCharged)
	}
	if result.Billing.Refunded != expectedReserve-11 {
		t.Fatalf("expected refunded %d, got %d", expectedReserve-11, result.Billing.Refunded)
	}

	assertUserQuotaState(t, db, user.ID, 989, 11)
	assertRequestLogBilledQuota(t, db, "req-non-stream", 11, limits.RequestLogStatusCompleted)
	assertQuotaLogTypes(t, db, user.ID, map[account.QuotaLogType]int{
		account.QuotaLogTypePreDeduct:   1,
		account.QuotaLogTypeRefund:      1,
		account.QuotaLogTypeFinalCharge: 1,
	})
}

func TestStreamCompletionSettlesQuotaAndEmitsBilling(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openChatBillingTestDB(t)
	service, user, modelID := seedChatBillingFixture(t, ctx, db, newStreamingCompletionServer(t))

	inputMessages := []CompletionMessageInput{
		{Role: "user", Content: "Stream a short reply"},
	}

	normalizedMessages, _, err := normalizeCompletionMessages(inputMessages)
	if err != nil {
		t.Fatalf("normalize messages: %v", err)
	}

	var payloads []any
	err = service.StreamCompletion(ctx, CompletionInput{
		RequestID: "req-stream",
		UserID:    user.ID,
		ModelID:   &modelID,
		Messages:  inputMessages,
		Stream:    true,
	}, func(payload any) error {
		payloads = append(payloads, payload)
		return nil
	})
	if err != nil {
		t.Fatalf("stream completion: %v", err)
	}

	expectedReserve := estimatePromptTokens(normalizedMessages) + 20
	completed := findCompletedPayload(t, payloads)
	billing, ok := completed["billing"].(map[string]any)
	if !ok {
		t.Fatalf("expected billing payload map, got %#v", completed["billing"])
	}

	if toInt64(t, billing["pre_deducted"]) != expectedReserve {
		t.Fatalf("expected stream pre_deducted %d, got %v", expectedReserve, billing["pre_deducted"])
	}
	if toInt64(t, billing["final_charged"]) != 9 {
		t.Fatalf("expected stream final_charged 9, got %v", billing["final_charged"])
	}
	if toInt64(t, billing["refunded"]) != expectedReserve-9 {
		t.Fatalf("expected stream refunded %d, got %v", expectedReserve-9, billing["refunded"])
	}

	assertUserQuotaState(t, db, user.ID, 991, 9)
	assertRequestLogBilledQuota(t, db, "req-stream", 9, limits.RequestLogStatusCompleted)
	assertQuotaLogTypes(t, db, user.ID, map[account.QuotaLogType]int{
		account.QuotaLogTypePreDeduct:   1,
		account.QuotaLogTypeRefund:      1,
		account.QuotaLogTypeFinalCharge: 1,
	})
}

func openChatBillingTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	models := []any{
		&account.User{},
		&account.Auth{},
		&account.UserGroup{},
		&account.QuotaLog{},
		&catalog.Upstream{},
		&catalog.Channel{},
		&catalog.Model{},
		&catalog.ModelRouteBinding{},
		&Conversation{},
		&Message{},
		&limits.UserGroupModelLimitPolicy{},
		&limits.UserLimitAdjustment{},
		&limits.LLMRequestLog{},
	}
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	return db
}

func seedChatBillingFixture(t *testing.T, ctx context.Context, db *gorm.DB, server *httptest.Server) (*Service, *account.User, string) {
	t.Helper()
	t.Cleanup(server.Close)

	accountRepo := account.NewRepository(db)
	catalogRepo := catalog.NewRepository(db)
	limitsService := limits.NewService(limits.NewRepository(db), accountRepo)
	service := NewService(NewRepository(db), accountRepo, catalogRepo, limitsService)

	now := time.Now().UTC()
	user := &account.User{
		ID:          uuid.NewString(),
		Username:    "tester",
		Email:       "tester@example.com",
		DisplayName: "Tester",
		Role:        account.RoleUser,
		Status:      account.UserStatusActive,
		Quota:       1000,
		UsedQuota:   0,
		Settings: account.UserSettings{
			Timezone: "Asia/Shanghai",
			Locale:   "zh-CN",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.WithContext(ctx).Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	upstream, err := catalogRepo.CreateUpstream(ctx, catalog.CreateUpstreamInput{
		Name:             "local-upstream",
		BaseURL:          server.URL,
		AuthType:         "bearer",
		AuthConfig:       map[string]any{"api_key": "test-token"},
		Status:           string(catalog.UpstreamStatusActive),
		TimeoutSeconds:   10,
		CooldownSeconds:  30,
		FailureThreshold: 2,
	})
	if err != nil {
		t.Fatalf("create upstream: %v", err)
	}

	maxOutput := 20
	model, err := catalogRepo.CreateModel(ctx, catalog.CreateModelInput{
		ModelKey:        "test-model",
		DisplayName:     "Test Model",
		ProviderType:    "openai_compatible",
		ContextLength:   32000,
		MaxOutputTokens: &maxOutput,
		Status:          string(catalog.ModelStatusActive),
		RouteBindings: []catalog.RouteBindingInput{
			{
				UpstreamID: upstream.ID,
				Priority:   1,
				Status:     string(catalog.RouteBindingStatusActive),
			},
		},
	})
	if err != nil {
		t.Fatalf("create model: %v", err)
	}

	return service, user, model.Model.ID
}

func newCompletionServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"id":    "cmpl-test",
			"model": "test-model",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":              "assistant",
						"content":           "Settled response",
						"reasoning_content": "reserve then settle",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     5,
				"completion_tokens": 6,
				"total_tokens":      11,
			},
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("encode completion response: %v", err)
		}
	}))
}

func newStreamingCompletionServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			http.NotFound(w, r)
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatalf("response writer does not support flushing")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		chunks := []string{
			`data: {"id":"stream-test","model":"test-model","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}` + "\n\n",
			`data: {"id":"stream-test","model":"test-model","choices":[{"index":0,"delta":{"reasoning_content":"Thinking"},"finish_reason":null}]}` + "\n\n",
			`data: {"id":"stream-test","model":"test-model","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":4,"completion_tokens":5,"total_tokens":9}}` + "\n\n",
			"data: [DONE]\n\n",
		}

		for _, chunk := range chunks {
			if _, err := fmt.Fprint(w, chunk); err != nil {
				t.Fatalf("write stream chunk: %v", err)
			}
			flusher.Flush()
		}
	}))
}

func findCompletedPayload(t *testing.T, payloads []any) map[string]any {
	t.Helper()

	for _, payload := range payloads {
		item, ok := payload.(map[string]any)
		if !ok {
			continue
		}
		if item["type"] == "response.completed" {
			return item
		}
	}

	t.Fatalf("response.completed payload not found in %#v", payloads)
	return nil
}

func assertUserQuotaState(t *testing.T, db *gorm.DB, userID string, expectedQuota, expectedUsed int64) {
	t.Helper()

	var user account.User
	if err := db.First(&user, "id = ?", userID).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if user.Quota != expectedQuota {
		t.Fatalf("expected quota %d, got %d", expectedQuota, user.Quota)
	}
	if user.UsedQuota != expectedUsed {
		t.Fatalf("expected used_quota %d, got %d", expectedUsed, user.UsedQuota)
	}
}

func assertRequestLogBilledQuota(t *testing.T, db *gorm.DB, requestID string, expectedBilled int64, expectedStatus limits.RequestLogStatus) {
	t.Helper()

	var item limits.LLMRequestLog
	if err := db.First(&item, "request_id = ?", requestID).Error; err != nil {
		t.Fatalf("load request log: %v", err)
	}
	if item.BilledQuota != expectedBilled {
		t.Fatalf("expected billed_quota %d, got %d", expectedBilled, item.BilledQuota)
	}
	if item.Status != expectedStatus {
		t.Fatalf("expected request log status %s, got %s", expectedStatus, item.Status)
	}
}

func assertQuotaLogTypes(t *testing.T, db *gorm.DB, userID string, expected map[account.QuotaLogType]int) {
	t.Helper()

	var items []account.QuotaLog
	if err := db.Order("created_at ASC").Find(&items, "user_id = ?", userID).Error; err != nil {
		t.Fatalf("load quota logs: %v", err)
	}

	if len(items) != len(expected) {
		t.Fatalf("expected %d quota logs, got %d", len(expected), len(items))
	}

	counts := make(map[account.QuotaLogType]int, len(items))
	for _, item := range items {
		counts[item.LogType]++
		if item.RequestID == nil || strings.TrimSpace(*item.RequestID) == "" {
			t.Fatalf("expected request_id on quota log %+v", item)
		}
	}

	for logType, want := range expected {
		if counts[logType] != want {
			t.Fatalf("expected %d quota logs of type %s, got %d (all counts: %#v)", want, logType, counts[logType], counts)
		}
	}
}

func toInt64(t *testing.T, value any) int64 {
	t.Helper()

	switch item := value.(type) {
	case int:
		return int64(item)
	case int64:
		return item
	case float64:
		return int64(item)
	default:
		t.Fatalf("unsupported numeric type %T", value)
		return 0
	}
}
