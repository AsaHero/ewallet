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

type GetStatsByCategoryUsecase struct {
	contextTimeout   time.Duration
	logger           *logger.Logger
	usersRepo        entities.UserRepository
	transactionsRepo entities.TransactionRepository
	categoriesRepo   entities.CategoryRepository
}

func NewGetStatsByCategoryUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	usersRepo entities.UserRepository,
	transactionsRepo entities.TransactionRepository,
	categoriesRepo entities.CategoryRepository,
) *GetStatsByCategoryUsecase {
	return &GetStatsByCategoryUsecase{
		contextTimeout:   timeout,
		transactionsRepo: transactionsRepo,
		usersRepo:        usersRepo,
		categoriesRepo:   categoriesRepo,
		logger:           logger,
	}
}

type GetStatsByCategoryQuery struct {
	From       string   `form:"from" validate:"required"`
	To         string   `form:"to" validate:"required"`
	Type       string   `form:"type"`
	AccountIDs []string `form:"account_ids"`
}

type CategoryStatsView struct {
	From   string              `json:"from"`
	To     string              `json:"to"`
	Type   string              `json:"type,omitempty"`
	Items  []CategoryStatsItem `json:"items"`
	Totals CategoryStatsTotals `json:"totals"`
}

type CategoryStatsItem struct {
	CategoryID int     `json:"category_id"`
	Name       string  `json:"name"`
	Emoji      string  `json:"emoji"`
	Total      float64 `json:"total"`
	Count      int     `json:"count"`
	Share      float64 `json:"share"`
}

type CategoryStatsTotals struct {
	Total float64 `json:"total"`
	Count int     `json:"count"`
}

func (u *GetStatsByCategoryUsecase) GetStatsByCategory(
	ctx context.Context,
	userID string,
	query *GetStatsByCategoryQuery,
) (_ *CategoryStatsView, err error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("transactions"), "GetStatsByCategory",
		attribute.String("user_id", userID),
		attribute.String("from", query.From),
		attribute.String("to", query.To),
	)
	defer func() { end(err) }()

	var input struct {
		userID     uuid.UUID
		from       time.Time
		to         time.Time
		accountIDs []uuid.UUID
		trnType    *entities.TrnType
	}

	// Parse and validate inputs
	{
		var err error
		input.userID, err = uuid.Parse(userID)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to parse user id", err)
			return nil, inerr.NewErrValidation("user_id", "invalid uuid type")
		}

		if query.From == "" {
			return nil, inerr.NewErrValidation("from", "from date is required")
		}
		input.from, err = time.Parse(time.DateOnly, query.From)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to parse from", err)
			return nil, inerr.NewErrValidation("from", "invalid date format, use YYYY-MM-DD")
		}

		if query.To == "" {
			return nil, inerr.NewErrValidation("to", "to date is required")
		}
		input.to, err = time.Parse(time.DateOnly, query.To)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to parse to", err)
			return nil, inerr.NewErrValidation("to", "invalid date format, use YYYY-MM-DD")
		}
		input.to = utils.EndOfDate(input.to)

		// Parse account IDs
		if len(query.AccountIDs) > 0 {
			input.accountIDs = make([]uuid.UUID, 0, len(query.AccountIDs))
			for _, idStr := range query.AccountIDs {
				accountID, err := uuid.Parse(idStr)
				if err != nil {
					u.logger.ErrorContext(ctx, "failed to parse account id", err)
					return nil, inerr.NewErrValidation("account_ids", "invalid uuid in account_ids")
				}
				input.accountIDs = append(input.accountIDs, accountID)
			}
		}

		// Parse transaction type
		if query.Type != "" {
			t := entities.TrnType(query.Type)
			if t != entities.Deposit && t != entities.Withdrawal && t != entities.Transfer && t != entities.Adjustment {
				return nil, inerr.NewErrValidation("type", "invalid transaction type")
			}
			input.trnType = &t
		}
	}

	// Get user for currency conversion
	user, err := u.usersRepo.FindByID(ctx, input.userID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get user", err)
		return nil, err
	}

	// Call repository to get category stats
	items, err := u.transactionsRepo.GetStatsByCategory(ctx, input.userID, input.from, input.to, input.accountIDs, input.trnType)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get category stats", err)
		return nil, err
	}

	// Build response
	response := &CategoryStatsView{
		From:  query.From,
		To:    query.To,
		Items: make([]CategoryStatsItem, 0, len(items)),
	}

	if query.Type != "" {
		response.Type = query.Type
	}

	scale := user.CurrencyCode.Scale()
	var totals CategoryStatsTotals

	for _, item := range items {
		category, err := u.categoriesRepo.FindByID(ctx, item.CategoryID)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to get category", err)
			continue
		}

		totalMajor := entities.MajorFromMinor(item.Total, scale)
		totals.Total += totalMajor
		totals.Count += item.Count

		response.Items = append(response.Items, CategoryStatsItem{
			CategoryID: item.CategoryID,
			Name:       category.GetName(user.LanguageCode),
			Emoji:      category.Emoji,
			Total:      totalMajor,
			Count:      item.Count,
			Share:      0, // Will calculate after we know the total
		})
	}

	// Calculate share percentages
	if totals.Total > 0 {
		for i := range response.Items {
			response.Items[i].Share = response.Items[i].Total / totals.Total
		}
	}

	response.Totals = totals

	return response, nil
}
