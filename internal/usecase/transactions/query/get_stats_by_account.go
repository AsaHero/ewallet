package query

import (
	"context"
	"math"
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

type GetStatsByAccountUsecase struct {
	contextTimeout   time.Duration
	logger           *logger.Logger
	usersRepo        entities.UserRepository
	transactionsRepo entities.TransactionRepository
	accountsRepo     entities.AccountRepository
}

func NewGetStatsByAccountUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	usersRepo entities.UserRepository,
	transactionsRepo entities.TransactionRepository,
	accountsRepo entities.AccountRepository,
) *GetStatsByAccountUsecase {
	return &GetStatsByAccountUsecase{
		contextTimeout:   timeout,
		transactionsRepo: transactionsRepo,
		usersRepo:        usersRepo,
		accountsRepo:     accountsRepo,
		logger:           logger,
	}
}

type AccountStatsView struct {
	From   string             `json:"from"`
	To     string             `json:"to"`
	Items  []AccountStatsItem `json:"items"`
	Totals AccountStatsTotals `json:"totals"`
}

type AccountStatsItem struct {
	AccountID string  `json:"account_id"`
	Name      string  `json:"name"`
	Income    float64 `json:"income"`
	Expense   float64 `json:"expense"`
	Net       float64 `json:"net"`
	Count     int     `json:"count"`
}

type AccountStatsTotals struct {
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Net     float64 `json:"net"`
	Count   int     `json:"count"`
}

func (u *GetStatsByAccountUsecase) GetStatsByAccount(
	ctx context.Context,
	userID string,
	from string,
	to string,
	trnType string,
) (_ *AccountStatsView, err error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("transactions"), "GetStatsByAccount",
		attribute.String("user_id", userID),
		attribute.String("from", from),
		attribute.String("to", to),
	)
	defer func() { end(err) }()

	var input struct {
		userID  uuid.UUID
		from    time.Time
		to      time.Time
		trnType *entities.TrnType
	}

	// Parse and validate inputs
	{
		var err error
		input.userID, err = uuid.Parse(userID)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to parse user id", err)
			return nil, inerr.NewErrValidation("user_id", "invalid uuid type")
		}

		if from == "" {
			return nil, inerr.NewErrValidation("from", "from date is required")
		}
		input.from, err = time.Parse(time.DateOnly, from)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to parse from", err)
			return nil, inerr.NewErrValidation("from", "invalid date format, use YYYY-MM-DD")
		}

		if to == "" {
			return nil, inerr.NewErrValidation("to", "to date is required")
		}
		input.to, err = time.Parse(time.DateOnly, to)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to parse to", err)
			return nil, inerr.NewErrValidation("to", "invalid date format, use YYYY-MM-DD")
		}
		input.to = utils.EndOfDate(input.to)

		// Parse transaction type
		if trnType != "" {
			t := entities.TrnType(trnType)
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

	// Call repository to get account stats
	items, err := u.transactionsRepo.GetStatsByAccount(ctx, input.userID, input.from, input.to, input.trnType)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get account stats", err)
		return nil, err
	}

	// Build response
	response := &AccountStatsView{
		From:  from,
		To:    to,
		Items: make([]AccountStatsItem, 0, len(items)),
	}

	scale := user.CurrencyCode.Scale()
	var totals AccountStatsTotals

	for _, item := range items {
		account, err := u.accountsRepo.GetByID(ctx, item.AccountID)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to get account", err)
			continue
		}

		// Expenses come as negative, convert to absolute values for display
		incomeMajor := entities.MajorFromMinor(item.Income, scale)
		expenseMajor := math.Abs(entities.MajorFromMinor(item.Expense, scale))
		netMajor := entities.MajorFromMinor(item.Net, scale)

		totals.Income += incomeMajor
		totals.Expense += expenseMajor
		totals.Net += netMajor
		totals.Count += item.Count

		response.Items = append(response.Items, AccountStatsItem{
			AccountID: item.AccountID.String(),
			Name:      account.Name,
			Income:    incomeMajor,
			Expense:   expenseMajor,
			Net:       netMajor,
			Count:     item.Count,
		})
	}

	response.Totals = totals

	return response, nil
}
