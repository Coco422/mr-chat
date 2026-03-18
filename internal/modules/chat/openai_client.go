package chat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"mrchat/internal/modules/catalog"
)

type openAICompatibleClient struct{}

type openAIChatCompletionRequest struct {
	Model     string              `json:"model"`
	Messages  []openAIChatMessage `json:"messages"`
	Stream    bool                `json:"stream"`
	MaxTokens *int                `json:"max_tokens,omitempty"`
	Metadata  map[string]any      `json:"metadata,omitempty"`
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatCompletionResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role             string `json:"role"`
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage openAIUsage `json:"usage"`
}

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type openAIChatCompletionChunk struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role             string  `json:"role"`
			Content          *string `json:"content"`
			ReasoningContent *string `json:"reasoning_content"`
		} `json:"delta"`
		FinishReason *string      `json:"finish_reason"`
		Usage        *openAIUsage `json:"usage,omitempty"`
	} `json:"choices"`
	Usage *openAIUsage `json:"usage,omitempty"`
}

type openAIChatCompletionStream struct {
	body    io.ReadCloser
	scanner *bufio.Scanner
	cancel  context.CancelFunc
}

func (c *openAICompatibleClient) ChatCompletion(ctx context.Context, upstream *catalog.Upstream, payload openAIChatCompletionRequest) (*openAIChatCompletionResponse, error) {
	req, cancel, err := c.buildRequest(ctx, upstream, payload, "application/json")
	if err != nil {
		return nil, err
	}
	defer cancel()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call upstream: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, fmt.Errorf("read upstream response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("upstream status %d: %s", resp.StatusCode, compactResponseBody(responseBody))
	}

	var parsed openAIChatCompletionResponse
	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		return nil, fmt.Errorf("decode upstream response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("upstream returned no choices")
	}

	return &parsed, nil
}

func (c *openAICompatibleClient) OpenChatCompletionStream(ctx context.Context, upstream *catalog.Upstream, payload openAIChatCompletionRequest) (*openAIChatCompletionStream, error) {
	req, cancel, err := c.buildRequest(ctx, upstream, payload, "text/event-stream")
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call upstream: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		cancel()
		responseBody, readErr := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		if readErr != nil {
			return nil, fmt.Errorf("read upstream error response: %w", readErr)
		}
		return nil, fmt.Errorf("upstream status %d: %s", resp.StatusCode, compactResponseBody(responseBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 2<<20)

	return &openAIChatCompletionStream{
		body:    resp.Body,
		scanner: scanner,
		cancel:  cancel,
	}, nil
}

func (c *openAICompatibleClient) buildRequest(ctx context.Context, upstream *catalog.Upstream, payload openAIChatCompletionRequest, accept string) (*http.Request, context.CancelFunc, error) {
	if upstream == nil {
		return nil, nil, fmt.Errorf("missing upstream")
	}

	endpoint, err := buildChatCompletionsURL(upstream.BaseURL)
	if err != nil {
		return nil, nil, err
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal upstream payload: %w", err)
	}

	timeout := time.Duration(upstream.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	requestCtx, cancel := context.WithTimeout(ctx, timeout)

	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("build upstream request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", accept)

	if token := extractBearerToken(upstream.AuthConfig); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req, cancel, nil
}

func buildChatCompletionsURL(baseURL string) (string, error) {
	trimmed := strings.TrimSpace(baseURL)
	if trimmed == "" {
		return "", fmt.Errorf("base_url is empty")
	}
	if !strings.Contains(trimmed, "://") {
		trimmed = "http://" + trimmed
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("invalid base_url: %w", err)
	}

	path := strings.TrimRight(parsed.Path, "/")
	switch {
	case path == "":
		parsed.Path = "/v1/chat/completions"
	case strings.HasSuffix(path, "/v1"):
		parsed.Path = path + "/chat/completions"
	default:
		parsed.Path = path + "/v1/chat/completions"
	}

	return parsed.String(), nil
}

func extractBearerToken(authConfig map[string]any) string {
	if authConfig == nil {
		return ""
	}

	candidates := []string{"api_key", "token", "key"}
	for _, key := range candidates {
		value, ok := authConfig[key]
		if !ok {
			continue
		}
		token := strings.TrimSpace(fmt.Sprint(value))
		if token != "" && token != "<nil>" {
			return token
		}
	}

	return ""
}

func compactResponseBody(body []byte) string {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return "empty response"
	}
	if len(trimmed) <= 300 {
		return trimmed
	}
	return trimmed[:300]
}

func (r *openAIChatCompletionResponse) assistantContent() string {
	if r == nil || len(r.Choices) == 0 {
		return ""
	}
	return r.Choices[0].Message.Content
}

func (r *openAIChatCompletionResponse) reasoningContent() string {
	if r == nil || len(r.Choices) == 0 {
		return ""
	}
	return r.Choices[0].Message.ReasoningContent
}

func (r *openAIChatCompletionResponse) finishReason() string {
	if r == nil || len(r.Choices) == 0 {
		return ""
	}
	return r.Choices[0].FinishReason
}

func (s *openAIChatCompletionStream) Close() error {
	if s == nil {
		return nil
	}
	if s.cancel != nil {
		s.cancel()
	}
	if s.body == nil {
		return nil
	}
	return s.body.Close()
}

func (s *openAIChatCompletionStream) Next() (*openAIChatCompletionChunk, bool, error) {
	if s == nil || s.scanner == nil {
		return nil, true, nil
	}

	for s.scanner.Scan() {
		line := strings.TrimSpace(s.scanner.Text())
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}

		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" {
			continue
		}
		if payload == "[DONE]" {
			return nil, true, nil
		}

		var chunk openAIChatCompletionChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			return nil, false, fmt.Errorf("decode upstream stream chunk: %w", err)
		}
		return &chunk, false, nil
	}

	if err := s.scanner.Err(); err != nil {
		return nil, false, fmt.Errorf("read upstream stream: %w", err)
	}

	return nil, true, nil
}

func (c *openAIChatCompletionChunk) contentDelta() string {
	if c == nil || len(c.Choices) == 0 || c.Choices[0].Delta.Content == nil {
		return ""
	}
	return *c.Choices[0].Delta.Content
}

func (c *openAIChatCompletionChunk) reasoningDelta() string {
	if c == nil || len(c.Choices) == 0 || c.Choices[0].Delta.ReasoningContent == nil {
		return ""
	}
	return *c.Choices[0].Delta.ReasoningContent
}

func (c *openAIChatCompletionChunk) finishReason() string {
	if c == nil || len(c.Choices) == 0 || c.Choices[0].FinishReason == nil {
		return ""
	}
	return strings.TrimSpace(*c.Choices[0].FinishReason)
}

func (c *openAIChatCompletionChunk) usage() openAIUsage {
	if c == nil {
		return openAIUsage{}
	}
	if c.Usage != nil {
		return *c.Usage
	}
	if len(c.Choices) > 0 && c.Choices[0].Usage != nil {
		return *c.Choices[0].Usage
	}
	return openAIUsage{}
}
