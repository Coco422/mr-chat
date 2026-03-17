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
)

var (
	ErrQuotaWouldBecomeNegative = errors.New("quota would become negative")
)

type Service struct {
	accountRepo *account.Repository
	catalogRepo *catalog.Repository
	auditRepo   *audit.Repository
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

type UserListResult struct {
	Items []account.User
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

func NewService(accountRepo *account.Repository, catalogRepo *catalog.Repository, auditRepo *audit.Repository) *Service {
	return &Service{
		accountRepo: accountRepo,
		catalogRepo: catalogRepo,
		auditRepo:   auditRepo,
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
				"model_key":         item.Model.ModelKey,
				"allowed_group_ids": item.Model.AllowedGroupIDs,
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
				"model_key":         item.Model.ModelKey,
				"allowed_group_ids": item.Model.AllowedGroupIDs,
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

	var items []account.User
	if err := query.Order("created_at DESC").
		Limit(filter.PageSize).
		Offset((filter.Page - 1) * filter.PageSize).
		Find(&items).Error; err != nil {
		return UserListResult{}, fmt.Errorf("list admin users: %w", err)
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

func (s *Service) ListAuditLogs(ctx context.Context, filter audit.ListFilter) (audit.ListResult, error) {
	return s.auditRepo.List(ctx, filter)
}

func optionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
