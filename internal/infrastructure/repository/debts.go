package repository

import (
	"context"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/pkg/database/postgres"
	"github.com/google/uuid"
	"github.com/shogo82148/pointer"
	"github.com/uptrace/bun"
)

type Debts struct {
	bun.BaseModel `bun:"table:debts,alias:d"`

	ID            string     `bun:"id,type:uuid,pk"`
	UserID        string     `bun:"user_id,type:uuid"`
	TransactionID string     `bun:"transaction_id,type:uuid"`
	Type          string     `bun:"type"`
	Status        string     `bun:"status"`
	Amount        int64      `bun:"amount"`
	CurrencyCode  string     `bun:"currency_code"`
	RemindAt      *time.Time `bun:"remind_at,nullzero"`
	PaidAt        *time.Time `bun:"paid_at,nullzero"`
	CreatedAt     time.Time  `bun:"created_at,default:current_timestamp"`
	UpdatedAt     *time.Time `bun:"updated_at,nullzero"`
}

type debtsRepo struct {
	db bun.IDB
}

func NewDebtsRepo(db bun.IDB) entities.DebtRepository {
	return &debtsRepo{
		db: db,
	}
}

func (r *debtsRepo) Save(ctx context.Context, debt *entities.Debt) error {
	db := postgres.FromContext(ctx, r.db)
	var model = r.ToModel(debt)

	_, err := db.NewInsert().Model(model).
		On("CONFLICT (id) DO UPDATE").
		Set("user_id = EXCLUDED.user_id").
		Set("transaction_id = EXCLUDED.transaction_id").
		Set("type = EXCLUDED.type").
		Set("status = EXCLUDED.status").
		Set("amount = EXCLUDED.amount").
		Set("currency_code = EXCLUDED.currency_code").
		Set("remind_at = EXCLUDED.remind_at").
		Set("paid_at = EXCLUDED.paid_at").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	if err != nil {
		return postgres.Error(err, model)
	}

	return err
}

func (r *debtsRepo) GetByID(ctx context.Context, debtID uuid.UUID) (*entities.Debt, error) {
	db := postgres.FromContext(ctx, r.db)

	var model Debts
	err := db.NewSelect().Model(&model).
		Where("id = ?", debtID.String()).
		Scan(ctx)
	if err != nil {
		return nil, postgres.Error(err, model)
	}

	return r.ToEntity(&model), nil
}

func (r *debtsRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Debt, error) {
	db := postgres.FromContext(ctx, r.db)

	var models []Debts
	err := db.NewSelect().Model(&models).
		Where("user_id = ?", userID.String()).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, postgres.Error(err, models)
	}

	var debts []*entities.Debt
	for _, model := range models {
		debts = append(debts, r.ToEntity(&model))
	}

	return debts, nil
}

func (r *debtsRepo) GetByTransactionID(ctx context.Context, transactionID uuid.UUID) (*entities.Debt, error) {
	db := postgres.FromContext(ctx, r.db)

	var model Debts
	err := db.NewSelect().Model(&model).
		Where("transaction_id = ?", transactionID.String()).
		Scan(ctx)
	if err != nil {
		return nil, postgres.Error(err, model)
	}

	return r.ToEntity(&model), nil
}

func (r *debtsRepo) GetByFilter(ctx context.Context, filter *entities.DebtFilter) ([]*entities.Debt, int, error) {
	db := postgres.FromContext(ctx, r.db)

	var models []Debts
	query := db.NewSelect().Model(&models).
		Where("user_id = ?", filter.UserID.String()).
		Order("created_at DESC")

	if len(filter.TransactionIDs) > 0 {
		query = query.Where("transaction_id IN (?)", bun.In(filter.TransactionIDs))
	}

	if len(filter.Types) > 0 {
		query = query.Where("type IN (?)", bun.In(filter.Types))
	}

	if len(filter.Statuses) > 0 {
		query = query.Where("status IN (?)", bun.In(filter.Statuses))
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, 0, postgres.Error(err, models)
	}

	var debts []*entities.Debt
	for _, model := range models {
		debts = append(debts, r.ToEntity(&model))
	}

	count, err := query.Count(ctx)
	if err != nil {
		return nil, 0, postgres.Error(err, models)
	}

	return debts, count, nil
}

func (r *debtsRepo) GetDueReminders(ctx context.Context, before time.Time) ([]*entities.Debt, error) {
	db := postgres.FromContext(ctx, r.db)

	var models []Debts
	err := db.NewSelect().Model(&models).
		Where("remind_at IS NOT NULL").
		Where("remind_at <= ?", before).
		Where("status = ?", entities.Open.String()).
		Order("remind_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, postgres.Error(err, models)
	}

	var debts []*entities.Debt
	for _, model := range models {
		debts = append(debts, r.ToEntity(&model))
	}

	return debts, nil
}

func (r *debtsRepo) Delete(ctx context.Context, debt *entities.Debt) error {
	db := postgres.FromContext(ctx, r.db)

	_, err := db.NewDelete().
		Model((*Debts)(nil)).
		Where("id = ?", debt.ID.String()).
		Exec(ctx)
	if err != nil {
		return postgres.Error(err, Debts{})
	}

	return nil
}

func (r *debtsRepo) ToModel(e *entities.Debt) *Debts {
	if e == nil {
		return nil
	}

	debt := &Debts{
		ID:            e.ID.String(),
		UserID:        e.UserID.String(),
		TransactionID: e.TransactionID.String(),
		Type:          e.Type.String(),
		Status:        e.Status.String(),
		Amount:        e.Amount,
		CurrencyCode:  e.CurrencyCode.String(),
		RemindAt:      pointer.TimeOrNil(e.RemindAt),
		PaidAt:        pointer.TimeOrNil(e.PaidAt),
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     pointer.TimeOrNil(e.UpdatedAt),
	}

	return debt
}

func (r *debtsRepo) ToEntity(m *Debts) *entities.Debt {
	if m == nil {
		return nil
	}

	id, _ := uuid.Parse(m.ID)
	userID, _ := uuid.Parse(m.UserID)
	transactionID, _ := uuid.Parse(m.TransactionID)

	e := &entities.Debt{
		ID:            id,
		UserID:        userID,
		TransactionID: transactionID,
		Type:          entities.DebtType(m.Type),
		Status:        entities.DebtStatus(m.Status),
		Amount:        m.Amount,
		CurrencyCode:  entities.Currency(m.CurrencyCode),
		RemindAt:      pointer.TimeValue(m.RemindAt),
		PaidAt:        pointer.TimeValue(m.PaidAt),
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     pointer.TimeValue(m.UpdatedAt),
	}

	return e
}
