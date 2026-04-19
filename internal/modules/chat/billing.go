package chat

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
)

var ErrInsufficientQuota = errors.New("insufficient quota")

const (
	chatQuotaReserveReason = "chat.reserve_quota"
	chatQuotaRefundReason  = "chat.release_reserved_quota"
	chatQuotaChargeReason  = "chat.final_charge"
)

func (s *Service) reserveQuota(ctx context.Context, userID, requestID string, reservedQuota int64) error {
	if reservedQuota <= 0 {
		return nil
	}

	return s.accountRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user, err := s.lockUserForQuota(ctx, tx, userID)
		if err != nil {
			return err
		}

		if user.Quota < reservedQuota {
			return ErrInsufficientQuota
		}

		nextQuota := user.Quota - reservedQuota
		now := time.Now().UTC()
		user.Quota = nextQuota
		user.UpdatedAt = now
		if err := tx.WithContext(ctx).Save(user).Error; err != nil {
			return fmt.Errorf("reserve user quota: %w", err)
		}

		if err := s.createQuotaLogWithDB(ctx, tx, user.ID, requestID, account.QuotaLogTypePreDeduct, -reservedQuota, nextQuota, chatQuotaReserveReason, now); err != nil {
			return err
		}

		return nil
	})
}

func (s *Service) refundReservedQuota(ctx context.Context, userID, requestID string, reservedQuota int64) error {
	if reservedQuota <= 0 {
		return nil
	}

	return s.accountRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		_, err := s.settleReservedQuotaWithDB(ctx, tx, userID, requestID, reservedQuota, 0)
		return err
	})
}

func (s *Service) settleReservedQuotaWithDB(ctx context.Context, tx *gorm.DB, userID, requestID string, reservedQuota, finalCharge int64) (CompletionBilling, error) {
	billing := CompletionBilling{
		PreDeducted:  reservedQuota,
		FinalCharged: finalCharge,
	}

	if reservedQuota <= 0 && finalCharge <= 0 {
		return billing, nil
	}
	if tx == nil {
		tx = s.accountRepo.DB()
	}

	user, err := s.lockUserForQuota(ctx, tx, userID)
	if err != nil {
		return CompletionBilling{}, err
	}

	now := time.Now().UTC()
	currentQuota := user.Quota
	if reservedQuota > 0 {
		currentQuota += reservedQuota
		if err := s.createQuotaLogWithDB(ctx, tx, user.ID, requestID, account.QuotaLogTypeRefund, reservedQuota, currentQuota, chatQuotaRefundReason, now); err != nil {
			return CompletionBilling{}, err
		}
	}

	if finalCharge > 0 {
		if currentQuota < finalCharge {
			return CompletionBilling{}, ErrInsufficientQuota
		}

		currentQuota -= finalCharge
		user.UsedQuota += finalCharge
		if err := s.createQuotaLogWithDB(ctx, tx, user.ID, requestID, account.QuotaLogTypeFinalCharge, -finalCharge, currentQuota, chatQuotaChargeReason, now); err != nil {
			return CompletionBilling{}, err
		}
	}

	user.Quota = currentQuota
	user.UpdatedAt = now
	if err := tx.WithContext(ctx).Save(user).Error; err != nil {
		return CompletionBilling{}, fmt.Errorf("update user quota after settlement: %w", err)
	}

	if reservedQuota > finalCharge {
		billing.Refunded = reservedQuota - finalCharge
	}

	return billing, nil
}

func (s *Service) createQuotaLogWithDB(ctx context.Context, tx *gorm.DB, userID, requestID string, logType account.QuotaLogType, delta, balanceAfter int64, reason string, createdAt time.Time) error {
	if tx == nil {
		tx = s.accountRepo.DB()
	}

	item := &account.QuotaLog{
		ID:           uuid.NewString(),
		UserID:       userID,
		RequestID:    optionalString(requestID),
		LogType:      logType,
		DeltaQuota:   delta,
		BalanceAfter: balanceAfter,
		Reason:       optionalString(reason),
		CreatedAt:    createdAt.UTC(),
	}
	if err := tx.WithContext(ctx).Create(item).Error; err != nil {
		return fmt.Errorf("create quota log: %w", err)
	}

	return nil
}

func (s *Service) lockUserForQuota(ctx context.Context, tx *gorm.DB, userID string) (*account.User, error) {
	if tx == nil {
		tx = s.accountRepo.DB()
	}

	var user account.User
	if err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, account.ErrUserNotFound
		}
		return nil, fmt.Errorf("lock user for quota: %w", err)
	}

	return &user, nil
}

func optionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
