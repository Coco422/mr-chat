package limits

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"mrchat/internal/modules/account"
)

var ErrLimitExceeded = errors.New("user model limit exceeded")

type Service struct {
	repo        *Repository
	accountRepo *account.Repository
}

type UsageReport struct {
	UserID          string          `json:"user_id"`
	UserGroupID     *string         `json:"user_group_id"`
	ModelID         *string         `json:"model_id"`
	EffectivePolicy EffectivePolicy `json:"effective_policy"`
	Usage           UsageCounters   `json:"usage"`
	Adjustments     UsageCounters   `json:"adjustments"`
	Remaining       UsageCounters   `json:"remaining"`
}

type LimitCheckInput struct {
	UserID                   string
	ModelID                  *string
	PromptTokens             int64
	ReservedCompletionTokens int64
	Now                      time.Time
}

type LimitCheckResult struct {
	Report UsageReport
}

func NewService(repo *Repository, accountRepo *account.Repository) *Service {
	return &Service{
		repo:        repo,
		accountRepo: accountRepo,
	}
}

func (s *Service) ListPoliciesByUserGroup(ctx context.Context, userGroupID string) ([]UserGroupModelLimitPolicy, error) {
	return s.repo.ListPoliciesByUserGroup(ctx, userGroupID)
}

func (s *Service) ReplacePoliciesByUserGroup(ctx context.Context, userGroupID string, inputs []PolicyUpsertInput) ([]UserGroupModelLimitPolicy, error) {
	return s.repo.ReplacePoliciesByUserGroup(ctx, userGroupID, inputs)
}

func (s *Service) ReplacePoliciesByUserGroupWithDB(ctx context.Context, db *gorm.DB, userGroupID string, inputs []PolicyUpsertInput) ([]UserGroupModelLimitPolicy, error) {
	return s.repo.ReplacePoliciesByUserGroupWithDB(ctx, db, userGroupID, inputs)
}

func (s *Service) CreateAdjustment(ctx context.Context, input AdjustmentCreateInput) (*UserLimitAdjustment, error) {
	if input.ExpiresAt == nil {
		now := time.Now().UTC()
		input.ExpiresAt = buildExpiresAt(input.WindowType, now)
	}

	return s.repo.CreateAdjustment(ctx, input)
}

func (s *Service) CreateAdjustmentWithDB(ctx context.Context, db *gorm.DB, input AdjustmentCreateInput) (*UserLimitAdjustment, error) {
	if input.ExpiresAt == nil {
		now := time.Now().UTC()
		input.ExpiresAt = buildExpiresAt(input.WindowType, now)
	}

	return s.repo.CreateAdjustmentWithDB(ctx, db, input)
}

func (s *Service) ListAdjustments(ctx context.Context, filter AdjustmentListFilter) (AdjustmentListResult, error) {
	return s.repo.ListAdjustments(ctx, filter)
}

func (s *Service) CreateRequestLog(ctx context.Context, input RequestLogCreateInput) (*LLMRequestLog, error) {
	return s.repo.CreateRequestLog(ctx, input)
}

func (s *Service) CreateRequestLogWithDB(ctx context.Context, db *gorm.DB, input RequestLogCreateInput) (*LLMRequestLog, error) {
	return s.repo.CreateRequestLogWithDB(ctx, db, input)
}

func (s *Service) UpdateRequestLogByRequestID(ctx context.Context, requestID string, input RequestLogUpdateInput) (*LLMRequestLog, error) {
	return s.repo.UpdateRequestLogByRequestID(ctx, requestID, input)
}

func (s *Service) UpdateRequestLogByRequestIDWithDB(ctx context.Context, db *gorm.DB, requestID string, input RequestLogUpdateInput) (*LLMRequestLog, error) {
	return s.repo.UpdateRequestLogByRequestIDWithDB(ctx, db, requestID, input)
}

func (s *Service) GetUserLimitUsage(ctx context.Context, userID string, modelID *string, now time.Time) (UsageReport, error) {
	user, err := s.accountRepo.GetUserByID(ctx, userID)
	if err != nil {
		return UsageReport{}, err
	}

	report := UsageReport{
		UserID:      user.ID,
		UserGroupID: user.UserGroupID,
		ModelID:     modelID,
		EffectivePolicy: EffectivePolicy{
			Source:  "none",
			ModelID: nil,
		},
	}

	if user.UserGroupID != nil && *user.UserGroupID != "" {
		if policy, err := s.repo.ResolvePolicy(ctx, *user.UserGroupID, modelID); err == nil {
			report.EffectivePolicy = toEffectivePolicy(policy, modelID)
		} else if !errors.Is(err, ErrPolicyNotFound) {
			return UsageReport{}, err
		}
	}

	usage, err := s.repo.GetRequestUsage(ctx, RequestUsageFilter{
		UserID:  userID,
		ModelID: modelID,
		Now:     now.UTC(),
	})
	if err != nil {
		return UsageReport{}, err
	}

	adjustments, err := s.repo.GetAdjustmentUsage(ctx, userID, modelID, now.UTC())
	if err != nil {
		return UsageReport{}, err
	}

	report.Usage = usage
	report.Adjustments = adjustments
	report.Remaining = UsageCounters{
		Hour: UsageCounter{
			Requests: computeRemaining(report.EffectivePolicy.HourRequestLimit, adjustments.Hour.Requests, usage.Hour.Requests),
			Tokens:   computeRemaining(report.EffectivePolicy.HourTokenLimit, adjustments.Hour.Tokens, usage.Hour.Tokens),
		},
		Week: UsageCounter{
			Requests: computeRemaining(report.EffectivePolicy.WeekRequestLimit, adjustments.Week.Requests, usage.Week.Requests),
			Tokens:   computeRemaining(report.EffectivePolicy.WeekTokenLimit, adjustments.Week.Tokens, usage.Week.Tokens),
		},
		Lifetime: UsageCounter{
			Requests: computeRemaining(report.EffectivePolicy.LifetimeRequestLimit, adjustments.Lifetime.Requests, usage.Lifetime.Requests),
			Tokens:   computeRemaining(report.EffectivePolicy.LifetimeTokenLimit, adjustments.Lifetime.Tokens, usage.Lifetime.Tokens),
		},
	}

	return report, nil
}

func (s *Service) CheckUserModelLimit(ctx context.Context, input LimitCheckInput) (LimitCheckResult, error) {
	now := input.Now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}

	report, err := s.GetUserLimitUsage(ctx, input.UserID, input.ModelID, now)
	if err != nil {
		return LimitCheckResult{}, err
	}

	if exceedsRequestLimit(report, 1) {
		return LimitCheckResult{Report: report}, ErrLimitExceeded
	}

	reservedTokens := input.PromptTokens + input.ReservedCompletionTokens
	if exceedsTokenLimit(report, reservedTokens) {
		return LimitCheckResult{Report: report}, ErrLimitExceeded
	}

	return LimitCheckResult{Report: report}, nil
}

func RecordRejectedRequest(ctx context.Context, repo *Repository, requestID string, report UsageReport, errorCode string, metadata map[string]any) error {
	if repo == nil {
		return nil
	}

	_, err := repo.CreateRequestLog(ctx, RequestLogCreateInput{
		RequestID:   requestID,
		UserID:      report.UserID,
		UserGroupID: report.UserGroupID,
		ModelID:     report.ModelID,
		Status:      RequestLogStatusRejected,
		ErrorCode:   &errorCode,
		StartedAt:   time.Now().UTC(),
		CompletedAt: timePtr(time.Now().UTC()),
		Metadata:    metadata,
	})
	return err
}

func (s *Service) RecordRejectedRequest(ctx context.Context, requestID string, report UsageReport, errorCode string, metadata map[string]any) error {
	return RecordRejectedRequest(ctx, s.repo, requestID, report, errorCode, metadata)
}

func toEffectivePolicy(policy *UserGroupModelLimitPolicy, modelID *string) EffectivePolicy {
	if policy == nil {
		return EffectivePolicy{
			Source:  "none",
			ModelID: modelID,
		}
	}

	source := "group_default"
	if policy.ModelID != nil {
		source = "model_override"
	}

	return EffectivePolicy{
		Source:               source,
		ModelID:              policy.ModelID,
		HourRequestLimit:     policy.HourRequestLimit,
		WeekRequestLimit:     policy.WeekRequestLimit,
		LifetimeRequestLimit: policy.LifetimeRequestLimit,
		HourTokenLimit:       policy.HourTokenLimit,
		WeekTokenLimit:       policy.WeekTokenLimit,
		LifetimeTokenLimit:   policy.LifetimeTokenLimit,
	}
}

func buildExpiresAt(windowType WindowType, now time.Time) *time.Time {
	switch windowType {
	case WindowTypeRollingHour:
		return timePtr(now.Add(1 * time.Hour))
	case WindowTypeRollingWeek:
		return timePtr(now.Add(7 * 24 * time.Hour))
	default:
		return nil
	}
}

func computeRemaining(limit *int64, adjustment, used int64) int64 {
	if limit == nil {
		return -1
	}
	return (*limit + adjustment) - used
}

func exceedsRequestLimit(report UsageReport, increment int64) bool {
	return exceedsLimit(report.EffectivePolicy.HourRequestLimit, report.Adjustments.Hour.Requests, report.Usage.Hour.Requests, increment) ||
		exceedsLimit(report.EffectivePolicy.WeekRequestLimit, report.Adjustments.Week.Requests, report.Usage.Week.Requests, increment) ||
		exceedsLimit(report.EffectivePolicy.LifetimeRequestLimit, report.Adjustments.Lifetime.Requests, report.Usage.Lifetime.Requests, increment)
}

func exceedsTokenLimit(report UsageReport, reserved int64) bool {
	return exceedsLimit(report.EffectivePolicy.HourTokenLimit, report.Adjustments.Hour.Tokens, report.Usage.Hour.Tokens, reserved) ||
		exceedsLimit(report.EffectivePolicy.WeekTokenLimit, report.Adjustments.Week.Tokens, report.Usage.Week.Tokens, reserved) ||
		exceedsLimit(report.EffectivePolicy.LifetimeTokenLimit, report.Adjustments.Lifetime.Tokens, report.Usage.Lifetime.Tokens, reserved)
}

func exceedsLimit(limit *int64, adjustment, used, increment int64) bool {
	if limit == nil {
		return false
	}

	return used+increment > (*limit + adjustment)
}

func timePtr(value time.Time) *time.Time {
	return &value
}

func (r UsageReport) String() string {
	return fmt.Sprintf("user=%s model=%v source=%s", r.UserID, r.ModelID, r.EffectivePolicy.Source)
}
