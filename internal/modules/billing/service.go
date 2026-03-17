package billing

import (
	"context"
	"errors"

	"mrchat/internal/modules/account"
)

var ErrInvalidBillingLogType = errors.New("invalid billing log type")

type Service struct {
	repo *account.Repository
}

type LogItem struct {
	ID           string               `json:"id"`
	Type         account.QuotaLogType `json:"type"`
	DeltaQuota   int64                `json:"delta_quota"`
	BalanceAfter int64                `json:"balance_after"`
	Reason       *string              `json:"reason"`
	CreatedAt    string               `json:"created_at"`
}

type LogList struct {
	Items []LogItem
	Total int64
}

func NewService(repo *account.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetSummary(ctx context.Context, userID string) (account.BillingSummary, error) {
	return s.repo.GetBillingSummary(ctx, userID)
}

func (s *Service) ListLogs(ctx context.Context, userID string, page, pageSize int, filter string) (LogList, error) {
	logType, err := mapBillingLogFilter(filter)
	if err != nil {
		return LogList{}, err
	}

	result, err := s.repo.ListQuotaLogs(ctx, userID, page, pageSize, logType)
	if err != nil {
		return LogList{}, err
	}

	items := make([]LogItem, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, LogItem{
			ID:           item.ID,
			Type:         item.LogType,
			DeltaQuota:   item.DeltaQuota,
			BalanceAfter: item.BalanceAfter,
			Reason:       item.Reason,
			CreatedAt:    item.CreatedAt.UTC().Format(timeLayout),
		})
	}

	return LogList{
		Items: items,
		Total: result.Total,
	}, nil
}

func mapBillingLogFilter(filter string) (account.QuotaLogType, error) {
	switch filter {
	case "", "all":
		return "", nil
	case "consume":
		return account.QuotaLogTypeFinalCharge, nil
	case "refund":
		return account.QuotaLogTypeRefund, nil
	case "redeem":
		return account.QuotaLogTypeRedeem, nil
	case "admin_adjust":
		return account.QuotaLogTypeAdminAdjust, nil
	default:
		return "", ErrInvalidBillingLogType
	}
}

const timeLayout = "2006-01-02T15:04:05Z07:00"
