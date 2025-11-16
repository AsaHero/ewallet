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

type Accounts struct {
	bun.BaseModel `bun:"table:accounts,alias:a"`

	ID        string     `bun:"id,type:uuid,pk"`
	UserID    string     `bun:"user_id,type:uuid"`
	Name      string     `bun:"name"`
	Balance   int64      `bun:"balance"`
	IsDefault bool       `bun:"is_default"`
	CreatedAt time.Time  `bun:"created_at,default:current_timestamp"`
	UpdatedAt *time.Time `bun:"updated_at,nullzero"`
}

type accountsRepo struct {
	db bun.IDB
}

func NewAccountsRepo(db bun.IDB) entities.AccountRepository {
	return &accountsRepo{
		db: db,
	}
}

func (r *accountsRepo) Save(ctx context.Context, account *entities.Account) error {
	db := postgres.FromContext(ctx, r.db)
	var model = r.ToModel(account)

	_, err := db.NewInsert().Model(model).
		On("CONFLICT (id) DO UPDATE").
		Set("user_id = EXCLUDED.user_id").
		Set("name = EXCLUDED.name").
		Set("balance = EXCLUDED.balance").
		Set("is_default = EXCLUDED.is_default").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	if err != nil {
		return postgres.Error(err, model)
	}

	return err
}
func (r *accountsRepo) GetByID(ctx context.Context, accountID uuid.UUID) (*entities.Account, error) {
	db := postgres.FromContext(ctx, r.db)

	var model Accounts
	err := db.NewSelect().Model(&model).
		Where("id = ?", accountID.String()).
		Scan(ctx)
	if err != nil {
		return nil, postgres.Error(err, model)
	}

	return r.ToEntity(&model), nil
}

func (r *accountsRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Account, error) {
	db := postgres.FromContext(ctx, r.db)

	var model []Accounts
	err := db.NewSelect().Model(&model).
		Where("user_id = ?", userID.String()).
		Scan(ctx)
	if err != nil {
		return nil, postgres.Error(err, model)
	}

	var accounts []*entities.Account
	for _, m := range model {
		accounts = append(accounts, r.ToEntity(&m))
	}

	return accounts, nil
}

func (r *accountsRepo) GetTotalBalance(ctx context.Context, userID uuid.UUID) (int64, error) {
	db := postgres.FromContext(ctx, r.db)

	var total int64
	err := db.NewSelect().
		Model((*Accounts)(nil)).
		ColumnExpr("COALESCE(SUM(balance), 0)").
		Where("user_id = ?", userID.String()).
		Scan(ctx, &total)
	if err != nil {
		return 0, postgres.Error(err, Accounts{})
	}

	return total, nil
}
func (r *accountsRepo) Delete(ctx context.Context, account *entities.Account) error {
	db := postgres.FromContext(ctx, r.db)
	var model Accounts

	_, err := db.NewDelete().Model(&model).
		Where("id = ?", account.ID.String()).
		Exec(ctx)
	if err != nil {
		return postgres.Error(err, model)
	}

	return err
}

func (r *accountsRepo) ToModel(e *entities.Account) *Accounts {
	if e == nil {
		return nil
	}

	accounts := &Accounts{
		ID:        e.ID.String(),
		UserID:    e.UserID.String(),
		Name:      e.Name,
		Balance:   e.Balance,
		IsDefault: e.IsDefault,
		CreatedAt: e.CreatedAt,
		UpdatedAt: pointer.TimeOrNil(e.UpdatedAt),
	}

	return accounts
}

func (r *accountsRepo) ToEntity(m *Accounts) *entities.Account {
	if m == nil {
		return nil
	}

	id, _ := uuid.Parse(m.ID)
	userID, _ := uuid.Parse(m.UserID)

	e := &entities.Account{
		ID:        id,
		UserID:    userID,
		Name:      m.Name,
		Balance:   m.Balance,
		IsDefault: m.IsDefault,
		CreatedAt: m.CreatedAt,
		UpdatedAt: pointer.TimeValue(m.UpdatedAt),
	}

	return e
}
