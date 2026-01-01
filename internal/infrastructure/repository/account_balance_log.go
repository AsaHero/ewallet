package repository

import (
	"context"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/pkg/database/postgres"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type AccountBalanceLog struct {
	bun.BaseModel `bun:"table:account_balance_log,alias:abl"`

	ID            int64     `bun:"id,pk,autoincrement"`
	UserID        string    `bun:"user_id,type:uuid"`
	AccountID     string    `bun:"account_id,type:uuid"`
	TransactionID string    `bun:"transaction_id,type:uuid"`
	Delta         int64     `bun:"delta"`
	BalanceBefore int64     `bun:"balance_before"`
	BalanceAfter  int64     `bun:"balance_after"`
	OccurredAt    time.Time `bun:"occurred_at"`
	CreatedAt     time.Time `bun:"created_at,default:current_timestamp"`
}

type accountBalanceLogRepo struct {
	db bun.IDB
}

func NewAccountBalanceLogRepo(db bun.IDB) entities.AccountBalanceLogRepository {
	return &accountBalanceLogRepo{
		db: db,
	}
}

func (r *accountBalanceLogRepo) Save(ctx context.Context, e *entities.AccountBalanceLog) error {
	db := postgres.FromContext(ctx, r.db)
	var model = r.ToModel(e)

	var id int64
	err := db.NewInsert().Model(model).
		On("CONFLICT (id) DO UPDATE").
		Set("user_id = EXCLUDED.user_id").
		Set("account_id = EXCLUDED.account_id").
		Set("transaction_id = EXCLUDED.transaction_id").
		Set("delta = EXCLUDED.delta").
		Set("balance_before = EXCLUDED.balance_before").
		Set("balance_after = EXCLUDED.balance_after").
		Set("occurred_at = EXCLUDED.occurred_at").
		Returning("id").
		Scan(ctx, &id)
	if err != nil {
		return postgres.Error(err, model)
	}

	e.ID = id

	return err
}

func (r *accountBalanceLogRepo) FindByID(ctx context.Context, id int64) (*entities.AccountBalanceLog, error) {
	db := postgres.FromContext(ctx, r.db)

	var model AccountBalanceLog
	err := db.NewSelect().Model(&model).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, postgres.Error(err, model)
	}

	return r.ToEntity(&model), nil
}

func (r *accountBalanceLogRepo) FindAllByAccountID(ctx context.Context, accountID uuid.UUID) ([]*entities.AccountBalanceLog, error) {
	db := postgres.FromContext(ctx, r.db)

	var model []AccountBalanceLog
	err := db.NewSelect().Model(&model).
		Where("account_id = ?", accountID.String()).
		Scan(ctx)
	if err != nil {
		return nil, postgres.Error(err, model)
	}

	var logs []*entities.AccountBalanceLog
	for _, m := range model {
		logs = append(logs, r.ToEntity(&m))
	}

	return logs, nil
}

func (r *accountBalanceLogRepo) FindIntervalByAccountID(ctx context.Context, accountID uuid.UUID, from time.Time, to time.Time) ([]*entities.AccountBalanceLog, error) {
	db := postgres.FromContext(ctx, r.db)

	var model []AccountBalanceLog
	err := db.NewSelect().Model(&model).
		Where("account_id = ?", accountID.String()).
		Where("occurred_at >= ?", from).
		Where("occurred_at <= ?", to).
		Scan(ctx)
	if err != nil {
		return nil, postgres.Error(err, model)
	}

	var logs []*entities.AccountBalanceLog
	for _, m := range model {
		logs = append(logs, r.ToEntity(&m))
	}

	return logs, nil
}

func (r *accountBalanceLogRepo) ToModel(e *entities.AccountBalanceLog) *AccountBalanceLog {
	if e == nil {
		return nil
	}

	return &AccountBalanceLog{
		ID:            e.ID,
		UserID:        e.UserID.String(),
		AccountID:     e.AccountID.String(),
		TransactionID: e.TransactionID.String(),
		Delta:         e.Delta,
		BalanceBefore: e.BalanceBefore,
		BalanceAfter:  e.BalanceAfter,
		OccurredAt:    e.OccurredAt,
		CreatedAt:     e.CreatedAt,
	}
}

func (r *accountBalanceLogRepo) ToEntity(m *AccountBalanceLog) *entities.AccountBalanceLog {
	if m == nil {
		return nil
	}

	userID, _ := uuid.Parse(m.UserID)
	accountID, _ := uuid.Parse(m.AccountID)
	transactionID, _ := uuid.Parse(m.TransactionID)

	return &entities.AccountBalanceLog{
		ID:            m.ID,
		UserID:        userID,
		AccountID:     accountID,
		TransactionID: transactionID,
		Delta:         m.Delta,
		BalanceBefore: m.BalanceBefore,
		BalanceAfter:  m.BalanceAfter,
		OccurredAt:    m.OccurredAt,
		CreatedAt:     m.CreatedAt,
	}
}
