package query

import (
	"context"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/inerr"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
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
	TotalIncome  float64
	TotalExpense float64
	Balance      float64
	ByCategory   []CategoryStat
}

type CategoryStat struct {
	CategoryID int
	Total      float64
}

func (u *GetStatsUsecase) GetStats(ctx context.Context, userID string) (_ *GetStatsView, err error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("transactions"), "GetStats",
		attribute.String("user_id", userID),
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

	totalIncome, err := u.transactionsRepo.GetTotalByType(ctx, user.ID, entities.Deposit)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get total income", err)
		return nil, err
	}

	totalExpense, err := u.transactionsRepo.GetTotalByType(ctx, user.ID, entities.Withdrawal)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get total expense", err)
		return nil, err
	}

	balance, err := u.accountsRepo.GetTotalBalance(ctx, user.ID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get balance", err)
		return nil, err
	}

	byCategory, err := u.transactionsRepo.GetTotalsByCategories(ctx, user.ID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get stats by category", err)
		return nil, err
	}

	response := &GetStatsView{
		TotalIncome:  entities.MajorFromMinor(totalIncome, user.CurrencyCode.Scale()),
		TotalExpense: entities.MajorFromMinor(totalExpense, user.CurrencyCode.Scale()),
		Balance:      entities.MajorFromMinor(balance, user.CurrencyCode.Scale()),
	}

	for categoryID, total := range byCategory {
		response.ByCategory = append(response.ByCategory, CategoryStat{
			CategoryID: categoryID,
			Total:      entities.MajorFromMinor(total, user.CurrencyCode.Scale()),
		})
	}

	return response, nil
}
