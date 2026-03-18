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
	ErrQuotaWouldBecomeNegative = errors.New("quota would become negative")
)

type Service struct {
	accountRepo   *account.Repository
	catalogRepo   *catalog.Repository
	limitsService *limits.Service
	auditRepo     *audit.Repository
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

func NewService(accountRepo *account.Repository, catalogRepo *catalog.Repository, limitsService *limits.Service, auditRepo *audit.Repository) *Service {
	return &Service{
		accountRepo:   accountRepo,
		catalogRepo:   catalogRepo,
		limitsService: limitsService,
		auditRepo:     auditRepo,
	}
}

func (s *Service) ListUpstreams(ctx context.Context) ([]catalog.Upstream, error) {
	return s.catalogRepo.ListUpstreams(ctx)
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

func (s *Service) ListModels(ctx context.Context) ([]catalog.ModelWithBindings, error) {
	return s.catalogRepo.ListModels(ctx)
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
