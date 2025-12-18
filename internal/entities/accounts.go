package entities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Name      string
	Balance   int64
	IsDefault bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewAccount(userID uuid.UUID, name string) (*Account, error) {
	if userID == uuid.Nil {
		return nil, errors.New("invalid user id")
	}

	return &Account{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      name,
		Balance:   0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (t *Account) SetAmountMajor(major float64, currency Currency) error {
	if major < 0 {
		return fmt.Errorf("amount must be > 0")
	}

	t.Balance = MinorFromMajor(major, currency.Scale())
	return nil
}

func (t *Account) SetAmountMinor(minor int64, currency Currency) error {
	if minor < 0 {
		return fmt.Errorf("amount must be > 0")
	}

	t.Balance = minor
	return nil
}

func (t *Account) AmountMinor() int64 {
	return t.Balance
}

func (t *Account) AmountMajor(currency Currency) float64 {
	return MajorFromMinor(t.Balance, currency.Scale())
}

func (t *Account) UpdateDefault(isDefault bool) {
	t.IsDefault = isDefault
	t.UpdatedAt = time.Now()
}

func (t *Account) UpdateName(name string) {
	t.Name = name
	t.UpdatedAt = time.Now()
}

func (t *Account) ApplyTransaction(transaction *Transaction) error {
	if transaction == nil {
		return nil
	}

	switch transaction.Type {
	case Deposit:
		t.Balance += transaction.AmountMinor()
	case Withdrawal:
		t.Balance -= transaction.AmountMinor()
	}

	t.UpdatedAt = time.Now()

	return nil
}

func (t *Account) RevertTransaction(transaction *Transaction) error {
	if transaction == nil {
		return nil
	}

	switch transaction.Type {
	case Deposit:
		t.Balance -= transaction.AmountMinor()
	case Withdrawal:
		t.Balance += transaction.AmountMinor()
	}

	t.UpdatedAt = time.Now()

	return nil
}

// Domain Service
type AccountsService struct {
	repo AccountRepository
}

func NewAccountsService(repo AccountRepository) *AccountsService {
	return &AccountsService{
		repo: repo,
	}
}

func (s *AccountsService) MakeDefault(ctx context.Context, account *Account) error {
	allAccounts, err := s.repo.GetByUserID(ctx, account.UserID)
	if err != nil {
		return err
	}

	for _, a := range allAccounts {
		if a.ID != account.ID {
			a.UpdateDefault(false)
		}

		err = s.repo.Save(ctx, a)
		if err != nil {
			return err
		}
	}

	account.UpdateDefault(true)
	err = s.repo.Save(ctx, account)
	if err != nil {
		return err
	}

	return err
}

// Repository
type AccountRepository interface {
	Save(ctx context.Context, account *Account) error
	GetByID(ctx context.Context, id uuid.UUID) (*Account, error)
	GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*Account, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Account, error)
	GetTotalBalance(ctx context.Context, userID uuid.UUID) (int64, error)
	Delete(ctx context.Context, account *Account) error
}
