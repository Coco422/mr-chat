package catalog

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrUpstreamNotFound = errors.New("upstream not found")
	ErrModelNotFound    = errors.New("model not found")
	ErrChannelNotFound  = errors.New("channel not found")
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) DB() *gorm.DB {
	return r.db
}

func (r *Repository) ListUpstreams(ctx context.Context) ([]Upstream, error) {
	var items []Upstream
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list upstreams: %w", err)
	}

	return items, nil
}

func (r *Repository) ListChannels(ctx context.Context) ([]Channel, error) {
	var items []Channel
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}

	return items, nil
}

func (r *Repository) CreateChannel(ctx context.Context, input CreateChannelInput) (*Channel, error) {
	return r.CreateChannelWithDB(ctx, r.db, input)
}

func (r *Repository) CreateChannelWithDB(ctx context.Context, db *gorm.DB, input CreateChannelInput) (*Channel, error) {
	if db == nil {
		db = r.db
	}

	now := time.Now().UTC()
	item := &Channel{
		ID:            uuid.NewString(),
		Name:          strings.TrimSpace(input.Name),
		Description:   sanitizeOptionalString(input.Description),
		Status:        ChannelStatus(defaultString(input.Status, string(ChannelStatusActive))),
		BillingConfig: nonNilMap(input.BillingConfig),
		Metadata:      nonNilMap(input.Metadata),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := db.WithContext(ctx).Create(item).Error; err != nil {
		return nil, fmt.Errorf("create channel: %w", err)
	}

	return item, nil
}

func (r *Repository) UpdateChannel(ctx context.Context, channelID string, input UpdateChannelInput) (*Channel, error) {
	return r.UpdateChannelWithDB(ctx, r.db, channelID, input)
}

func (r *Repository) UpdateChannelWithDB(ctx context.Context, db *gorm.DB, channelID string, input UpdateChannelInput) (*Channel, error) {
	if db == nil {
		db = r.db
	}

	item, err := r.getChannelByID(ctx, db, channelID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		item.Name = strings.TrimSpace(*input.Name)
	}
	if input.Description != nil {
		item.Description = sanitizeOptionalString(input.Description)
	}
	if input.Status != nil {
		item.Status = ChannelStatus(strings.TrimSpace(*input.Status))
	}
	if input.BillingConfig != nil {
		item.BillingConfig = nonNilMap(input.BillingConfig)
	}
	if input.Metadata != nil {
		item.Metadata = nonNilMap(input.Metadata)
	}

	item.UpdatedAt = time.Now().UTC()
	if err := db.WithContext(ctx).Save(item).Error; err != nil {
		return nil, fmt.Errorf("update channel: %w", err)
	}

	return item, nil
}

func (r *Repository) CreateUpstream(ctx context.Context, input CreateUpstreamInput) (*Upstream, error) {
	return r.CreateUpstreamWithDB(ctx, r.db, input)
}

func (r *Repository) CreateUpstreamWithDB(ctx context.Context, db *gorm.DB, input CreateUpstreamInput) (*Upstream, error) {
	if db == nil {
		db = r.db
	}

	now := time.Now().UTC()
	item := &Upstream{
		ID:               uuid.NewString(),
		Name:             strings.TrimSpace(input.Name),
		ProviderType:     defaultString(input.ProviderType, "openai_compatible"),
		BaseURL:          strings.TrimSpace(input.BaseURL),
		AuthType:         defaultString(input.AuthType, "bearer"),
		AuthConfig:       nonNilMap(input.AuthConfig),
		Status:           UpstreamStatus(defaultString(input.Status, string(UpstreamStatusActive))),
		TimeoutSeconds:   defaultInt(input.TimeoutSeconds, 60),
		CooldownSeconds:  defaultInt(input.CooldownSeconds, 60),
		FailureThreshold: defaultInt(input.FailureThreshold, 3),
		Metadata:         nonNilMap(input.Metadata),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := db.WithContext(ctx).Create(item).Error; err != nil {
		return nil, fmt.Errorf("create upstream: %w", err)
	}

	return item, nil
}

func (r *Repository) UpdateUpstream(ctx context.Context, upstreamID string, input UpdateUpstreamInput) (*Upstream, error) {
	return r.UpdateUpstreamWithDB(ctx, r.db, upstreamID, input)
}

func (r *Repository) UpdateUpstreamWithDB(ctx context.Context, db *gorm.DB, upstreamID string, input UpdateUpstreamInput) (*Upstream, error) {
	if db == nil {
		db = r.db
	}

	item, err := r.getUpstreamByID(ctx, db, upstreamID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		item.Name = strings.TrimSpace(*input.Name)
	}
	if input.ProviderType != nil {
		item.ProviderType = strings.TrimSpace(*input.ProviderType)
	}
	if input.BaseURL != nil {
		item.BaseURL = strings.TrimSpace(*input.BaseURL)
	}
	if input.AuthType != nil {
		item.AuthType = strings.TrimSpace(*input.AuthType)
	}
	if input.AuthConfig != nil {
		item.AuthConfig = nonNilMap(input.AuthConfig)
	}
	if input.Status != nil {
		item.Status = UpstreamStatus(strings.TrimSpace(*input.Status))
	}
	if input.TimeoutSeconds != nil {
		item.TimeoutSeconds = *input.TimeoutSeconds
	}
	if input.CooldownSeconds != nil {
		item.CooldownSeconds = *input.CooldownSeconds
	}
	if input.FailureThreshold != nil {
		item.FailureThreshold = *input.FailureThreshold
	}
	if input.Metadata != nil {
		item.Metadata = nonNilMap(input.Metadata)
	}

	item.UpdatedAt = time.Now().UTC()
	if err := db.WithContext(ctx).Save(item).Error; err != nil {
		return nil, fmt.Errorf("update upstream: %w", err)
	}

	return item, nil
}

func (r *Repository) ListModels(ctx context.Context) ([]ModelWithBindings, error) {
	var models []Model
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}

	return r.attachBindings(ctx, r.db, models)
}

func (r *Repository) ListActiveModels(ctx context.Context) ([]ModelWithBindings, error) {
	var models []Model
	if err := r.db.WithContext(ctx).
		Where("status = ?", ModelStatusActive).
		Order("display_name ASC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list active models: %w", err)
	}

	return r.attachBindings(ctx, r.db, models)
}

func (r *Repository) CreateModel(ctx context.Context, input CreateModelInput) (*ModelWithBindings, error) {
	return r.CreateModelWithDB(ctx, r.db, input)
}

func (r *Repository) CreateModelWithDB(ctx context.Context, db *gorm.DB, input CreateModelInput) (*ModelWithBindings, error) {
	if db == nil {
		db = r.db
	}

	now := time.Now().UTC()
	item := &Model{
		ID:                  uuid.NewString(),
		ModelKey:            strings.TrimSpace(input.ModelKey),
		DisplayName:         strings.TrimSpace(input.DisplayName),
		ProviderType:        defaultString(input.ProviderType, "openai_compatible"),
		ContextLength:       input.ContextLength,
		MaxOutputTokens:     input.MaxOutputTokens,
		Pricing:             nonNilMap(input.Pricing),
		Capabilities:        nonNilMap(input.Capabilities),
		VisibleUserGroupIDs: sanitizeStrings(input.VisibleUserGroupIDs),
		Status:              ModelStatus(defaultString(input.Status, string(ModelStatusActive))),
		Metadata:            nonNilMap(input.Metadata),
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(item).Error; err != nil {
			return fmt.Errorf("create model: %w", err)
		}
		if err := r.replaceRouteBindings(ctx, tx, item.ID, input.RouteBindings); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	result, err := r.getModelWithBindingsByID(ctx, db, item.ID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *Repository) UpdateModel(ctx context.Context, modelID string, input UpdateModelInput) (*ModelWithBindings, error) {
	return r.UpdateModelWithDB(ctx, r.db, modelID, input)
}

func (r *Repository) UpdateModelWithDB(ctx context.Context, db *gorm.DB, modelID string, input UpdateModelInput) (*ModelWithBindings, error) {
	if db == nil {
		db = r.db
	}

	existing, err := r.getModelRowByID(ctx, db, modelID)
	if err != nil {
		return nil, err
	}

	if input.ModelKey != nil {
		existing.ModelKey = strings.TrimSpace(*input.ModelKey)
	}
	if input.DisplayName != nil {
		existing.DisplayName = strings.TrimSpace(*input.DisplayName)
	}
	if input.ProviderType != nil {
		existing.ProviderType = strings.TrimSpace(*input.ProviderType)
	}
	if input.ContextLength != nil {
		existing.ContextLength = *input.ContextLength
	}
	if input.MaxOutputTokens != nil {
		existing.MaxOutputTokens = input.MaxOutputTokens
	}
	if input.Pricing != nil {
		existing.Pricing = nonNilMap(input.Pricing)
	}
	if input.Capabilities != nil {
		existing.Capabilities = nonNilMap(input.Capabilities)
	}
	if input.VisibleUserGroupIDs != nil {
		existing.VisibleUserGroupIDs = sanitizeStrings(input.VisibleUserGroupIDs)
	}
	if input.Status != nil {
		existing.Status = ModelStatus(strings.TrimSpace(*input.Status))
	}
	if input.Metadata != nil {
		existing.Metadata = nonNilMap(input.Metadata)
	}

	existing.UpdatedAt = time.Now().UTC()
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(existing).Error; err != nil {
			return fmt.Errorf("update model: %w", err)
		}
		if err := r.replaceRouteBindings(ctx, tx, existing.ID, input.RouteBindings); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return r.getModelWithBindingsByID(ctx, db, modelID)
}

func (r *Repository) GetModelByID(ctx context.Context, modelID string) (*ModelWithBindings, error) {
	return r.getModelWithBindingsByID(ctx, r.db, modelID)
}

func (r *Repository) attachBindings(ctx context.Context, db *gorm.DB, models []Model) ([]ModelWithBindings, error) {
	if len(models) == 0 {
		return []ModelWithBindings{}, nil
	}

	modelIDs := make([]string, 0, len(models))
	for _, item := range models {
		modelIDs = append(modelIDs, item.ID)
	}

	var bindings []ModelRouteBinding
	if err := db.WithContext(ctx).
		Where("model_id IN ?", modelIDs).
		Order("priority ASC").
		Find(&bindings).Error; err != nil {
		return nil, fmt.Errorf("list model route bindings: %w", err)
	}

	grouped := make(map[string][]ModelRouteBinding, len(modelIDs))
	for _, binding := range bindings {
		grouped[binding.ModelID] = append(grouped[binding.ModelID], binding)
	}

	result := make([]ModelWithBindings, 0, len(models))
	for _, item := range models {
		result = append(result, ModelWithBindings{
			Model:         item,
			RouteBindings: grouped[item.ID],
		})
	}

	return result, nil
}

func (r *Repository) getModelWithBindingsByID(ctx context.Context, db *gorm.DB, modelID string) (*ModelWithBindings, error) {
	modelRow, err := r.getModelRowByID(ctx, db, modelID)
	if err != nil {
		return nil, err
	}

	items, err := r.attachBindings(ctx, db, []Model{*modelRow})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, ErrModelNotFound
	}

	return &items[0], nil
}

func (r *Repository) replaceRouteBindings(ctx context.Context, db *gorm.DB, modelID string, bindings []RouteBindingInput) error {
	if err := db.WithContext(ctx).
		Where("model_id = ?", modelID).
		Delete(&ModelRouteBinding{}).Error; err != nil {
		return fmt.Errorf("clear route bindings: %w", err)
	}

	if len(bindings) == 0 {
		return nil
	}

	now := time.Now().UTC()
	rows := make([]ModelRouteBinding, 0, len(bindings))
	for _, binding := range bindings {
		channelID := sanitizeOptionalString(binding.ChannelID)
		rows = append(rows, ModelRouteBinding{
			ID:         uuid.NewString(),
			ModelID:    modelID,
			ChannelID:  channelID,
			UpstreamID: strings.TrimSpace(binding.UpstreamID),
			Priority:   defaultInt(binding.Priority, 1),
			Status:     RouteBindingStatus(defaultString(binding.Status, string(RouteBindingStatusActive))),
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}

	if err := db.WithContext(ctx).Create(&rows).Error; err != nil {
		return fmt.Errorf("create route bindings: %w", err)
	}

	return nil
}

func (r *Repository) getUpstreamByID(ctx context.Context, db *gorm.DB, upstreamID string) (*Upstream, error) {
	var item Upstream
	if err := db.WithContext(ctx).First(&item, "id = ?", upstreamID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUpstreamNotFound
		}
		return nil, fmt.Errorf("get upstream by id: %w", err)
	}
	return &item, nil
}

func (r *Repository) getChannelByID(ctx context.Context, db *gorm.DB, channelID string) (*Channel, error) {
	var item Channel
	if err := db.WithContext(ctx).First(&item, "id = ?", channelID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChannelNotFound
		}
		return nil, fmt.Errorf("get channel by id: %w", err)
	}
	return &item, nil
}

func (r *Repository) getModelRowByID(ctx context.Context, db *gorm.DB, modelID string) (*Model, error) {
	var item Model
	if err := db.WithContext(ctx).First(&item, "id = ?", modelID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrModelNotFound
		}
		return nil, fmt.Errorf("get model by id: %w", err)
	}
	return &item, nil
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func defaultInt(value, fallback int) int {
	if value == 0 {
		return fallback
	}
	return value
}

func nonNilMap(value map[string]any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	return value
}

func sanitizeStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}

	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || slices.Contains(result, trimmed) {
			continue
		}
		result = append(result, trimmed)
	}
	return result
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
