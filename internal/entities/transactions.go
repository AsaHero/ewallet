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
	Category             *Category
	Subcategory          *Subcategory
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
	trnType TrnType,
	rowText string,
) (*Transaction, error) {
	if userID == uuid.Nil {
		return nil, errors.New("invalid user id")
	}
	if accountID == uuid.Nil {
		return nil, errors.New("invalid account id")
	}

	return &Transaction{
		ID:        uuid.New(),
		UserID:    userID,
		AccountID: accountID,
		Type:      trnType,
		Status:    New,
		RowText:   rowText,
		CreatedAt: time.Now(),
	}, nil
}

func (t *Transaction) Categorise(category *Category, subcategory *Subcategory) error {
	if category != nil {
		t.Category = category
	}
	if subcategory != nil {
		t.Subcategory = subcategory
	}

	return nil
}

func (t *Transaction) SetAmountMajor(major float64, currency Currency) error {
	if major < 0 {
		return fmt.Errorf("amount must be > 0")
	}

	if currency == "" {
		return fmt.Errorf("currency code must not be empty")
	}

	t.Amount = MinorFromMajor(major, currency.Scale())
	t.CurrencyCode = currency
	return nil
}

func (t *Transaction) SetAmountMinor(minor int64, currency Currency) error {
	if minor < 0 {
		return fmt.Errorf("amount must be > 0")
	}

	if currency == "" {
		return fmt.Errorf("currency code must not be empty")
	}

	t.Amount = minor
	t.CurrencyCode = currency
	return nil
}

func (t *Transaction) AmountMinor() int64 {
	return t.Amount
}

func (t *Transaction) AmountMajor() float64 {
	return MajorFromMinor(t.Amount, t.CurrencyCode.Scale())
}

func (t *Transaction) SetOriginalAmountMajor(major float64, currency Currency) error {
	if major < 0 {
		return fmt.Errorf("amount must be > 0")
	}

	if currency == "" {
		return fmt.Errorf("original currency code must not be empty")
	}

	t.OriginalAmount = MinorFromMajor(major, currency.Scale())
	t.OriginalCurrencyCode = currency
	return nil
}

func (t *Transaction) SetOriginalAmountMinor(minor int64, currency Currency) error {
	if minor < 0 {
		return fmt.Errorf("amount must be > 0")
	}

	if currency == "" {
		return fmt.Errorf("original currency code must not be empty")
	}

	t.OriginalAmount = minor
	t.OriginalCurrencyCode = currency
	return nil
}

func (t *Transaction) OriginalAmountMinor() int64 {
	return t.OriginalAmount
}

func (t *Transaction) OriginalAmountMajor() float64 {
	return MajorFromMinor(t.OriginalAmount, t.OriginalCurrencyCode.Scale())
}

func (t *Transaction) SetFxRate(rate float64) error {
	if rate <= 0 {
		return fmt.Errorf("fx rate must be > 0")
	}

	t.FxRate = rate
	return nil
}

func (t *Transaction) Performed(performedAt time.Time) {
	t.Status = Completed
	t.PerformedAt = performedAt
}

func (t *Transaction) Update(
	category *Category,
	subcategory *Subcategory,
	trnType TrnType,
	amount float64,
	currency Currency,
	originalAmount *float64,
	originalCurrency *string,
	fxRate *float64,
	rowText string,
	performedAt *time.Time,
) error {
	t.Type = trnType
	t.RowText = rowText

	err := t.Categorise(category, subcategory)
	if err != nil {
		return err
	}

	err = t.SetAmountMajor(amount, currency)
	if err != nil {
		return err
	}

	if originalAmount != nil && originalCurrency != nil {
		err = t.SetOriginalAmountMajor(*originalAmount, Currency(*originalCurrency))
		if err != nil {
			return err
		}

		if fxRate != nil {
			err = t.SetFxRate(*fxRate)
			if err != nil {
				return err
			}
		}
	} else {
		// User deleted original amount data
		t.OriginalAmount = 0
		t.OriginalCurrencyCode = ""
		t.FxRate = 0
	}

	if performedAt != nil {
		t.Performed(*performedAt)
	} else if t.PerformedAt.IsZero() {
		t.Performed(time.Now())
	}

	return nil
}

// Repository

type TransactionRepository interface {
	Save(ctx context.Context, transaction *Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*Transaction, error)
	GetByUserID(ctx context.Context, limit, offset int, userID uuid.UUID, trnType []TrnType) ([]*Transaction, int, error)
	GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]*Transaction, error)
	GetTotalByType(ctx context.Context, userID uuid.UUID, trnType TrnType, from, to *time.Time) (int64, error)
	GetTotalByTypeAndAccount(ctx context.Context, userID uuid.UUID, accountID *uuid.UUID, trnType TrnType, from, to *time.Time) (int64, error)
	GetTotalsByCategories(ctx context.Context, userID uuid.UUID, trnType TrnType, from, to *time.Time) (map[int]int64, []int, error)
	GetTotalsByCategoriesAndAccount(ctx context.Context, userID uuid.UUID, accountID *uuid.UUID, trnType TrnType, from, to *time.Time) (map[int]int64, []int, error)
	GetAllBetween(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*Transaction, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
