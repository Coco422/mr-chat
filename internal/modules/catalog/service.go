package catalog

import (
	"context"
	"slices"

	"mrchat/internal/modules/account"
)

type Service struct {
	repo        *Repository
	accountRepo *account.Repository
}

type UserModel struct {
	ID              string         `json:"id"`
	ModelKey        string         `json:"model_key"`
	DisplayName     string         `json:"display_name"`
	ProviderType    string         `json:"provider_type"`
	ContextLength   int            `json:"context_length"`
	MaxOutputTokens *int           `json:"max_output_tokens,omitempty"`
	Pricing         map[string]any `json:"pricing"`
	Capabilities    map[string]any `json:"capabilities"`
	Status          ModelStatus    `json:"status"`
}

func NewService(repo *Repository, accountRepo *account.Repository) *Service {
	return &Service{
		repo:        repo,
		accountRepo: accountRepo,
	}
}

func (s *Service) ListVisibleModels(ctx context.Context, userID string) ([]UserModel, error) {
	user, err := s.accountRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.ListActiveModels(ctx)
	if err != nil {
		return nil, err
	}

	if user.Role == account.RoleAdmin || user.Role == account.RoleRoot {
		return toUserModels(items), nil
	}

	allowedGroupIDs := loadUserGroupIDs(ctx, s.accountRepo, user)
	result := make([]UserModel, 0, len(items))
	for _, item := range items {
		if isVisibleToGroups(item.Model.AllowedGroupIDs, allowedGroupIDs) {
			result = append(result, toUserModel(item.Model))
		}
	}

	return result, nil
}

func loadUserGroupIDs(ctx context.Context, repo *account.Repository, user *account.User) []string {
	if user == nil {
		return nil
	}

	result := make([]string, 0)
	if user.PrimaryGroupID != nil && *user.PrimaryGroupID != "" {
		result = append(result, *user.PrimaryGroupID)
	}

	var memberGroupIDs []string
	repo.DB().
		WithContext(ctx).
		Model(&account.GroupMember{}).
		Where("user_id = ?", user.ID).
		Pluck("group_id", &memberGroupIDs)

	for _, groupID := range memberGroupIDs {
		if groupID != "" && !slices.Contains(result, groupID) {
			result = append(result, groupID)
		}
	}

	return result
}

func isVisibleToGroups(allowedGroupIDs, userGroupIDs []string) bool {
	if len(allowedGroupIDs) == 0 {
		return true
	}
	for _, allowed := range allowedGroupIDs {
		if slices.Contains(userGroupIDs, allowed) {
			return true
		}
	}
	return false
}

func toUserModels(items []ModelWithBindings) []UserModel {
	result := make([]UserModel, 0, len(items))
	for _, item := range items {
		result = append(result, toUserModel(item.Model))
	}
	return result
}

func toUserModel(item Model) UserModel {
	return UserModel{
		ID:              item.ID,
		ModelKey:        item.ModelKey,
		DisplayName:     item.DisplayName,
		ProviderType:    item.ProviderType,
		ContextLength:   item.ContextLength,
		MaxOutputTokens: item.MaxOutputTokens,
		Pricing:         item.Pricing,
		Capabilities:    item.Capabilities,
		Status:          item.Status,
	}
}
