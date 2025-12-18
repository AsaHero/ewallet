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

type Transactions struct {
	bun.BaseModel `bun:"table:transactions,alias:t"`

	ID                   string     `bun:"id,type:uuid,pk"`
	UserID               string     `bun:"user_id,type:uuid"`
	AccountID            string     `bun:"account_id,type:uuid"`
	CategoryID           *int       `bun:"category_id,nullzero"`
	SubcategoryID        *int       `bun:"subcategory_id,nullzero"`
	Type                 string     `bun:"type"`
	Status               string     `bun:"status"`
	Amount               int64      `bun:"amount"`
	CurrencyCode         string     `bun:"currency_code"`
	OriginalAmount       *int64     `bun:"original_amount,nullzero"`
	OriginalCurrencyCode *string    `bun:"original_currency_code,nullzero"`
	FxRate               *float64   `bun:"fx_rate,nullzero"`
	RowText              string     `bun:"row_text"`
	PerformedAt          *time.Time `bun:"performed_at,nullzero"`
	RejectedAt           *time.Time `bun:"rejected_at,nullzero"`
	CreatedAt            time.Time  `bun:"created_at,default:current_timestamp"`
}

type transactionsRepo struct {
	db                bun.IDB
	categoriesRepo    entities.CategoryRepository
	subcategoriesRepo entities.SubcategoryRepository
}

func NewTransactionsRepo(db bun.IDB, categoriesRepo entities.CategoryRepository, subcategoriesRepo entities.SubcategoryRepository) entities.TransactionRepository {
	return &transactionsRepo{
		db:                db,
		categoriesRepo:    categoriesRepo,
		subcategoriesRepo: subcategoriesRepo,
	}
}

func (r *transactionsRepo) Save(ctx context.Context, transaction *entities.Transaction) error {
	db := postgres.FromContext(ctx, r.db)
	var model = r.ToModel(transaction)

	_, err := db.NewInsert().Model(model).
		On("CONFLICT (id) DO UPDATE").
		Set("user_id = EXCLUDED.user_id").
		Set("account_id = EXCLUDED.account_id").
		Set("category_id = EXCLUDED.category_id").
		Set("type = EXCLUDED.type").
		Set("status = EXCLUDED.status").
		Set("amount = EXCLUDED.amount").
		Set("currency_code = EXCLUDED.currency_code").
		Set("original_amount = EXCLUDED.original_amount").
		Set("original_currency_code = EXCLUDED.original_currency_code").
		Set("fx_rate = EXCLUDED.fx_rate").
		Set("row_text = EXCLUDED.row_text").
		Set("performed_at = EXCLUDED.performed_at").
		Set("rejected_at = EXCLUDED.rejected_at").
		Exec(ctx)
	if err != nil {
		return postgres.Error(err, model)
	}

	return err
}

func (r *transactionsRepo) Delete(ctx context.Context, id uuid.UUID) error {
	db := postgres.FromContext(ctx, r.db)

	_, err := db.NewDelete().
		Model((*Transactions)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return postgres.Error(err, Transactions{})
	}

	return nil
}

func (r *transactionsRepo) GetByID(ctx context.Context, transactionID uuid.UUID) (*entities.Transaction, error) {
	db := postgres.FromContext(ctx, r.db)

	var model Transactions
	err := db.NewSelect().Model(&model).
		Where("id = ?", transactionID.String()).
		Scan(ctx)
	if err != nil {
		return nil, postgres.Error(err, model)
	}

	return r.ToEntity(ctx, &model), nil
}

func (r *transactionsRepo) GetByUserID(ctx context.Context, limit, offset int, userID uuid.UUID, trnType []entities.TrnType) ([]*entities.Transaction, int, error) {
	db := postgres.FromContext(ctx, r.db)

	var models []Transactions
	query := db.NewSelect().Model(&models).
		Where("user_id = ?", userID.String()).
		Order("created_at desc")

	if len(trnType) > 0 {
		query = query.Where("type IN (?)", bun.In(trnType))
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, 0, postgres.Error(err, models)
	}

	var transactions []*entities.Transaction
	for _, model := range models {
		transactions = append(transactions, r.ToEntity(ctx, &model))
	}

	count, err := query.Count(ctx)
	if err != nil {
		return nil, 0, postgres.Error(err, models)
	}

	return transactions, count, nil
}

func (r *transactionsRepo) GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]*entities.Transaction, error) {
	db := postgres.FromContext(ctx, r.db)

	var models []Transactions
	err := db.NewSelect().Model(&models).
		Where("account_id = ?", accountID.String()).
		Scan(ctx)
	if err != nil {
		return nil, postgres.Error(err, models)
	}

	var transactions []*entities.Transaction
	for _, model := range models {
		transactions = append(transactions, r.ToEntity(ctx, &model))
	}

	return transactions, nil
}

func (r *transactionsRepo) GetTotalByType(ctx context.Context, userID uuid.UUID, trnType entities.TrnType, from, to *time.Time) (int64, error) {
	db := postgres.FromContext(ctx, r.db)

	var total int64
	query := db.NewSelect().
		Model((*Transactions)(nil)).
		ColumnExpr("COALESCE(SUM(amount), 0)").
		Where("user_id = ?", userID.String()).
		Where("type = ?", trnType.String())

	if from != nil {
		query = query.Where("created_at >= ?", from)
	}
	if to != nil {
		query = query.Where("created_at < ?", to)
	}

	err := query.Scan(ctx, &total)
	if err != nil {
		return 0, postgres.Error(err, Transactions{})
	}

	return total, nil
}

func (r *transactionsRepo) GetTotalsByCategories(ctx context.Context, userID uuid.UUID, trnType entities.TrnType, from, to *time.Time) (map[int]int64, []int, error) {
	db := postgres.FromContext(ctx, r.db)

	var results []struct {
		CategoryID int   `bun:"category_id"`
		Total      int64 `bun:"total"`
	}

	query := db.NewSelect().
		Model((*Transactions)(nil)).
		Column("category_id").
		ColumnExpr("SUM(amount) as total").
		Where("user_id = ?", userID.String()).
		Where("type = ?", trnType.String()).
		Group("category_id").
		Order("total desc")

	if from != nil {
		query = query.Where("created_at >= ?", from)
	}
	if to != nil {
		query = query.Where("created_at < ?", to)
	}

	err := query.Scan(ctx, &results)
	if err != nil {
		return nil, nil, postgres.Error(err, Transactions{})
	}

	totals := make(map[int]int64)
	categories := make([]int, 0, len(results))
	for _, result := range results {
		categories = append(categories, result.CategoryID)
		totals[result.CategoryID] = result.Total
	}

	return totals, categories, nil
}

func (r *transactionsRepo) GetTotalByTypeAndAccount(ctx context.Context, userID uuid.UUID, accountID *uuid.UUID, trnType entities.TrnType, from, to *time.Time) (int64, error) {
	db := postgres.FromContext(ctx, r.db)

	var total int64
	query := db.NewSelect().
		Model((*Transactions)(nil)).
		ColumnExpr("COALESCE(SUM(amount), 0)").
		Where("user_id = ?", userID.String()).
		Where("type = ?", trnType.String())

	if accountID != nil {
		query = query.Where("account_id = ?", accountID.String())
	}

	if from != nil {
		query = query.Where("created_at >= ?", from)
	}
	if to != nil {
		query = query.Where("created_at < ?", to)
	}

	err := query.Scan(ctx, &total)
	if err != nil {
		return 0, postgres.Error(err, Transactions{})
	}

	return total, nil
}

func (r *transactionsRepo) GetTotalsByCategoriesAndAccount(ctx context.Context, userID uuid.UUID, accountID *uuid.UUID, trnType entities.TrnType, from, to *time.Time) (map[int]int64, []int, error) {
	db := postgres.FromContext(ctx, r.db)

	var results []struct {
		CategoryID int   `bun:"category_id"`
		Total      int64 `bun:"total"`
	}

	query := db.NewSelect().
		Model((*Transactions)(nil)).
		Column("category_id").
		ColumnExpr("SUM(amount) as total").
		Where("user_id = ?", userID.String()).
		Where("type = ?", trnType.String()).
		Group("category_id").
		Order("total desc")

	if accountID != nil {
		query = query.Where("account_id = ?", accountID.String())
	}

	if from != nil {
		query = query.Where("created_at >= ?", from)
	}
	if to != nil {
		query = query.Where("created_at < ?", to)
	}

	err := query.Scan(ctx, &results)
	if err != nil {
		return nil, nil, postgres.Error(err, Transactions{})
	}

	totals := make(map[int]int64)
	categories := make([]int, 0, len(results))
	for _, result := range results {
		categories = append(categories, result.CategoryID)
		totals[result.CategoryID] = result.Total
	}

	return totals, categories, nil
}

func (r *transactionsRepo) GetAllBetween(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*entities.Transaction, error) {
	db := postgres.FromContext(ctx, r.db)

	var models []Transactions
	err := db.NewSelect().Model(&models).
		Where("user_id = ?", userID.String()).
		Where("created_at >= ?", from).
		Where("created_at < ?", to).
		Where("status = ?", entities.Completed.String()).
		Scan(ctx)
	if err != nil {
		return nil, postgres.Error(err, models)
	}

	var transactions []*entities.Transaction
	for _, model := range models {
		transactions = append(transactions, r.ToEntity(ctx, &model))
	}

	return transactions, nil
}

func (r *transactionsRepo) ToModel(e *entities.Transaction) *Transactions {
	if e == nil {
		return nil
	}

	transactions := &Transactions{
		ID:                   e.ID.String(),
		UserID:               e.UserID.String(),
		AccountID:            e.AccountID.String(),
		Type:                 e.Type.String(),
		Status:               e.Status.String(),
		Amount:               e.Amount,
		CurrencyCode:         e.CurrencyCode.String(),
		OriginalAmount:       pointer.Int64OrNil(e.OriginalAmount),
		OriginalCurrencyCode: pointer.StringOrNil(e.OriginalCurrencyCode.String()),
		FxRate:               pointer.Float64OrNil(e.FxRate),
		RowText:              e.RowText,
		PerformedAt:          pointer.TimeOrNil(e.PerformedAt),
		RejectedAt:           pointer.TimeOrNil(e.RejectedAt),
		CreatedAt:            e.CreatedAt,
	}

	if e.Category != nil {
		transactions.CategoryID = pointer.Int(e.Category.ID.Int())
	}

	if e.Subcategory != nil {
		transactions.SubcategoryID = pointer.Int(e.Subcategory.ID)
	}

	return transactions
}

func (r *transactionsRepo) ToEntity(ctx context.Context, m *Transactions) *entities.Transaction {
	if m == nil {
		return nil
	}

	id, _ := uuid.Parse(m.ID)
	userID, _ := uuid.Parse(m.UserID)
	accountID, _ := uuid.Parse(m.AccountID)

	e := &entities.Transaction{
		ID:                   id,
		UserID:               userID,
		AccountID:            accountID,
		Type:                 entities.TrnType(m.Type),
		Status:               entities.TrnStatus(m.Status),
		Amount:               m.Amount,
		CurrencyCode:         entities.Currency(m.CurrencyCode),
		OriginalAmount:       pointer.Int64Value(m.OriginalAmount),
		OriginalCurrencyCode: entities.Currency(pointer.StringValue(m.OriginalCurrencyCode)),
		FxRate:               pointer.Float64Value(m.FxRate),
		RowText:              m.RowText,
		PerformedAt:          pointer.TimeValue(m.PerformedAt),
		RejectedAt:           pointer.TimeValue(m.RejectedAt),
		CreatedAt:            m.CreatedAt,
	}

	if m.CategoryID != nil {
		category, err := r.categoriesRepo.FindByID(ctx, *m.CategoryID)
		if err == nil && category != nil {
			e.Category = category
		}
	}

	if m.SubcategoryID != nil {
		subcategory, err := r.subcategoriesRepo.FindByID(ctx, *m.SubcategoryID)
		if err == nil && subcategory != nil {
			e.Subcategory = subcategory
		}
	}

	return e
}
