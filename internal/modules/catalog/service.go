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

	result := make([]UserModel, 0, len(items))
	for _, item := range items {
		if isVisibleToUserGroup(item.Model.VisibleUserGroupIDs, user.UserGroupID) {
			result = append(result, toUserModel(item.Model))
		}
	}

	return result, nil
}

func isVisibleToUserGroup(visibleUserGroupIDs []string, userGroupID *string) bool {
	if len(visibleUserGroupIDs) == 0 {
		return true
	}
	if userGroupID == nil || *userGroupID == "" {
		return false
	}

	return slices.Contains(visibleUserGroupIDs, *userGroupID)
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
