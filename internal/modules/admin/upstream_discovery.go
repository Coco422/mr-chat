package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"mrchat/internal/modules/catalog"
)

type UpstreamModelDiscoveryResult struct {
	Upstream  *catalog.Upstream      `json:"upstream"`
	Items     []DiscoveredModel      `json:"items"`
	FetchedAt string                 `json:"fetched_at"`
	Summary   map[string]any         `json:"summary"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type DiscoveredModel struct {
	ModelKey               string                `json:"model_key"`
	DisplayName            string                `json:"display_name"`
	ProviderType           string                `json:"provider_type"`
	Object                 string                `json:"object,omitempty"`
	OwnedBy                string                `json:"owned_by,omitempty"`
	SupportedEndpointTypes []string              `json:"supported_endpoint_types,omitempty"`
	Capabilities           map[string]any        `json:"capabilities"`
	Raw                    map[string]any        `json:"raw"`
	AlreadyImported        bool                  `json:"already_imported"`
	ExistingModel          *ImportedModelSummary `json:"existing_model,omitempty"`
}

type ImportedModelSummary struct {
	ID          string              `json:"id"`
	ModelKey    string              `json:"model_key"`
	DisplayName string              `json:"display_name"`
	Status      catalog.ModelStatus `json:"status"`
}

type openAIModelsResponse struct {
	Object string                   `json:"object"`
	Data   []openAIModelDescription `json:"data"`
}

type openAIModelDescription struct {
	ID                     string         `json:"id"`
	Object                 string         `json:"object"`
	OwnedBy                string         `json:"owned_by"`
	SupportedEndpointTypes []string       `json:"supported_endpoint_types"`
	Metadata               map[string]any `json:"-"`
}

func (c *openAIModelDescription) UnmarshalJSON(data []byte) error {
	type alias openAIModelDescription
	var parsed alias
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}

	raw := map[string]any{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*c = openAIModelDescription(parsed)
	c.Metadata = raw
	return nil
}

type upstreamDiscoveryClient struct{}

func (c *upstreamDiscoveryClient) DiscoverModels(ctx context.Context, upstream *catalog.Upstream) ([]DiscoveredModel, error) {
	if upstream == nil {
		return nil, fmt.Errorf("missing upstream")
	}

	switch strings.TrimSpace(upstream.ProviderType) {
	case "", "openai", "openai_compatible":
		return c.discoverOpenAICompatibleModels(ctx, upstream)
	default:
		return nil, ErrUpstreamDiscoveryUnsupported
	}
}

func (c *upstreamDiscoveryClient) discoverOpenAICompatibleModels(ctx context.Context, upstream *catalog.Upstream) ([]DiscoveredModel, error) {
	endpoint, err := buildModelsURL(upstream.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUpstreamDiscoveryFailed, err)
	}

	timeout := time.Duration(upstream.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	requestCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %v", ErrUpstreamDiscoveryFailed, err)
	}
	req.Header.Set("Accept", "application/json")
	if token := extractBearerToken(upstream.AuthConfig); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: call upstream: %v", ErrUpstreamDiscoveryFailed, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, fmt.Errorf("%w: read response: %v", ErrUpstreamDiscoveryFailed, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: upstream status %d: %s", ErrUpstreamDiscoveryFailed, resp.StatusCode, compactResponseBody(body))
	}

	var parsed openAIModelsResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("%w: decode response: %v", ErrUpstreamDiscoveryFailed, err)
	}

	items := make([]DiscoveredModel, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		modelKey := strings.TrimSpace(item.ID)
		if modelKey == "" {
			continue
		}

		supportedEndpointTypes := sanitizeStringSlice(item.SupportedEndpointTypes)
		if len(supportedEndpointTypes) == 0 {
			supportedEndpointTypes = []string{"openai"}
		}

		items = append(items, DiscoveredModel{
			ModelKey:               modelKey,
			DisplayName:            modelKey,
			ProviderType:           defaultProviderType(upstream.ProviderType),
			Object:                 strings.TrimSpace(item.Object),
			OwnedBy:                strings.TrimSpace(item.OwnedBy),
			SupportedEndpointTypes: supportedEndpointTypes,
			Capabilities: map[string]any{
				"chat":      slices.Contains(supportedEndpointTypes, "openai"),
				"streaming": true,
			},
			Raw: nonNilMap(item.Metadata),
		})
	}

	slices.SortFunc(items, func(a, b DiscoveredModel) int {
		return strings.Compare(strings.ToLower(a.ModelKey), strings.ToLower(b.ModelKey))
	})

	return items, nil
}

func buildModelsURL(baseURL string) (string, error) {
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
		parsed.Path = "/v1/models"
	case strings.HasSuffix(path, "/v1"):
		parsed.Path = path + "/models"
	default:
		parsed.Path = path + "/v1/models"
	}

	return parsed.String(), nil
}

func defaultProviderType(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "openai_compatible"
	}
	return trimmed
}

func sanitizeStringSlice(items []string) []string {
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func extractBearerToken(authConfig map[string]any) string {
	if authConfig == nil {
		return ""
	}

	for _, key := range []string{"api_key", "token", "key"} {
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

func nonNilMap(value map[string]any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	return value
}
