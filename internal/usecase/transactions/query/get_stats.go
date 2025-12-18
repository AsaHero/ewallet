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
	TotalIncome       float64        `json:"total_income"`
	TotalExpense      float64        `json:"total_expense"`
	Balance           float64        `json:"balance"`
	IncomeByCategory  []CategoryStat `json:"income_by_category"`
	ExpenseByCategory []CategoryStat `json:"expense_by_category"`
}

type CategoryStat struct {
	CategoryID    int     `json:"category_id"`
	CategoryName  string  `json:"category_name"`
	CategoryEmoji string  `json:"category_emoji"`
	Total         float64 `json:"total"`
}

func (u *GetStatsUsecase) GetStats(ctx context.Context, userID string, accountID string, from string, to string) (_ *GetStatsView, err error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("transactions"), "GetStats",
		attribute.String("user_id", userID),
		attribute.String("account_id", accountID),
		attribute.String("from", from),
		attribute.String("to", to),
	)
	defer func() { end(err) }()

	var input struct {
		userID    uuid.UUID
		accountID *uuid.UUID
		from      *time.Time
		to        *time.Time
	}
	{
		var err error
		input.userID, err = uuid.Parse(userID)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to parse user id", err)
			return nil, inerr.NewErrValidation("user_id", "invalud uuid type")
		}

		if accountID != "" {
			accountUUID, err := uuid.Parse(accountID)
			if err != nil {
				u.logger.ErrorContext(ctx, "failed to parse account id", err)
				return nil, inerr.NewErrValidation("account_id", "invalid uuid type")
			}
			input.accountID = &accountUUID
		}

		if from != "" {
			from, err := time.Parse(time.DateOnly, from)
			if err != nil {
				u.logger.ErrorContext(ctx, "failed to parse from", err)
				return nil, inerr.NewErrValidation("from", "invalud date format")
			}
			input.from = &from
		}

		if to != "" {
			to, err := time.Parse(time.DateOnly, to)
			if err != nil {
				u.logger.ErrorContext(ctx, "failed to parse to", err)
				return nil, inerr.NewErrValidation("to", "invalud date format")
			}
			to = utils.EndOfDate(to)
			input.to = &to
		}
	}

	user, err := u.usersRepo.FindByID(ctx, input.userID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get user", err)
		return nil, err
	}

	totalIncome, err := u.transactionsRepo.GetTotalByTypeAndAccount(ctx, user.ID, input.accountID, entities.Deposit, input.from, input.to)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get total income", err)
		return nil, err
	}

	totalExpense, err := u.transactionsRepo.GetTotalByTypeAndAccount(ctx, user.ID, input.accountID, entities.Withdrawal, input.from, input.to)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get total expense", err)
		return nil, err
	}

	balance, err := u.accountsRepo.GetTotalBalance(ctx, user.ID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get balance", err)
		return nil, err
	}

	incomeByCategory, incomeCategories, err := u.transactionsRepo.GetTotalsByCategoriesAndAccount(ctx, user.ID, input.accountID, entities.Deposit, input.from, input.to)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get stats by category", err)
		return nil, err
	}

	expenseByCategory, expenseCategories, err := u.transactionsRepo.GetTotalsByCategoriesAndAccount(ctx, user.ID, input.accountID, entities.Withdrawal, input.from, input.to)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get stats by category", err)
		return nil, err
	}

	response := &GetStatsView{
		TotalIncome:  entities.MajorFromMinor(totalIncome, user.CurrencyCode.Scale()),
		TotalExpense: entities.MajorFromMinor(totalExpense, user.CurrencyCode.Scale()),
		Balance:      entities.MajorFromMinor(balance, user.CurrencyCode.Scale()),
	}

	for _, categoryID := range incomeCategories {
		category, err := u.categoriesRepo.FindByID(ctx, categoryID)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to get catzegory", err)
			return nil, err
		}

		total, ok := incomeByCategory[categoryID]
		if !ok {
			continue
		}

		response.IncomeByCategory = append(response.IncomeByCategory, CategoryStat{
			CategoryID:    category.ID.Int(),
			CategoryName:  category.GetName(user.LanguageCode),
			CategoryEmoji: category.Emoji,
			Total:         entities.MajorFromMinor(total, user.CurrencyCode.Scale()),
		})
	}

	for _, categoryID := range expenseCategories {
		category, err := u.categoriesRepo.FindByID(ctx, categoryID)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to get catzegory", err)
			return nil, err
		}

		total, ok := expenseByCategory[categoryID]
		if !ok {
			continue
		}

		response.ExpenseByCategory = append(response.ExpenseByCategory, CategoryStat{
			CategoryID:    category.ID.Int(),
			CategoryName:  category.GetName(user.LanguageCode),
			CategoryEmoji: category.Emoji,
			Total:         entities.MajorFromMinor(total, user.CurrencyCode.Scale()),
		})
	}

	return response, nil
}
