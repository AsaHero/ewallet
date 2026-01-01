package entities

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type AccountBalanceLog struct {
	ID            int64     `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	AccountID     uuid.UUID `json:"account_id"`
	TransactionID uuid.UUID `json:"transaction_id"`
	Delta         int64     `json:"delta"`
	BalanceBefore int64     `json:"balance_before"`
	BalanceAfter  int64     `json:"balance_after"`
	OccurredAt    time.Time `json:"occurred_at"`
	CreatedAt     time.Time `json:"created_at"`
}

func NewAccountBalanceLog(
	userID uuid.UUID,
	accountID uuid.UUID,
	transactionID uuid.UUID,
	balanceBefore int64,
	balanceAfter int64,
	occurredAt time.Time,
) (*AccountBalanceLog, error) {
	if userID == uuid.Nil || accountID == uuid.Nil || transactionID == uuid.Nil {
		return nil, errors.New("userID, accountID, and transactionID must not be nil")
	}

	return &AccountBalanceLog{
		UserID:        userID,
		AccountID:     accountID,
		TransactionID: transactionID,
		Delta:         balanceAfter - balanceBefore,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		OccurredAt:    occurredAt,
		CreatedAt:     time.Now(),
	}, nil
}

// Repository
type AccountBalanceLogRepository interface {
	Save(ctx context.Context, e *AccountBalanceLog) error
	FindByID(ctx context.Context, id int64) (*AccountBalanceLog, error)
	FindAllByAccountID(ctx context.Context, accountID uuid.UUID) ([]*AccountBalanceLog, error)
	FindIntervalByAccountID(ctx context.Context, accountID uuid.UUID, from time.Time, to time.Time) ([]*AccountBalanceLog, error)
}
