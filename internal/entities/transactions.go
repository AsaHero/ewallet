package entities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TrnType string

const (
	Deposit    TrnType = "deposit"
	Withdrawal TrnType = "withdrawal"
	Transfer   TrnType = "transfer"
	Adjustment TrnType = "adjustment"
)

func (t TrnType) String() string {
	return string(t)
}

type TrnStatus string

const (
	New       TrnStatus = "new"
	Pending   TrnStatus = "pending"
	Completed TrnStatus = "success"
	Rejected  TrnStatus = "rejected"
)

func (t TrnStatus) String() string {
	return string(t)
}

type Transaction struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	AccountID            uuid.UUID
	Category             Category
	Type                 TrnType
	Status               TrnStatus
	Amount               int64
	CurrencyCode         Currency
	OriginalAmount       int64
	OriginalCurrencyCode Currency
	FxRate               float64
	RowText              string
	PerformedAt          time.Time
	RejectedAt           time.Time
	CreatedAt            time.Time
}

func NewTransaction(
	userID uuid.UUID,
	accountID uuid.UUID,
	category Category,
	trnType TrnType,
	amount int64,
	currencyCode Currency,
	rowText string,
) (*Transaction, error) {
	if userID == uuid.Nil {
		return nil, errors.New("invalid user id")
	}
	if accountID == uuid.Nil {
		return nil, errors.New("invalid account id")
	}
	if amount < 0 {
		return nil, errors.New("invalid amount")
	}

	return &Transaction{
		ID:           uuid.New(),
		UserID:       userID,
		AccountID:    accountID,
		Category:     category,
		Type:         trnType,
		Status:       New,
		Amount:       amount,
		CurrencyCode: currencyCode,
		RowText:      rowText,
		CreatedAt:    time.Now(),
	}, nil
}

func (t *Transaction) SetAmountMajor(major float64) error {
	if major <= 0 {
		return fmt.Errorf("amount must be > 0")
	}

	t.Amount = MinorFromMajor(major, t.CurrencyCode.Scale())
	return nil
}

func (t *Transaction) SetAmountMinor(minor int64, currency Currency) error {
	if minor <= 0 {
		return fmt.Errorf("amount must be > 0")
	}

	t.Amount = minor
	return nil
}

func (t *Transaction) AmountMinor() int64 {
	return t.Amount
}

func (t *Transaction) AmountMajor() float64 {
	return MajorFromMinor(t.Amount, t.CurrencyCode.Scale())
}

func (t *Transaction) Performed(performedAt time.Time) {
	t.Status = Completed
	t.PerformedAt = performedAt
}

// Repository

type TransactionRepository interface {
	Save(ctx context.Context, transaction *Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*Transaction, error)
	GetByUserID(ctx context.Context, limit, offset int, userID uuid.UUID) ([]*Transaction, int, error)
	GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]*Transaction, error)
	GetTotalByType(ctx context.Context, userID uuid.UUID, trnType TrnType) (int64, error)
	GetTotalsByCategories(ctx context.Context, userID uuid.UUID) (map[int]int64, []int, error)
}
