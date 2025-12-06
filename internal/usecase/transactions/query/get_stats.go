package query

import (
	"context"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/inerr"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/AsaHero/e-wallet/pkg/utils"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type GetStatsUsecase struct {
	contextTimeout   time.Duration
	logger           *logger.Logger
	usersRepo        entities.UserRepository
	accountsRepo     entities.AccountRepository
	transactionsRepo entities.TransactionRepository
	categoriesRepo   entities.CategoryRepository
}

func NewGetStatsUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	usersRepo entities.UserRepository,
	accountsRepo entities.AccountRepository,
	transactionsRepo entities.TransactionRepository,
	categoriesRepo entities.CategoryRepository,
) *GetStatsUsecase {
	return &GetStatsUsecase{
		contextTimeout:   timeout,
		transactionsRepo: transactionsRepo,
		usersRepo:        usersRepo,
		accountsRepo:     accountsRepo,
		categoriesRepo:   categoriesRepo,
		logger:           logger,
	}
}

type GetStatsView struct {
	TotalIncome  float64        `json:"total_income"`
	TotalExpense float64        `json:"total_expense"`
	Balance      float64        `json:"balance"`
	ByCategory   []CategoryStat `json:"by_category"`
}

type CategoryStat struct {
	CategoryID   int     `json:"category_id"`
	CategoryName string  `json:"category_name"`
	CategorySlug string  `json:"category_slug"`
	Total        float64 `json:"total"`
}

func (u *GetStatsUsecase) GetStats(ctx context.Context, userID string, period string) (_ *GetStatsView, err error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("transactions"), "GetStats",
		attribute.String("user_id", userID),
		attribute.String("period", period),
	)
	defer func() { end(err) }()

	var input struct {
		userID uuid.UUID
	}
	{
		var err error
		input.userID, err = uuid.Parse(userID)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to parse user id", err)
			return nil, inerr.NewErrValidation("user_id", "invalud uuid type")
		}
	}

	user, err := u.usersRepo.FindByID(ctx, input.userID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get user", err)
		return nil, err
	}

	var from, to *time.Time
	now := time.Now().UTC()

	from = utils.GetStartDateByPeriod(period, now)

	totalIncome, err := u.transactionsRepo.GetTotalByType(ctx, user.ID, entities.Deposit, from, to)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get total income", err)
		return nil, err
	}

	totalExpense, err := u.transactionsRepo.GetTotalByType(ctx, user.ID, entities.Withdrawal, from, to)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get total expense", err)
		return nil, err
	}

	balance, err := u.accountsRepo.GetTotalBalance(ctx, user.ID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get balance", err)
		return nil, err
	}

	byCategory, categories, err := u.transactionsRepo.GetTotalsByCategories(ctx, user.ID, from, to)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get stats by category", err)
		return nil, err
	}

	response := &GetStatsView{
		TotalIncome:  entities.MajorFromMinor(totalIncome, user.CurrencyCode.Scale()),
		TotalExpense: entities.MajorFromMinor(totalExpense, user.CurrencyCode.Scale()),
		Balance:      entities.MajorFromMinor(balance, user.CurrencyCode.Scale()),
	}

	for _, categoryID := range categories {
		category, err := u.categoriesRepo.FindByID(ctx, categoryID)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to get catzegory", err)
			return nil, err
		}

		total, ok := byCategory[categoryID]
		if !ok {
			continue
		}

		response.ByCategory = append(response.ByCategory, CategoryStat{
			CategoryID:   category.ID,
			CategoryName: category.Name,
			CategorySlug: category.Slug,
			Total:        entities.MajorFromMinor(total, user.CurrencyCode.Scale()),
		})
	}

	return response, nil
}
