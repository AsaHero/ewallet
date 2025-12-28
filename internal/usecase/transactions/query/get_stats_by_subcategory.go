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

type GetStatsBySubcategoryUsecase struct {
	contextTimeout    time.Duration
	logger            *logger.Logger
	usersRepo         entities.UserRepository
	transactionsRepo  entities.TransactionRepository
	subcategoriesRepo entities.SubcategoryRepository
}

func NewGetStatsBySubcategoryUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	usersRepo entities.UserRepository,
	transactionsRepo entities.TransactionRepository,
	subcategoriesRepo entities.SubcategoryRepository,
) *GetStatsBySubcategoryUsecase {
	return &GetStatsBySubcategoryUsecase{
		contextTimeout:    timeout,
		transactionsRepo:  transactionsRepo,
		usersRepo:         usersRepo,
		subcategoriesRepo: subcategoriesRepo,
		logger:            logger,
	}
}

type GetStatsBySubcategoryQuery struct {
	From        string   `form:"from" binding:"required"`
	To          string   `form:"to" binding:"required"`
	AccountIDs  []string `form:"account_ids"`
	CategoryIDs []int    `form:"category_ids"`
	Type        string   `form:"type"`
}

type SubcategoryStatsView struct {
	From   string                 `json:"from"`
	To     string                 `json:"to"`
	Type   string                 `json:"type,omitempty"`
	Items  []SubcategoryStatsItem `json:"items"`
	Totals SubcategoryStatsTotals `json:"totals"`
}

type SubcategoryStatsItem struct {
	SubcategoryID int     `json:"subcategory_id"`
	CategoryID    int     `json:"category_id"`
	Name          string  `json:"name"`
	Emoji         string  `json:"emoji"`
	Total         float64 `json:"total"`
	Count         int     `json:"count"`
	Share         float64 `json:"share"`
}

type SubcategoryStatsTotals struct {
	Total float64 `json:"total"`
	Count int     `json:"count"`
}

func (u *GetStatsBySubcategoryUsecase) GetStatsBySubcategory(
	ctx context.Context,
	userID string,
	query *GetStatsBySubcategoryQuery,
) (_ *SubcategoryStatsView, err error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("transactions"), "GetStatsBySubcategory",
		attribute.String("user_id", userID),
		attribute.String("from", query.From),
		attribute.String("to", query.To),
	)
	defer func() { end(err) }()

	var input struct {
		userID      uuid.UUID
		from        time.Time
		to          time.Time
		accountIDs  []uuid.UUID
		categoryIDs []int
		trnType     *entities.TrnType
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

		// Parse category IDs
		input.categoryIDs = query.CategoryIDs

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

	// Call repository to get subcategory stats
	items, err := u.transactionsRepo.GetStatsBySubcategory(ctx, input.userID, input.from, input.to, input.accountIDs, input.categoryIDs, input.trnType)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get subcategory stats", err)
		return nil, err
	}

	// Build response
	response := &SubcategoryStatsView{
		From:  query.From,
		To:    query.To,
		Items: make([]SubcategoryStatsItem, 0, len(items)),
	}

	if query.Type != "" {
		response.Type = query.Type
	}

	scale := user.CurrencyCode.Scale()
	var totals SubcategoryStatsTotals

	for _, item := range items {
		subcategory, err := u.subcategoriesRepo.FindByID(ctx, item.SubcategoryID)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to get subcategory", err)
			continue
		}

		totalMajor := entities.MajorFromMinor(item.Total, scale)
		totals.Total += totalMajor
		totals.Count += item.Count

		response.Items = append(response.Items, SubcategoryStatsItem{
			SubcategoryID: item.SubcategoryID,
			CategoryID:    item.CategoryID,
			Name:          subcategory.GetName(user.LanguageCode),
			Emoji:         subcategory.Emoji,
			Total:         totalMajor,
			Count:         item.Count,
			Share:         0, // Will calculate after we know the total
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
