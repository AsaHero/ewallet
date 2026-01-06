package entities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type DebtType string

const (
	Borrow DebtType = "borrow"
	Lend   DebtType = "lend"
)

func (t DebtType) String() string {
	return string(t)
}

type DebtStatus string

const (
	Open      DebtStatus = "open"
	Paid      DebtStatus = "paid"
	Cancelled DebtStatus = "cancelled"
)

func (s DebtStatus) String() string {
	return string(s)
}

type Debt struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	TransactionID uuid.UUID
	Type          DebtType
	Status        DebtStatus
	Name          string
	Amount        int64
	CurrencyCode  Currency
	Note          string
	DueAt         time.Time
	PaidAt        time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewDebt(
	userID uuid.UUID,
	transactionID uuid.UUID,
	debtType DebtType,
) (*Debt, error) {
	if userID == uuid.Nil {
		return nil, errors.New("invalid user id")
	}
	if transactionID == uuid.Nil {
		return nil, errors.New("invalid transaction id")
	}

	return &Debt{
		ID:            uuid.New(),
		UserID:        userID,
		TransactionID: transactionID,
		Type:          debtType,
		Status:        Open,
		CreatedAt:     time.Now(),
	}, nil
}

func (d *Debt) SetAmountMajor(major float64, currency Currency) error {
	if currency == "" {
		return fmt.Errorf("currency code must not be empty")
	}

	d.Amount = MinorFromMajor(major, currency.Scale())
	d.CurrencyCode = currency
	return nil
}

func (d *Debt) SetAmountMinor(minor int64, currency Currency) error {
	if currency == "" {
		return fmt.Errorf("currency code must not be empty")
	}

	d.Amount = minor
	d.CurrencyCode = currency
	return nil
}

func (d *Debt) SetName(name string) {
	d.Name = name
	d.UpdatedAt = time.Now()
}

func (d *Debt) SetNote(note string) {
	d.Note = note
	d.UpdatedAt = time.Now()
}

func (d *Debt) AmountMinor() int64 {
	return d.Amount
}

func (d *Debt) AmountMajor() float64 {
	return MajorFromMinor(d.Amount, d.CurrencyCode.Scale())
}

func (d *Debt) SetDueAt(dueAt time.Time) {
	d.DueAt = dueAt
	d.UpdatedAt = time.Now()
}

func (d *Debt) MarkAsPaid(paidAt time.Time) {
	d.Status = Paid
	d.PaidAt = paidAt
	d.UpdatedAt = time.Now()
}

func (d *Debt) Cancel() {
	d.Status = Cancelled
	d.UpdatedAt = time.Now()
}

func (d *Debt) Update(
	amount *int64,
	name *string,
	dueAt *time.Time,
	note *string,
) error {
	if amount != nil {
		d.Amount = *amount
	}
	if name != nil {
		d.Name = *name
	}
	if note != nil {
		d.Note = *note
	}

	if dueAt != nil {
		d.DueAt = *dueAt
	}
	d.UpdatedAt = time.Now()
	return nil
}

// Repository

type DebtFilter struct {
	UserID         uuid.UUID
	TransactionIDs []uuid.UUID
	Types          []DebtType
	Statuses       []DebtStatus
	Limit          int
	Offset         int
}

type DebtRepository interface {
	Save(ctx context.Context, debt *Debt) error
	GetByID(ctx context.Context, id uuid.UUID) (*Debt, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Debt, error)
	GetByTransactionID(ctx context.Context, transactionID uuid.UUID) (*Debt, error)
	GetByFilter(ctx context.Context, filter *DebtFilter) ([]*Debt, int, error)
	GetDueReminders(ctx context.Context, before time.Time) ([]*Debt, error)
	Delete(ctx context.Context, debt *Debt) error
}
