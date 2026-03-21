package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"mrchat/internal/modules/account"
	"mrchat/internal/modules/audit"
	"mrchat/internal/modules/catalog"
	"mrchat/internal/modules/limits"
)

var (
	ErrQuotaWouldBecomeNegative     = errors.New("quota would become negative")
	ErrUpstreamDiscoveryFailed      = errors.New("upstream discovery failed")
	ErrUpstreamDiscoveryUnsupported = errors.New("upstream discovery unsupported")
	ErrModelImportInvalid           = errors.New("model import invalid")
)

type Service struct {
	accountRepo   *account.Repository
	catalogRepo   *catalog.Repository
	limitsService *limits.Service
	auditRepo     *audit.Repository
	discovery     *upstreamDiscoveryClient
}

type ActorContext struct {
	ActorUserID string
	ActorRole   account.Role
	RequestID   string
	IPAddress   string
	UserAgent   string
}

type ListUsersFilter struct {
	Page     int
	PageSize int
	Keyword  string
	Status   string
}

type AdminUserRecord struct {
	User      account.User
	UserGroup *account.UserGroup
}

type UserListResult struct {
	Items []AdminUserRecord
	Total int64
}

type AdminModelRouteBindingView struct {
	Binding  catalog.ModelRouteBinding
	Channel  *catalog.Channel
	Upstream *catalog.Upstream
}

type AdminModelView struct {
	Item               catalog.ModelWithBindings
	VisibleUserGroups  []account.UserGroup
	HydratedBindings   []AdminModelRouteBindingView
	VisibilitySummary  string
	RouteRuleSummaries []string
}

type CreateUpstreamInput struct {
	Name             string
	ProviderType     string
	BaseURL          string
	AuthType         string
	AuthConfig       map[string]any
	Status           string
	TimeoutSeconds   int
	CooldownSeconds  int
	FailureThreshold int
	Metadata         map[string]any
}

type UpdateUpstreamInput = catalog.UpdateUpstreamInput
type CreateModelInput = catalog.CreateModelInput
type UpdateModelInput = catalog.UpdateModelInput
type CreateChannelInput = catalog.CreateChannelInput
type UpdateChannelInput = catalog.UpdateChannelInput
type PolicyUpsertInput = limits.PolicyUpsertInput

type CreateUserGroupInput struct {
	Name        string
	Description *string
	Status      account.UserGroupStatus
	Permissions map[string]any
	Metadata    map[string]any
}

type UpdateUserGroupInput = account.UpdateUserGroupInput

type CreateUserLimitAdjustmentInput struct {
	ModelID    *string
	MetricType limits.MetricType
	WindowType limits.WindowType
	Delta      int64
	Reason     *string
}

type ImportModelsInput struct {
	UpstreamID string
	Items      []ImportModelItemInput
}

type ImportModelItemInput struct {
	ModelKey            string
	DisplayName         string
	ProviderType        string
	ContextLength       int
	MaxOutputTokens     *int
	Pricing             map[string]any
	Capabilities        map[string]any
	VisibleUserGroupIDs []string
	Status              string
	Metadata            map[string]any
	ChannelID           *string
	Priority            int
}

type ImportedModelRecord struct {
	RequestedModelKey string
	Model             *AdminModelView
	ExistingModel     *ImportedModelSummary
	Status            string
}

type ImportModelsResult struct {
	Upstream *catalog.Upstream
	Items    []ImportedModelRecord
	Summary  map[string]any
}

func NewService(accountRepo *account.Repository, catalogRepo *catalog.Repository, limitsService *limits.Service, auditRepo *audit.Repository) *Service {
	return &Service{
		accountRepo:   accountRepo,
		catalogRepo:   catalogRepo,
		limitsService: limitsService,
		auditRepo:     auditRepo,
		discovery:     &upstreamDiscoveryClient{},
	}
}

func (s *Service) ListUpstreams(ctx context.Context) ([]catalog.Upstream, error) {
	return s.catalogRepo.ListUpstreams(ctx)
}

func (s *Service) GetUpstream(ctx context.Context, upstreamID string) (*catalog.Upstream, error) {
	return s.catalogRepo.GetUpstreamByID(ctx, upstreamID)
}

func (s *Service) DiscoverUpstreamModels(ctx context.Context, actor ActorContext, upstreamID string) (*UpstreamModelDiscoveryResult, error) {
	upstream, err := s.catalogRepo.GetUpstreamByID(ctx, upstreamID)
	if err != nil {
		return nil, err
	}

	items, err := s.discovery.DiscoverModels(ctx, upstream)
	if err != nil {
		return nil, err
	}

	existingModels, err := s.catalogRepo.ListModels(ctx)
	if err != nil {
		return nil, err
	}

	existingByKey := make(map[string]catalog.ModelWithBindings, len(existingModels))
	for _, item := range existingModels {
		existingByKey[strings.ToLower(strings.TrimSpace(item.Model.ModelKey))] = item
	}

	importedCount := 0
	for index := range items {
		modelKey := strings.ToLower(strings.TrimSpace(items[index].ModelKey))
		existing, ok := existingByKey[modelKey]
		if !ok {
			continue
		}

		importedCount++
		items[index].AlreadyImported = true
		items[index].ExistingModel = &ImportedModelSummary{
			ID:          existing.Model.ID,
			ModelKey:    existing.Model.ModelKey,
			DisplayName: existing.Model.DisplayName,
			Status:      existing.Model.Status,
		}
	}

	_ = s.auditRepo.Create(ctx, audit.CreateInput{
		ActorUserID:  optionalString(actor.ActorUserID),
		ActorRole:    optionalString(string(actor.ActorRole)),
		Action:       "admin.upstream.discover_models",
		ResourceType: "upstream",
		ResourceID:   optionalString(upstream.ID),
		RequestID:    optionalString(actor.RequestID),
		IPAddress:    optionalString(actor.IPAddress),
		UserAgent:    optionalString(actor.UserAgent),
		Result:       audit.ResultSuccess,
		Details: map[string]any{
			"upstream_name":    upstream.Name,
			"provider_type":    upstream.ProviderType,
			"discovered":       len(items),
			"already_imported": importedCount,
		},
	})

	return &UpstreamModelDiscoveryResult{
		Upstream:  upstream,
		Items:     items,
		FetchedAt: time.Now().UTC().Format(time.RFC3339),
		Summary: map[string]any{
			"total":            len(items),
			"already_imported": importedCount,
			"new_candidates":   len(items) - importedCount,
		},
	}, nil
}

func (s *Service) CreateUpstream(ctx context.Context, actor ActorContext, input CreateUpstreamInput) (*catalog.Upstream, error) {
	var created *catalog.Upstream
	err := s.catalogRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		item, err := s.catalogRepo.CreateUpstreamWithDB(ctx, tx, catalog.CreateUpstreamInput(input))
		if err != nil {
			return err
		}
		created = item
		return s.auditRepo.CreateWithDB(ctx, tx, audit.CreateInput{
			ActorUserID:  optionalString(actor.ActorUserID),
			ActorRole:    optionalString(string(actor.ActorRole)),
			Action:       "admin.upstream.create",
			ResourceType: "upstream",
			ResourceID:   optionalString(item.ID),
			RequestID:    optionalString(actor.RequestID),
			IPAddress:    optionalString(actor.IPAddress),
			UserAgent:    optionalString(actor.UserAgent),
			Result:       audit.ResultSuccess,
			Details: map[string]any{
				"name":          item.Name,
				"provider_type": item.ProviderType,
				"status":        item.Status,
			},
		})
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *Service) UpdateUpstream(ctx context.Context, actor ActorContext, upstreamID string, input UpdateUpstreamInput) (*catalog.Upstream, error) {
	var updated *catalog.Upstream
	err := s.catalogRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		item, err := s.catalogRepo.UpdateUpstreamWithDB(ctx, tx, upstreamID, catalog.UpdateUpstreamInput(input))
		if err != nil {
			return err
		}
		updated = item
		return s.auditRepo.CreateWithDB(ctx, tx, audit.CreateInput{
			ActorUserID:  optionalString(actor.ActorUserID),
			ActorRole:    optionalString(string(actor.ActorRole)),
			Action:       "admin.upstream.update",
			ResourceType: "upstream",
			ResourceID:   optionalString(item.ID),
			RequestID:    optionalString(actor.RequestID),
			IPAddress:    optionalString(actor.IPAddress),
			UserAgent:    optionalString(actor.UserAgent),
			Result:       audit.ResultSuccess,
			Details: map[string]any{
				"name":          item.Name,
				"provider_type": item.ProviderType,
				"status":        item.Status,
			},
		})
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) ListChannels(ctx context.Context) ([]catalog.Channel, error) {
	return s.catalogRepo.ListChannels(ctx)
}

func (s *Service) GetChannel(ctx context.Context, channelID string) (*catalog.Channel, error) {
	return s.catalogRepo.GetChannelByID(ctx, channelID)
}

func (s *Service) CreateChannel(ctx context.Context, actor ActorContext, input CreateChannelInput) (*catalog.Channel, error) {
	var created *catalog.Channel
	err := s.catalogRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		item, err := s.catalogRepo.CreateChannelWithDB(ctx, tx, catalog.CreateChannelInput(input))
		if err != nil {
			return err
		}
		created = item
		return s.auditRepo.CreateWithDB(ctx, tx, audit.CreateInput{
			ActorUserID:  optionalString(actor.ActorUserID),
			ActorRole:    optionalString(string(actor.ActorRole)),
			Action:       "admin.channel.create",
			ResourceType: "channel",
			ResourceID:   optionalString(item.ID),
			RequestID:    optionalString(actor.RequestID),
			IPAddress:    optionalString(actor.IPAddress),
			UserAgent:    optionalString(actor.UserAgent),
			Result:       audit.ResultSuccess,
			Details: map[string]any{
				"name":   item.Name,
				"status": item.Status,
			},
		})
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *Service) UpdateChannel(ctx context.Context, actor ActorContext, channelID string, input UpdateChannelInput) (*catalog.Channel, error) {
	var updated *catalog.Channel
	err := s.catalogRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		item, err := s.catalogRepo.UpdateChannelWithDB(ctx, tx, channelID, catalog.UpdateChannelInput(input))
		if err != nil {
			return err
		}
		updated = item
		return s.auditRepo.CreateWithDB(ctx, tx, audit.CreateInput{
			ActorUserID:  optionalString(actor.ActorUserID),
			ActorRole:    optionalString(string(actor.ActorRole)),
			Action:       "admin.channel.update",
			ResourceType: "channel",
			ResourceID:   optionalString(item.ID),
			RequestID:    optionalString(actor.RequestID),
			IPAddress:    optionalString(actor.IPAddress),
			UserAgent:    optionalString(actor.UserAgent),
			Result:       audit.ResultSuccess,
			Details: map[string]any{
				"name":   item.Name,
				"status": item.Status,
			},
		})
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) ListModels(ctx context.Context) ([]AdminModelView, error) {
	items, err := s.catalogRepo.ListModels(ctx)
	if err != nil {
		return nil, err
	}
	return s.hydrateModelViews(ctx, items)
}

func (s *Service) GetModel(ctx context.Context, modelID string) (*AdminModelView, error) {
	item, err := s.catalogRepo.GetModelByID(ctx, modelID)
	if err != nil {
		return nil, err
	}

	items, err := s.hydrateModelViews(ctx, []catalog.ModelWithBindings{*item})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, catalog.ErrModelNotFound
	}
	return &items[0], nil
}

func (s *Service) CreateModel(ctx context.Context, actor ActorContext, input CreateModelInput) (*catalog.ModelWithBindings, error) {
	var created *catalog.ModelWithBindings
	err := s.catalogRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		item, err := s.catalogRepo.CreateModelWithDB(ctx, tx, catalog.CreateModelInput(input))
		if err != nil {
			return err
		}
		created = item
		return s.auditRepo.CreateWithDB(ctx, tx, audit.CreateInput{
			ActorUserID:  optionalString(actor.ActorUserID),
			ActorRole:    optionalString(string(actor.ActorRole)),
			Action:       "admin.model.create",
			ResourceType: "model",
			ResourceID:   optionalString(item.Model.ID),
			RequestID:    optionalString(actor.RequestID),
			IPAddress:    optionalString(actor.IPAddress),
			UserAgent:    optionalString(actor.UserAgent),
			Result:       audit.ResultSuccess,
			Details: map[string]any{
				"model_key":              item.Model.ModelKey,
				"visible_user_group_ids": item.Model.VisibleUserGroupIDs,
			},
		})
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *Service) UpdateModel(ctx context.Context, actor ActorContext, modelID string, input UpdateModelInput) (*catalog.ModelWithBindings, error) {
	var updated *catalog.ModelWithBindings
	err := s.catalogRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		item, err := s.catalogRepo.UpdateModelWithDB(ctx, tx, modelID, catalog.UpdateModelInput(input))
		if err != nil {
			return err
		}
		updated = item
		return s.auditRepo.CreateWithDB(ctx, tx, audit.CreateInput{
			ActorUserID:  optionalString(actor.ActorUserID),
			ActorRole:    optionalString(string(actor.ActorRole)),
			Action:       "admin.model.update",
			ResourceType: "model",
			ResourceID:   optionalString(item.Model.ID),
			RequestID:    optionalString(actor.RequestID),
			IPAddress:    optionalString(actor.IPAddress),
			UserAgent:    optionalString(actor.UserAgent),
			Result:       audit.ResultSuccess,
			Details: map[string]any{
				"model_key":              item.Model.ModelKey,
				"visible_user_group_ids": item.Model.VisibleUserGroupIDs,
			},
		})
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) ListUserGroups(ctx context.Context) ([]account.UserGroup, error) {
	return s.accountRepo.ListUserGroups(ctx)
}

func (s *Service) GetUserGroup(ctx context.Context, userGroupID string) (*account.UserGroup, error) {
	return s.accountRepo.GetUserGroupByID(ctx, userGroupID)
}

func (s *Service) ImportModels(ctx context.Context, actor ActorContext, input ImportModelsInput) (*ImportModelsResult, error) {
	upstreamID := strings.TrimSpace(input.UpstreamID)
	if upstreamID == "" || len(input.Items) == 0 {
		return nil, ErrModelImportInvalid
	}

	upstream, err := s.catalogRepo.GetUpstreamByID(ctx, upstreamID)
	if err != nil {
		return nil, err
	}

	if err := s.validateImportInputs(ctx, input.Items); err != nil {
		return nil, err
	}

	existingModels, err := s.catalogRepo.ListModels(ctx)
	if err != nil {
		return nil, err
	}

	existingByKey := make(map[string]catalog.ModelWithBindings, len(existingModels))
	for _, item := range existingModels {
		existingByKey[strings.ToLower(strings.TrimSpace(item.Model.ModelKey))] = item
	}

	records := make([]ImportedModelRecord, 0, len(input.Items))
	createdModelIDs := make([]string, 0, len(input.Items))

	err = s.catalogRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, requestItem := range input.Items {
			modelKey := strings.TrimSpace(requestItem.ModelKey)
			if modelKey == "" {
				return ErrModelImportInvalid
			}

			existing, ok := existingByKey[strings.ToLower(modelKey)]
			if ok {
				records = append(records, ImportedModelRecord{
					RequestedModelKey: modelKey,
					ExistingModel: &ImportedModelSummary{
						ID:          existing.Model.ID,
						ModelKey:    existing.Model.ModelKey,
						DisplayName: existing.Model.DisplayName,
						Status:      existing.Model.Status,
					},
					Status: "skipped_existing",
				})
				continue
			}

			channelID := sanitizeOptionalID(requestItem.ChannelID)
			routeBindings := []catalog.RouteBindingInput{{
				ChannelID:  channelID,
				UpstreamID: upstream.ID,
				Priority:   defaultPositiveInt(requestItem.Priority, 1),
				Status:     string(catalog.RouteBindingStatusActive),
			}}

			modelInput := catalog.CreateModelInput{
				ModelKey:            modelKey,
				DisplayName:         defaultString(strings.TrimSpace(requestItem.DisplayName), modelKey),
				ProviderType:        defaultString(strings.TrimSpace(requestItem.ProviderType), upstream.ProviderType),
				ContextLength:       requestItem.ContextLength,
				MaxOutputTokens:     requestItem.MaxOutputTokens,
				Pricing:             nonNilMap(requestItem.Pricing),
				Capabilities:        withDefaultCapabilities(requestItem.Capabilities),
				VisibleUserGroupIDs: sanitizeStringSlice(requestItem.VisibleUserGroupIDs),
				Status:              defaultString(strings.TrimSpace(requestItem.Status), string(catalog.ModelStatusActive)),
				Metadata:            nonNilMap(requestItem.Metadata),
				RouteBindings:       routeBindings,
			}

			created, createErr := s.catalogRepo.CreateModelWithDB(ctx, tx, modelInput)
			if createErr != nil {
				return createErr
			}

			existingByKey[strings.ToLower(modelKey)] = *created
			createdModelIDs = append(createdModelIDs, created.Model.ID)
			records = append(records, ImportedModelRecord{
				RequestedModelKey: modelKey,
				ExistingModel: &ImportedModelSummary{
					ID:          created.Model.ID,
					ModelKey:    created.Model.ModelKey,
					DisplayName: created.Model.DisplayName,
					Status:      created.Model.Status,
				},
				Status: "imported",
			})
		}

		return s.auditRepo.CreateWithDB(ctx, tx, audit.CreateInput{
			ActorUserID:  optionalString(actor.ActorUserID),
			ActorRole:    optionalString(string(actor.ActorRole)),
			Action:       "admin.model.import",
			ResourceType: "upstream",
			ResourceID:   optionalString(upstream.ID),
			RequestID:    optionalString(actor.RequestID),
			IPAddress:    optionalString(actor.IPAddress),
			UserAgent:    optionalString(actor.UserAgent),
			Result:       audit.ResultSuccess,
			Details: map[string]any{
				"upstream_name":        upstream.Name,
				"requested":            len(input.Items),
				"imported":             len(createdModelIDs),
				"skipped_existing":     len(input.Items) - len(createdModelIDs),
				"created_model_ids":    createdModelIDs,
				"requested_model_keys": collectRequestedModelKeys(input.Items),
			},
		})
	})
	if err != nil {
		return nil, err
	}

	createdViews := make(map[string]*AdminModelView, len(createdModelIDs))
	if len(createdModelIDs) > 0 {
		models, loadErr := s.loadModelsByIDs(ctx, createdModelIDs)
		if loadErr != nil {
			return nil, loadErr
		}
		hydrated, hydrateErr := s.hydrateModelViews(ctx, models)
		if hydrateErr != nil {
			return nil, hydrateErr
		}
		for index := range hydrated {
			view := hydrated[index]
			createdViews[view.Item.Model.ModelKey] = &view
		}
	}

	importedCount := 0
	skippedCount := 0
	for index := range records {
		if records[index].Status == "imported" {
			importedCount++
			if records[index].ExistingModel != nil {
				records[index].Model = createdViews[records[index].ExistingModel.ModelKey]
			}
			continue
		}
		skippedCount++
	}

	return &ImportModelsResult{
		Upstream: upstream,
		Items:    records,
		Summary: map[string]any{
			"requested":        len(input.Items),
			"imported":         importedCount,
			"skipped_existing": skippedCount,
		},
	}, nil
}

func (s *Service) CreateUserGroup(ctx context.Context, actor ActorContext, input CreateUserGroupInput) (*account.UserGroup, error) {
	var created *account.UserGroup
	err := s.accountRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		item, err := s.accountRepo.CreateUserGroupWithDB(ctx, tx, account.CreateUserGroupInput(input))
		if err != nil {
			return err
		}
		created = item
		return s.auditRepo.CreateWithDB(ctx, tx, audit.CreateInput{
			ActorUserID:  optionalString(actor.ActorUserID),
			ActorRole:    optionalString(string(actor.ActorRole)),
			Action:       "admin.user_group.create",
			ResourceType: "user_group",
			ResourceID:   optionalString(item.ID),
			RequestID:    optionalString(actor.RequestID),
			IPAddress:    optionalString(actor.IPAddress),
			UserAgent:    optionalString(actor.UserAgent),
			Result:       audit.ResultSuccess,
			Details: map[string]any{
				"name":   item.Name,
				"status": item.Status,
			},
		})
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *Service) UpdateUserGroup(ctx context.Context, actor ActorContext, userGroupID string, input UpdateUserGroupInput) (*account.UserGroup, error) {
	var updated *account.UserGroup
	err := s.accountRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		item, err := s.accountRepo.UpdateUserGroupWithDB(ctx, tx, userGroupID, account.UpdateUserGroupInput(input))
		if err != nil {
			return err
		}
		updated = item
		return s.auditRepo.CreateWithDB(ctx, tx, audit.CreateInput{
			ActorUserID:  optionalString(actor.ActorUserID),
			ActorRole:    optionalString(string(actor.ActorRole)),
			Action:       "admin.user_group.update",
			ResourceType: "user_group",
			ResourceID:   optionalString(item.ID),
			RequestID:    optionalString(actor.RequestID),
			IPAddress:    optionalString(actor.IPAddress),
			UserAgent:    optionalString(actor.UserAgent),
			Result:       audit.ResultSuccess,
			Details: map[string]any{
				"name":   item.Name,
				"status": item.Status,
			},
		})
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) AssignUserGroup(ctx context.Context, actor ActorContext, userID string, userGroupID *string) (*account.User, error) {
	if userGroupID != nil && strings.TrimSpace(*userGroupID) != "" {
		if _, err := s.accountRepo.GetUserGroupByID(ctx, strings.TrimSpace(*userGroupID)); err != nil {
			return nil, err
		}
	}

	var updated *account.User
	err := s.accountRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user, err := s.accountRepo.AssignUserGroupWithDB(ctx, tx, userID, userGroupID)
		if err != nil {
			return err
		}
		updated = user
		return s.auditRepo.CreateWithDB(ctx, tx, audit.CreateInput{
			ActorUserID:  optionalString(actor.ActorUserID),
			ActorRole:    optionalString(string(actor.ActorRole)),
			Action:       "admin.user.assign_group",
			ResourceType: "user",
			ResourceID:   optionalString(user.ID),
			TargetUserID: optionalString(user.ID),
			RequestID:    optionalString(actor.RequestID),
			IPAddress:    optionalString(actor.IPAddress),
			UserAgent:    optionalString(actor.UserAgent),
			Result:       audit.ResultSuccess,
			Details: map[string]any{
				"user_group_id": optionalStringValue(user.UserGroupID),
			},
		})
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) ListUsers(ctx context.Context, filter ListUsersFilter) (UserListResult, error) {
	query := s.accountRepo.DB().WithContext(ctx).Model(&account.User{})
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + strings.ToLower(keyword) + "%"
		query = query.Where("LOWER(username) LIKE ? OR LOWER(email) LIKE ? OR LOWER(display_name) LIKE ?", like, like, like)
	}
	if status := strings.TrimSpace(filter.Status); status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return UserListResult{}, fmt.Errorf("count admin users: %w", err)
	}

	var users []account.User
	if err := query.Order("created_at DESC").
		Limit(filter.PageSize).
		Offset((filter.Page - 1) * filter.PageSize).
		Find(&users).Error; err != nil {
		return UserListResult{}, fmt.Errorf("list admin users: %w", err)
	}

	groupMap, err := s.loadUserGroupsByIDs(ctx, collectUserGroupIDs(users))
	if err != nil {
		return UserListResult{}, err
	}

	items := make([]AdminUserRecord, 0, len(users))
	for _, user := range users {
		var userGroup *account.UserGroup
		if user.UserGroupID != nil {
			userGroup = groupMap[*user.UserGroupID]
		}
		items = append(items, AdminUserRecord{
			User:      user,
			UserGroup: userGroup,
		})
	}

	return UserListResult{
		Items: items,
		Total: total,
	}, nil
}

func (s *Service) AdjustUserQuota(ctx context.Context, actor ActorContext, targetUserID string, delta int64, reason string) (*account.User, error) {
	var updatedUser *account.User
	err := s.accountRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user account.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&user, "id = ?", targetUserID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return account.ErrUserNotFound
			}
			return fmt.Errorf("load user for quota adjustment: %w", err)
		}

		nextQuota := user.Quota + delta
		if nextQuota < 0 {
			return ErrQuotaWouldBecomeNegative
		}

		now := time.Now().UTC()
		user.Quota = nextQuota
		user.UpdatedAt = now
		if err := tx.Save(&user).Error; err != nil {
			return fmt.Errorf("update user quota: %w", err)
		}

		requestID := optionalString(actor.RequestID)
		operatorUserID := optionalString(actor.ActorUserID)
		quotaReason := optionalString(strings.TrimSpace(reason))
		if err := tx.Create(&account.QuotaLog{
			ID:           uuid.NewString(),
			UserID:       user.ID,
			RequestID:    requestID,
			LogType:      account.QuotaLogTypeAdminAdjust,
			DeltaQuota:   delta,
			BalanceAfter: nextQuota,
			Reason:       quotaReason,
			CreatedAt:    now,
		}).Error; err != nil {
			return fmt.Errorf("create quota log: %w", err)
		}

		if err := s.auditRepo.CreateWithDB(ctx, tx, audit.CreateInput{
			ActorUserID:  operatorUserID,
			ActorRole:    optionalString(string(actor.ActorRole)),
			Action:       "admin.user.quota_adjust",
			ResourceType: "user",
			ResourceID:   optionalString(user.ID),
			TargetUserID: optionalString(user.ID),
			RequestID:    requestID,
			IPAddress:    optionalString(actor.IPAddress),
			UserAgent:    optionalString(actor.UserAgent),
			Result:       audit.ResultSuccess,
			Details: map[string]any{
				"delta":         delta,
				"balance_after": nextQuota,
				"reason":        strings.TrimSpace(reason),
				"operator_user": actor.ActorUserID,
			},
		}); err != nil {
			return err
		}

		updatedUser = &user
		return nil
	})
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (s *Service) ListUserGroupLimitPolicies(ctx context.Context, userGroupID string) ([]limits.UserGroupModelLimitPolicy, error) {
	if _, err := s.accountRepo.GetUserGroupByID(ctx, userGroupID); err != nil {
		return nil, err
	}
	return s.limitsService.ListPoliciesByUserGroup(ctx, userGroupID)
}

func (s *Service) ReplaceUserGroupLimitPolicies(ctx context.Context, actor ActorContext, userGroupID string, inputs []PolicyUpsertInput) ([]limits.UserGroupModelLimitPolicy, error) {
	if _, err := s.accountRepo.GetUserGroupByID(ctx, userGroupID); err != nil {
		return nil, err
	}

	var items []limits.UserGroupModelLimitPolicy
	err := s.accountRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		policies, err := s.limitsService.ReplacePoliciesByUserGroupWithDB(ctx, tx, userGroupID, inputs)
		if err != nil {
			return err
		}
		items = policies
		return s.auditRepo.CreateWithDB(ctx, tx, audit.CreateInput{
			ActorUserID:  optionalString(actor.ActorUserID),
			ActorRole:    optionalString(string(actor.ActorRole)),
			Action:       "admin.user_group_limits.update",
			ResourceType: "user_group",
			ResourceID:   optionalString(userGroupID),
			RequestID:    optionalString(actor.RequestID),
			IPAddress:    optionalString(actor.IPAddress),
			UserAgent:    optionalString(actor.UserAgent),
			Result:       audit.ResultSuccess,
			Details: map[string]any{
				"policy_count": len(items),
			},
		})
	})
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (s *Service) GetUserLimitUsage(ctx context.Context, userID string, modelID *string) (limits.UsageReport, error) {
	return s.limitsService.GetUserLimitUsage(ctx, userID, modelID, time.Now().UTC())
}

func (s *Service) ListUserLimitAdjustments(ctx context.Context, userID string, modelID *string, page, pageSize int) (limits.AdjustmentListResult, error) {
	return s.limitsService.ListAdjustments(ctx, limits.AdjustmentListFilter{
		UserID:   userID,
		ModelID:  modelID,
		Page:     page,
		PageSize: pageSize,
	})
}

func (s *Service) CreateUserLimitAdjustment(ctx context.Context, actor ActorContext, userID string, input CreateUserLimitAdjustmentInput) (*limits.UserLimitAdjustment, error) {
	var created *limits.UserLimitAdjustment
	err := s.accountRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		item, err := s.limitsService.CreateAdjustmentWithDB(ctx, tx, limits.AdjustmentCreateInput{
			UserID:      userID,
			ModelID:     input.ModelID,
			MetricType:  input.MetricType,
			WindowType:  input.WindowType,
			Delta:       input.Delta,
			Reason:      input.Reason,
			ActorUserID: optionalString(actor.ActorUserID),
		})
		if err != nil {
			return err
		}
		created = item
		return s.auditRepo.CreateWithDB(ctx, tx, audit.CreateInput{
			ActorUserID:  optionalString(actor.ActorUserID),
			ActorRole:    optionalString(string(actor.ActorRole)),
			Action:       "admin.user_limit_adjustment.create",
			ResourceType: "user",
			ResourceID:   optionalString(userID),
			TargetUserID: optionalString(userID),
			RequestID:    optionalString(actor.RequestID),
			IPAddress:    optionalString(actor.IPAddress),
			UserAgent:    optionalString(actor.UserAgent),
			Result:       audit.ResultSuccess,
			Details: map[string]any{
				"model_id":    optionalStringValue(input.ModelID),
				"metric_type": input.MetricType,
				"window_type": input.WindowType,
				"delta":       input.Delta,
				"expires_at":  formatOptionalAdjustmentTime(created.ExpiresAt),
				"reason":      optionalStringValue(input.Reason),
			},
		})
	})
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) ListAuditLogs(ctx context.Context, filter audit.ListFilter) (audit.ListResult, error) {
	return s.auditRepo.List(ctx, filter)
}

func (s *Service) validateImportInputs(ctx context.Context, items []ImportModelItemInput) error {
	if len(items) == 0 {
		return ErrModelImportInvalid
	}

	seenModelKeys := make(map[string]struct{}, len(items))
	userGroupIDs := make([]string, 0)
	channelIDs := make([]string, 0)

	for _, item := range items {
		modelKey := strings.ToLower(strings.TrimSpace(item.ModelKey))
		if modelKey == "" {
			return ErrModelImportInvalid
		}
		if _, exists := seenModelKeys[modelKey]; exists {
			return ErrModelImportInvalid
		}
		seenModelKeys[modelKey] = struct{}{}

		userGroupIDs = append(userGroupIDs, sanitizeStringSlice(item.VisibleUserGroupIDs)...)
		if channelID := sanitizeOptionalID(item.ChannelID); channelID != nil {
			channelIDs = append(channelIDs, *channelID)
		}
	}

	groupMap, err := s.loadUserGroupsByIDs(ctx, userGroupIDs)
	if err != nil {
		return err
	}
	for _, userGroupID := range sanitizeStringSlice(userGroupIDs) {
		if _, ok := groupMap[userGroupID]; !ok {
			return account.ErrUserGroupNotFound
		}
	}

	channelMap, err := s.loadChannelsByIDs(ctx, channelIDs)
	if err != nil {
		return err
	}
	for _, channelID := range sanitizeStringSlice(channelIDs) {
		if _, ok := channelMap[channelID]; !ok {
			return catalog.ErrChannelNotFound
		}
	}

	return nil
}

func (s *Service) hydrateModelViews(ctx context.Context, items []catalog.ModelWithBindings) ([]AdminModelView, error) {
	if len(items) == 0 {
		return []AdminModelView{}, nil
	}

	userGroupIDs := make([]string, 0)
	channelIDs := make([]string, 0)
	upstreamIDs := make([]string, 0)

	for _, item := range items {
		userGroupIDs = append(userGroupIDs, item.Model.VisibleUserGroupIDs...)
		for _, binding := range item.RouteBindings {
			if binding.ChannelID != nil {
				channelIDs = append(channelIDs, *binding.ChannelID)
			}
			upstreamIDs = append(upstreamIDs, binding.UpstreamID)
		}
	}

	groupMap, err := s.loadUserGroupsByIDs(ctx, userGroupIDs)
	if err != nil {
		return nil, err
	}
	channelMap, err := s.loadChannelsByIDs(ctx, channelIDs)
	if err != nil {
		return nil, err
	}
	upstreamMap, err := s.loadUpstreamsByIDs(ctx, upstreamIDs)
	if err != nil {
		return nil, err
	}

	result := make([]AdminModelView, 0, len(items))
	for _, item := range items {
		visibleGroups := make([]account.UserGroup, 0, len(item.Model.VisibleUserGroupIDs))
		visibilityNames := make([]string, 0, len(item.Model.VisibleUserGroupIDs))
		for _, groupID := range item.Model.VisibleUserGroupIDs {
			group, ok := groupMap[groupID]
			if !ok {
				continue
			}
			visibleGroups = append(visibleGroups, *group)
			visibilityNames = append(visibilityNames, group.Name)
		}

		hydratedBindings := make([]AdminModelRouteBindingView, 0, len(item.RouteBindings))
		routeSummaries := make([]string, 0, len(item.RouteBindings))
		for _, binding := range item.RouteBindings {
			var channel *catalog.Channel
			if binding.ChannelID != nil {
				channel = channelMap[*binding.ChannelID]
			}
			upstream := upstreamMap[binding.UpstreamID]
			hydratedBindings = append(hydratedBindings, AdminModelRouteBindingView{
				Binding:  binding,
				Channel:  channel,
				Upstream: upstream,
			})

			routeSummaries = append(routeSummaries, summarizeRouteBinding(binding, channel, upstream))
		}

		visibilitySummary := "all users"
		if len(visibilityNames) > 0 {
			visibilitySummary = strings.Join(visibilityNames, ", ")
		}

		result = append(result, AdminModelView{
			Item:               item,
			VisibleUserGroups:  visibleGroups,
			HydratedBindings:   hydratedBindings,
			VisibilitySummary:  visibilitySummary,
			RouteRuleSummaries: routeSummaries,
		})
	}

	return result, nil
}

func (s *Service) loadUserGroupsByIDs(ctx context.Context, ids []string) (map[string]*account.UserGroup, error) {
	if len(ids) == 0 {
		return map[string]*account.UserGroup{}, nil
	}

	var groups []account.UserGroup
	if err := s.accountRepo.DB().WithContext(ctx).Where("id IN ?", ids).Find(&groups).Error; err != nil {
		return nil, fmt.Errorf("load user groups by ids: %w", err)
	}

	result := make(map[string]*account.UserGroup, len(groups))
	for i := range groups {
		group := groups[i]
		result[group.ID] = &group
	}

	return result, nil
}

func (s *Service) loadChannelsByIDs(ctx context.Context, ids []string) (map[string]*catalog.Channel, error) {
	normalized := sanitizeStringSlice(ids)
	if len(normalized) == 0 {
		return map[string]*catalog.Channel{}, nil
	}

	var items []catalog.Channel
	if err := s.catalogRepo.DB().WithContext(ctx).Where("id IN ?", normalized).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("load channels by ids: %w", err)
	}

	result := make(map[string]*catalog.Channel, len(items))
	for i := range items {
		item := items[i]
		result[item.ID] = &item
	}

	return result, nil
}

func (s *Service) loadUpstreamsByIDs(ctx context.Context, ids []string) (map[string]*catalog.Upstream, error) {
	normalized := sanitizeStringSlice(ids)
	if len(normalized) == 0 {
		return map[string]*catalog.Upstream{}, nil
	}

	var items []catalog.Upstream
	if err := s.catalogRepo.DB().WithContext(ctx).Where("id IN ?", normalized).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("load upstreams by ids: %w", err)
	}

	result := make(map[string]*catalog.Upstream, len(items))
	for i := range items {
		item := items[i]
		result[item.ID] = &item
	}

	return result, nil
}

func (s *Service) loadModelsByIDs(ctx context.Context, ids []string) ([]catalog.ModelWithBindings, error) {
	normalized := sanitizeStringSlice(ids)
	if len(normalized) == 0 {
		return []catalog.ModelWithBindings{}, nil
	}

	var models []catalog.Model
	if err := s.catalogRepo.DB().WithContext(ctx).
		Where("id IN ?", normalized).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("load models by ids: %w", err)
	}

	return s.catalogRepo.ListModelsByRows(ctx, models)
}

func collectUserGroupIDs(users []account.User) []string {
	result := make([]string, 0)
	seen := make(map[string]struct{})
	for _, user := range users {
		if user.UserGroupID == nil || *user.UserGroupID == "" {
			continue
		}
		if _, ok := seen[*user.UserGroupID]; ok {
			continue
		}
		seen[*user.UserGroupID] = struct{}{}
		result = append(result, *user.UserGroupID)
	}
	return result
}

func formatOptionalAdjustmentTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339)
}

func optionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func optionalStringValue(value *string) any {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	return strings.TrimSpace(*value)
}

func sanitizeOptionalID(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func defaultPositiveInt(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func withDefaultCapabilities(value map[string]any) map[string]any {
	result := nonNilMap(value)
	if len(result) == 0 {
		return map[string]any{
			"chat": true,
		}
	}
	return result
}

func collectRequestedModelKeys(items []ImportModelItemInput) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		modelKey := strings.TrimSpace(item.ModelKey)
		if modelKey == "" {
			continue
		}
		result = append(result, modelKey)
	}
	return result
}

func summarizeRouteBinding(binding catalog.ModelRouteBinding, channel *catalog.Channel, upstream *catalog.Upstream) string {
	channelName := "default route"
	if channel != nil {
		channelName = channel.Name
	}

	upstreamName := binding.UpstreamID
	if upstream != nil {
		upstreamName = upstream.Name
	}

	return fmt.Sprintf("%s -> %s (priority %d)", channelName, upstreamName, binding.Priority)
}
