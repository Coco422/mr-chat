package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mrchat/internal/modules/account"
	"mrchat/internal/modules/catalog"
	"mrchat/internal/modules/limits"
)

func TestCreateCompletionSkipsCoolingUpstreamAndRecovers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openChatBillingTestDB(t)

	var primaryCalls atomic.Int64
	primaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call := primaryCalls.Add(1)
		if r.URL.Path != "/v1/chat/completions" {
			http.NotFound(w, r)
			return
		}

		if call == 1 {
			http.Error(w, "primary unavailable", http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"id":    fmt.Sprintf("primary-%d", call),
			"model": "test-model",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": "primary recovered",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     4,
				"completion_tokens": 3,
				"total_tokens":      7,
			},
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("encode primary response: %v", err)
		}
	}))
	t.Cleanup(primaryServer.Close)

	var secondaryCalls atomic.Int64
	secondaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secondaryCalls.Add(1)
		if r.URL.Path != "/v1/chat/completions" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"id":    "secondary-success",
			"model": "test-model",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": "secondary fallback",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     4,
				"completion_tokens": 2,
				"total_tokens":      6,
			},
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("encode secondary response: %v", err)
		}
	}))
	t.Cleanup(secondaryServer.Close)

	service, user, modelID, primaryUpstreamID, secondaryUpstreamID := seedChatFailoverFixture(
		t,
		ctx,
		db,
		failoverFixtureInput{
			PrimaryBaseURL:       primaryServer.URL,
			SecondaryBaseURL:     secondaryServer.URL,
			PrimaryCooldown:      1,
			PrimaryFailThreshold: 1,
		},
	)

	first, err := service.CreateCompletion(ctx, CompletionInput{
		RequestID: "req-failover-1",
		UserID:    user.ID,
		ModelID:   &modelID,
		Messages: []CompletionMessageInput{
			{Role: "user", Content: "first request"},
		},
	})
	if err != nil {
		t.Fatalf("first completion: %v", err)
	}
	if first.Message.Content != "secondary fallback" {
		t.Fatalf("expected fallback response, got %q", first.Message.Content)
	}
	assertAttemptSequence(t, db, "req-failover-1", []string{primaryUpstreamID, secondaryUpstreamID})

	second, err := service.CreateCompletion(ctx, CompletionInput{
		RequestID: "req-failover-2",
		UserID:    user.ID,
		ModelID:   &modelID,
		Messages: []CompletionMessageInput{
			{Role: "user", Content: "second request"},
		},
	})
	if err != nil {
		t.Fatalf("second completion: %v", err)
	}
	if second.Message.Content != "secondary fallback" {
		t.Fatalf("expected cooldown fallback response, got %q", second.Message.Content)
	}

	if got := primaryCalls.Load(); got != 1 {
		t.Fatalf("expected primary upstream to stay in cooldown on second request, got %d calls", got)
	}
	if got := secondaryCalls.Load(); got != 2 {
		t.Fatalf("expected secondary upstream to handle two requests, got %d calls", got)
	}
	assertAttemptSequence(t, db, "req-failover-2", []string{primaryUpstreamID, secondaryUpstreamID})

	time.Sleep(1100 * time.Millisecond)

	third, err := service.CreateCompletion(ctx, CompletionInput{
		RequestID: "req-failover-3",
		UserID:    user.ID,
		ModelID:   &modelID,
		Messages: []CompletionMessageInput{
			{Role: "user", Content: "third request"},
		},
	})
	if err != nil {
		t.Fatalf("third completion: %v", err)
	}
	if third.Message.Content != "primary recovered" {
		t.Fatalf("expected recovered primary response, got %q", third.Message.Content)
	}

	if got := primaryCalls.Load(); got != 2 {
		t.Fatalf("expected primary upstream to be retried after cooldown, got %d calls", got)
	}
	if got := secondaryCalls.Load(); got != 2 {
		t.Fatalf("expected secondary upstream to remain at two calls after recovery, got %d calls", got)
	}
	assertAttemptSequence(t, db, "req-failover-3", []string{primaryUpstreamID})
}

func TestStreamCompletionSkipsCoolingUpstream(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openChatBillingTestDB(t)

	var primaryCalls atomic.Int64
	primaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		primaryCalls.Add(1)
		if r.URL.Path != "/v1/chat/completions" {
			http.NotFound(w, r)
			return
		}

		http.Error(w, "primary unavailable", http.StatusBadGateway)
	}))
	t.Cleanup(primaryServer.Close)

	var secondaryCalls atomic.Int64
	secondaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secondaryCalls.Add(1)
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
			`data: {"id":"secondary-stream","model":"test-model","choices":[{"index":0,"delta":{"content":"fallback"},"finish_reason":null}]}` + "\n\n",
			`data: {"id":"secondary-stream","model":"test-model","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":4,"completion_tokens":5,"total_tokens":9}}` + "\n\n",
			"data: [DONE]\n\n",
		}
		for _, chunk := range chunks {
			if _, err := fmt.Fprint(w, chunk); err != nil {
				t.Fatalf("write stream chunk: %v", err)
			}
			flusher.Flush()
		}
	}))
	t.Cleanup(secondaryServer.Close)

	service, user, modelID, primaryUpstreamID, secondaryUpstreamID := seedChatFailoverFixture(
		t,
		ctx,
		db,
		failoverFixtureInput{
			PrimaryBaseURL:       primaryServer.URL,
			SecondaryBaseURL:     secondaryServer.URL,
			PrimaryCooldown:      1,
			PrimaryFailThreshold: 1,
		},
	)

	runStreamRequest := func(requestID string) {
		t.Helper()

		var payloads []any
		err := service.StreamCompletion(ctx, CompletionInput{
			RequestID: requestID,
			UserID:    user.ID,
			ModelID:   &modelID,
			Messages: []CompletionMessageInput{
				{Role: "user", Content: "stream request"},
			},
			Stream: true,
		}, func(payload any) error {
			payloads = append(payloads, payload)
			return nil
		})
		if err != nil {
			t.Fatalf("stream completion %s: %v", requestID, err)
		}

		completed := findCompletedPayload(t, payloads)
		if usage, ok := completed["usage"].(map[string]any); !ok || toInt64(t, usage["total_tokens"]) != 9 {
			t.Fatalf("expected completed usage payload for %s, got %#v", requestID, completed["usage"])
		}
	}

	runStreamRequest("req-stream-failover-1")
	assertAttemptSequence(t, db, "req-stream-failover-1", []string{primaryUpstreamID, secondaryUpstreamID})

	runStreamRequest("req-stream-failover-2")
	assertAttemptSequence(t, db, "req-stream-failover-2", []string{primaryUpstreamID, secondaryUpstreamID})

	if got := primaryCalls.Load(); got != 1 {
		t.Fatalf("expected primary upstream to be skipped during cooldown on second stream request, got %d calls", got)
	}
	if got := secondaryCalls.Load(); got != 2 {
		t.Fatalf("expected secondary upstream to handle both stream requests, got %d calls", got)
	}
}

type failoverFixtureInput struct {
	PrimaryBaseURL       string
	SecondaryBaseURL     string
	PrimaryCooldown      int
	PrimaryFailThreshold int
}

func seedChatFailoverFixture(t *testing.T, ctx context.Context, db *gorm.DB, input failoverFixtureInput) (*Service, *account.User, string, string, string) {
	t.Helper()

	accountRepo := account.NewRepository(db)
	catalogRepo := catalog.NewRepository(db)
	limitsService := limits.NewService(limits.NewRepository(db), accountRepo)
	service := NewService(NewRepository(db), accountRepo, catalogRepo, limitsService)

	now := time.Now().UTC()
	user := &account.User{
		ID:          uuid.NewString(),
		Username:    "failover-tester",
		Email:       "failover@example.com",
		DisplayName: "Failover Tester",
		Role:        account.RoleUser,
		Status:      account.UserStatusActive,
		Quota:       10000,
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

	primary, err := catalogRepo.CreateUpstream(ctx, catalog.CreateUpstreamInput{
		Name:             "primary-upstream",
		BaseURL:          input.PrimaryBaseURL,
		AuthType:         "bearer",
		AuthConfig:       map[string]any{"api_key": "primary-token"},
		Status:           string(catalog.UpstreamStatusActive),
		TimeoutSeconds:   10,
		CooldownSeconds:  input.PrimaryCooldown,
		FailureThreshold: input.PrimaryFailThreshold,
	})
	if err != nil {
		t.Fatalf("create primary upstream: %v", err)
	}

	secondary, err := catalogRepo.CreateUpstream(ctx, catalog.CreateUpstreamInput{
		Name:             "secondary-upstream",
		BaseURL:          input.SecondaryBaseURL,
		AuthType:         "bearer",
		AuthConfig:       map[string]any{"api_key": "secondary-token"},
		Status:           string(catalog.UpstreamStatusActive),
		TimeoutSeconds:   10,
		CooldownSeconds:  1,
		FailureThreshold: 1,
	})
	if err != nil {
		t.Fatalf("create secondary upstream: %v", err)
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
				UpstreamID: primary.ID,
				Priority:   1,
				Status:     string(catalog.RouteBindingStatusActive),
			},
			{
				UpstreamID: secondary.ID,
				Priority:   2,
				Status:     string(catalog.RouteBindingStatusActive),
			},
		},
	})
	if err != nil {
		t.Fatalf("create model: %v", err)
	}

	return service, user, model.Model.ID, primary.ID, secondary.ID
}

func assertAttemptSequence(t *testing.T, db *gorm.DB, requestID string, upstreamIDs []string) {
	t.Helper()

	var item limits.LLMRequestLog
	if err := db.First(&item, "request_id = ?", requestID).Error; err != nil {
		t.Fatalf("load request log for %s: %v", requestID, err)
	}

	rawAttempts, ok := item.Metadata["attempts"]
	if !ok {
		t.Fatalf("expected attempts metadata for %s, got %#v", requestID, item.Metadata)
	}

	attempts, ok := rawAttempts.([]any)
	if !ok {
		t.Fatalf("expected attempts slice for %s, got %T", requestID, rawAttempts)
	}
	if len(attempts) != len(upstreamIDs) {
		t.Fatalf("expected %d attempts for %s, got %d", len(upstreamIDs), requestID, len(attempts))
	}

	for index, expectedUpstreamID := range upstreamIDs {
		attempt, ok := attempts[index].(map[string]any)
		if !ok {
			t.Fatalf("expected attempt metadata map at index %d for %s, got %T", index, requestID, attempts[index])
		}
		got := strings.TrimSpace(fmt.Sprint(attempt["upstream_id"]))
		if got != expectedUpstreamID {
			t.Fatalf("expected upstream %s at attempt %d for %s, got %s", expectedUpstreamID, index, requestID, got)
		}
	}
}
