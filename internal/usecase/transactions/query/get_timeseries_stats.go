package query

import (
	"context"
	"strconv"
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

type GetTimeseriesStatsUsecase struct {
	contextTimeout   time.Duration
	logger           *logger.Logger
	usersRepo        entities.UserRepository
	transactionsRepo entities.TransactionRepository
}

func NewGetTimeseriesStatsUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	usersRepo entities.UserRepository,
	transactionsRepo entities.TransactionRepository,
) *GetTimeseriesStatsUsecase {
	return &GetTimeseriesStatsUsecase{
		contextTimeout:   timeout,
		transactionsRepo: transactionsRepo,
		usersRepo:        usersRepo,
		logger:           logger,
	}
}

type GetTimeseriesStatsQuery struct {
	From           string   `form:"from"`
	To             string   `form:"to"`
	AccountIDs     []string `form:"account_ids"`
	CategoryIDs    []string `form:"category_ids"`
	SubcategoryIDs []string `form:"subcategory_ids"`
	Type           *string  `form:"type"`
	GroupBy        string   `form:"group_by"`
}

type TimeseriesStatsView struct {
	GroupBy string                `json:"group_by"`
	From    string                `json:"from"`
	To      string                `json:"to"`
	Points  []TimeseriesDataPoint `json:"points"`
	Totals  TimeseriesTotals      `json:"totals"`
}

type TimeseriesDataPoint struct {
	Timestamp string  `json:"ts"`
	Income    float64 `json:"income"`
	Expense   float64 `json:"expense"`
	Net       float64 `json:"net"`
	Count     int     `json:"count"`
}

type TimeseriesTotals struct {
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Net     float64 `json:"net"`
	Count   int     `json:"count"`
}

func (u *GetTimeseriesStatsUsecase) GetTimeseriesStats(
	ctx context.Context,
	userID string,
	query *GetTimeseriesStatsQuery,
) (_ *TimeseriesStatsView, err error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("transactions"), "GetTimeseriesStats",
		attribute.String("user_id", userID),
		attribute.String("from", query.From),
		attribute.String("to", query.To),
		attribute.String("group_by", query.GroupBy),
	)
	defer func() { end(err) }()

	var input struct {
		userID         uuid.UUID
		from           time.Time
		to             time.Time
		accountIDs     []uuid.UUID
		categoryIDs    []int
		subcategoryIDs []int
		trnType        *entities.TrnType
		groupBy        string
	}

	// Parse and validate inputs
	{
		var err error
		input.userID, err = uuid.Parse(userID)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to parse user id", err)
			return nil, inerr.NewErrValidation("user_id", "invalid uuid type")
		}

		// Parse and validate from date (required)
		if query.From == "" {
			return nil, inerr.NewErrValidation("from", "from date is required")
		}
		input.from, err = time.Parse(time.DateOnly, query.From)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to parse from", err)
			return nil, inerr.NewErrValidation("from", "invalid date format, use YYYY-MM-DD")
		}

		// Parse and validate to date (required)
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

		// Parse category IDs (already ints, just assign)
		if len(query.CategoryIDs) > 0 {
			input.categoryIDs = make([]int, 0, len(query.CategoryIDs))
			for _, idStr := range query.CategoryIDs {
				categoryID, err := strconv.Atoi(idStr)
				if err != nil {
					u.logger.ErrorContext(ctx, "failed to parse category id", err)
					return nil, inerr.NewErrValidation("category_ids", "invalid int in category_ids")
				}
				input.categoryIDs = append(input.categoryIDs, categoryID)
			}
		}
		if len(query.SubcategoryIDs) > 0 {
			input.subcategoryIDs = make([]int, 0, len(query.SubcategoryIDs))
			for _, idStr := range query.SubcategoryIDs {
				subcategoryID, err := strconv.Atoi(idStr)
				if err != nil {
					u.logger.ErrorContext(ctx, "failed to parse subcategory id", err)
					return nil, inerr.NewErrValidation("subcategory_ids", "invalid int in subcategory_ids")
				}
				input.subcategoryIDs = append(input.subcategoryIDs, subcategoryID)
			}
		}

		// Parse transaction type
		if query.Type != nil {
			t := entities.TrnType(*query.Type)
			if t != entities.Deposit && t != entities.Withdrawal && t != entities.Transfer && t != entities.Adjustment {
				return nil, inerr.NewErrValidation("type", "invalid transaction type, must be deposit, withdrawal, transfer, or adjustment")
			}
			input.trnType = &t
		}

		// Validate group_by
		if query.GroupBy == "" {
			query.GroupBy = "day" // default
		}
		if query.GroupBy != "day" && query.GroupBy != "week" && query.GroupBy != "month" {
			return nil, inerr.NewErrValidation("group_by", "invalid group_by value, must be day, week, or month")
		}
		input.groupBy = query.GroupBy
	}

	// Get user for currency conversion
	user, err := u.usersRepo.FindByID(ctx, input.userID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get user", err)
		return nil, err
	}

	// Call repository to get timeseries data
	filter := &entities.TimeseriesFilter{
		UserID:         input.userID,
		From:           input.from,
		To:             input.to,
		AccountIDs:     input.accountIDs,
		CategoryIDs:    input.categoryIDs,
		SubcategoryIDs: input.subcategoryIDs,
		Type:           input.trnType,
		GroupBy:        input.groupBy,
	}

	points, err := u.transactionsRepo.GetTimeseriesStats(ctx, filter)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get timeseries stats", err)
		return nil, err
	}

	// Convert to response format and calculate totals
	response := &TimeseriesStatsView{
		GroupBy: input.groupBy,
		From:    query.From,
		To:      query.To,
		Points:  make([]TimeseriesDataPoint, 0, len(points)),
	}

	var totals TimeseriesTotals
	scale := user.CurrencyCode.Scale()

	for _, point := range points {
		dataPoint := TimeseriesDataPoint{
			Timestamp: point.Timestamp,
			Income:    entities.MajorFromMinor(point.Income, scale),
			Expense:   entities.MajorFromMinor(point.Expense, scale),
			Net:       entities.MajorFromMinor(point.Net, scale),
			Count:     point.Count,
		}
		response.Points = append(response.Points, dataPoint)

		// Accumulate totals
		totals.Income += dataPoint.Income
		totals.Expense += dataPoint.Expense
		totals.Net += dataPoint.Net
		totals.Count += dataPoint.Count
	}

	response.Totals = totals

	return response, nil
}
