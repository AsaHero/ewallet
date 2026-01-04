package entities

import (
	"context"
	"errors"
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
	t.Balance = MinorFromMajor(major, currency.Scale())
	return nil
}

func (t *Account) SetAmountMinor(minor int64, currency Currency) error {
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

	// Transaction amount is signed: positive for deposit, negative for withdrawal
	t.Balance += transaction.AmountMinor()
	t.UpdatedAt = time.Now()

	return nil
}

func (t *Account) RevertTransaction(transaction *Transaction) error {
	if transaction == nil {
		return nil
	}

	// Transaction amount is signed: reverse by subtracting
	t.Balance -= transaction.AmountMinor()
	t.UpdatedAt = time.Now()

	return nil
}

// Domain Service
type AccountsService struct {
	repo           AccountRepository
	balanceLogRepo AccountBalanceLogRepository
}

func NewAccountsService(
	repo AccountRepository,
	balanceLogRepo AccountBalanceLogRepository,
) *AccountsService {
	return &AccountsService{
		repo:           repo,
		balanceLogRepo: balanceLogRepo,
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

// ApplyTransaction applies a transaction to an account and logs the balance change
func (s *AccountsService) ApplyTransaction(
	ctx context.Context,
	account *Account,
	transaction *Transaction,
) error {
	if account == nil || transaction == nil {
		return errors.New("account and transaction must not be nil")
	}

	// Capture balance before applying transaction
	balanceBefore := account.Balance

	// Apply transaction to account
	err := account.ApplyTransaction(transaction)
	if err != nil {
		return err
	}

	// Capture balance after applying transaction
	balanceAfter := account.Balance

	// Create and save balance log
	balanceLog, err := NewAccountBalanceLog(
		account.UserID,
		account.ID,
		transaction.ID,
		balanceBefore,
		balanceAfter,
		transaction.PerformedAt,
	)
	if err != nil {
		return err
	}

	err = s.balanceLogRepo.Save(ctx, balanceLog)
	if err != nil {
		return err
	}

	// Save updated account
	err = s.repo.Save(ctx, account)
	if err != nil {
		return err
	}

	return nil
}

// RevertTransaction reverts a transaction from an account and logs the balance change
func (s *AccountsService) RevertTransaction(
	ctx context.Context,
	account *Account,
	transaction *Transaction,
) error {
	if account == nil || transaction == nil {
		return errors.New("account and transaction must not be nil")
	}

	if transaction.IsReverted() {
		return nil
	}

	// Capture balance before reverting transaction
	balanceBefore := account.Balance

	// Revert transaction from account
	err := account.RevertTransaction(transaction)
	if err != nil {
		return err
	}

	// Capture balance after reverting transaction
	balanceAfter := account.Balance

	// Create and save balance log
	balanceLog, err := NewAccountBalanceLog(
		account.UserID,
		account.ID,
		transaction.ID,
		balanceBefore,
		balanceAfter,
		transaction.PerformedAt,
	)
	if err != nil {
		return err
	}

	err = s.balanceLogRepo.Save(ctx, balanceLog)
	if err != nil {
		return err
	}

	// Save updated account
	err = s.repo.Save(ctx, account)
	if err != nil {
		return err
	}

	return nil
}

// Repository
type BalanceTimeseriesFilter struct {
	UserID     uuid.UUID
	From       time.Time
	To         time.Time
	AccountIDs []uuid.UUID // if empty, use all user's accounts
	GroupBy    string      // "day", "week", or "month"
	Timezone   string      // user's timezone for bucket boundaries
}

type BalanceTimeseriesPoint struct {
	Timestamp    string // bucket key (e.g., "2026-01-01" or "2026-01")
	BalanceOpen  int64  // opening balance for the bucket
	BalanceClose int64  // closing balance for the bucket
	MinBalance   int64  // min balance observed inside bucket
	MaxBalance   int64  // max balance observed inside bucket
	Delta        int64  // balance_close - balance_open
	TxCount      int    // number of transactions in that bucket
}

type AccountRepository interface {
	Save(ctx context.Context, account *Account) error
	GetByID(ctx context.Context, id uuid.UUID) (*Account, error)
	GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*Account, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Account, error)
	GetTotalBalance(ctx context.Context, userID uuid.UUID) (int64, error)
	Delete(ctx context.Context, account *Account) error
}
