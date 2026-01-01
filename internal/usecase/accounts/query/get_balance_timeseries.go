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

type GetBalanceTimeseriesUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	usersRepo      entities.UserRepository
	accountsRepo   entities.AccountRepository
	balanceLogRepo entities.AccountBalanceLogRepository
}

func NewGetBalanceTimeseriesUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	usersRepo entities.UserRepository,
	accountsRepo entities.AccountRepository,
	balanceLogRepo entities.AccountBalanceLogRepository,
) *GetBalanceTimeseriesUsecase {
	return &GetBalanceTimeseriesUsecase{
		contextTimeout: timeout,
		logger:         logger,
		usersRepo:      usersRepo,
		accountsRepo:   accountsRepo,
		balanceLogRepo: balanceLogRepo,
	}
}

type GetBalanceTimeseriesQuery struct {
	From       string   `form:"from"`
	To         string   `form:"to"`
	GroupBy    string   `form:"group_by"`
	AccountIDs []string `form:"account_ids"`
	Mode       string   `form:"mode"` // "aggregate" or "per_account"
}

type BalanceTimeseriesView struct {
	From       string                       `json:"from"`
	To         string                       `json:"to"`
	GroupBy    string                       `json:"group_by"`
	AccountIDs []string                     `json:"account_ids,omitempty"`
	Mode       string                       `json:"mode"`
	Series     []AccountBalanceSeriesView   `json:"series,omitempty"` // for per_account mode
	Points     []BalanceTimeseriesPointView `json:"points,omitempty"` // for aggregate mode
	Totals     BalanceTimeseriesTotalsView  `json:"totals"`
}

type AccountBalanceSeriesView struct {
	AccountID string                       `json:"account_id"`
	Points    []BalanceTimeseriesPointView `json:"points"`
	Totals    BalanceTimeseriesTotalsView  `json:"totals"`
}

type BalanceTimeseriesPointView struct {
	Timestamp    string   `json:"ts"`
	BalanceOpen  float64  `json:"balance_open"`
	BalanceClose float64  `json:"balance_close"`
	MinBalance   *float64 `json:"min_balance,omitempty"`
	MaxBalance   *float64 `json:"max_balance,omitempty"`
	Delta        float64  `json:"delta"`
	TxCount      int      `json:"tx_count"`
}

type BalanceTimeseriesTotalsView struct {
	StartBalance float64  `json:"start_balance"`
	EndBalance   float64  `json:"end_balance"`
	Change       float64  `json:"change"`
	MinBalance   *float64 `json:"min_balance,omitempty"`
	MaxBalance   *float64 `json:"max_balance,omitempty"`
	TxCount      int      `json:"tx_count"`
}

func (u *GetBalanceTimeseriesUsecase) GetBalanceTimeseries(
	ctx context.Context,
	userID string,
	query *GetBalanceTimeseriesQuery,
) (_ *BalanceTimeseriesView, err error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("accounts"), "GetBalanceTimeseries",
		attribute.String("user_id", userID),
		attribute.String("from", query.From),
		attribute.String("to", query.To),
		attribute.String("group_by", query.GroupBy),
		attribute.String("mode", query.Mode),
	)
	defer func() { end(err) }()

	var input struct {
		userID     uuid.UUID
		from       time.Time
		to         time.Time
		accountIDs []uuid.UUID
		groupBy    string
		mode       string
		timezone   string
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

		// Validate group_by
		if query.GroupBy == "" {
			query.GroupBy = "day" // default
		}
		if query.GroupBy != "day" && query.GroupBy != "week" && query.GroupBy != "month" {
			return nil, inerr.NewErrValidation("group_by", "invalid group_by value, must be day, week, or month")
		}
		input.groupBy = query.GroupBy

		// Validate mode
		if query.Mode == "" {
			query.Mode = "aggregate" // default
		}
		if query.Mode != "aggregate" && query.Mode != "per_account" {
			return nil, inerr.NewErrValidation("mode", "invalid mode value, must be aggregate or per_account")
		}
		input.mode = query.Mode
	}

	// Get user for currency and timezone
	user, err := u.usersRepo.FindByID(ctx, input.userID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get user", err)
		return nil, err
	}

	input.timezone = user.Timezone
	if input.timezone == "" {
		input.timezone = "UTC" // default
	}

	// Get all user accounts if no specific accounts provided
	if len(input.accountIDs) == 0 {
		accounts, err := u.accountsRepo.GetByUserID(ctx, input.userID)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to get user accounts", err)
			return nil, err
		}
		for _, acc := range accounts {
			input.accountIDs = append(input.accountIDs, acc.ID)
		}
	}

	// Call domain service to get balance timeseries
	filter := &entities.BalanceTimeseriesFilter{
		UserID:     input.userID,
		From:       input.from,
		To:         input.to,
		AccountIDs: input.accountIDs,
		GroupBy:    input.groupBy,
		Timezone:   input.timezone,
	}

	seriesMap, err := u.balanceLogRepo.GetBalanceTimeseries(ctx, filter)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get balance timeseries", err)
		return nil, err
	}

	scale := user.CurrencyCode.Scale()

	// Build response based on mode
	response := &BalanceTimeseriesView{
		From:    query.From,
		To:      query.To,
		GroupBy: input.groupBy,
		Mode:    input.mode,
	}

	if len(input.accountIDs) > 0 {
		accountIDStrs := make([]string, 0, len(input.accountIDs))
		for _, id := range input.accountIDs {
			accountIDStrs = append(accountIDStrs, id.String())
		}
		response.AccountIDs = accountIDStrs
	}

	if input.mode == "per_account" {
		// Return separate series for each account
		response.Series = make([]AccountBalanceSeriesView, 0, len(seriesMap))

		for accountID, points := range seriesMap {
			series := AccountBalanceSeriesView{
				AccountID: accountID.String(),
				Points:    convertPoints(points, scale),
			}
			series.Totals = calculateTotals(series.Points)
			response.Series = append(response.Series, series)
		}

		// Calculate overall totals
		response.Totals = aggregateTotals(response.Series)
	} else {
		// Aggregate mode: combine all accounts into a single timeseries
		aggregatedPoints := aggregatePoints(seriesMap, scale)
		response.Points = aggregatedPoints
		response.Totals = calculateTotals(aggregatedPoints)
	}

	return response, nil
}

func convertPoints(points []entities.BalanceTimeseriesPoint, scale int) []BalanceTimeseriesPointView {
	result := make([]BalanceTimeseriesPointView, 0, len(points))
	for _, p := range points {
		minBal := entities.MajorFromMinor(p.MinBalance, scale)
		maxBal := entities.MajorFromMinor(p.MaxBalance, scale)

		result = append(result, BalanceTimeseriesPointView{
			Timestamp:    p.Timestamp,
			BalanceOpen:  entities.MajorFromMinor(p.BalanceOpen, scale),
			BalanceClose: entities.MajorFromMinor(p.BalanceClose, scale),
			MinBalance:   &minBal,
			MaxBalance:   &maxBal,
			Delta:        entities.MajorFromMinor(p.Delta, scale),
			TxCount:      p.TxCount,
		})
	}
	return result
}

func calculateTotals(points []BalanceTimeseriesPointView) BalanceTimeseriesTotalsView {
	if len(points) == 0 {
		return BalanceTimeseriesTotalsView{}
	}

	totals := BalanceTimeseriesTotalsView{
		StartBalance: points[0].BalanceOpen,
		EndBalance:   points[len(points)-1].BalanceClose,
	}
	totals.Change = totals.EndBalance - totals.StartBalance

	var minBal, maxBal float64
	hasMinMax := false

	for _, p := range points {
		totals.TxCount += p.TxCount
		if p.MinBalance != nil {
			if !hasMinMax || *p.MinBalance < minBal {
				minBal = *p.MinBalance
				hasMinMax = true
			}
		}
		if p.MaxBalance != nil {
			if !hasMinMax || *p.MaxBalance > maxBal {
				maxBal = *p.MaxBalance
			}
		}
	}

	if hasMinMax {
		totals.MinBalance = &minBal
		totals.MaxBalance = &maxBal
	}

	return totals
}

func aggregatePoints(seriesMap map[uuid.UUID][]entities.BalanceTimeseriesPoint, scale int) []BalanceTimeseriesPointView {
	if len(seriesMap) == 0 {
		return []BalanceTimeseriesPointView{}
	}

	// Build a map of timestamp -> aggregated values
	bucketMap := make(map[string]*BalanceTimeseriesPointView)

	for _, points := range seriesMap {
		for _, p := range points {
			if existing, exists := bucketMap[p.Timestamp]; exists {
				// Aggregate values
				existing.BalanceOpen += entities.MajorFromMinor(p.BalanceOpen, scale)
				existing.BalanceClose += entities.MajorFromMinor(p.BalanceClose, scale)

				if existing.MinBalance != nil && p.MinBalance > 0 {
					minVal := entities.MajorFromMinor(p.MinBalance, scale)
					*existing.MinBalance += minVal
				}
				if existing.MaxBalance != nil && p.MaxBalance > 0 {
					maxVal := entities.MajorFromMinor(p.MaxBalance, scale)
					*existing.MaxBalance += maxVal
				}

				existing.Delta += entities.MajorFromMinor(p.Delta, scale)
				existing.TxCount += p.TxCount
			} else {
				// Create new entry
				minBal := entities.MajorFromMinor(p.MinBalance, scale)
				maxBal := entities.MajorFromMinor(p.MaxBalance, scale)

				bucketMap[p.Timestamp] = &BalanceTimeseriesPointView{
					Timestamp:    p.Timestamp,
					BalanceOpen:  entities.MajorFromMinor(p.BalanceOpen, scale),
					BalanceClose: entities.MajorFromMinor(p.BalanceClose, scale),
					MinBalance:   &minBal,
					MaxBalance:   &maxBal,
					Delta:        entities.MajorFromMinor(p.Delta, scale),
					TxCount:      p.TxCount,
				}
			}
		}
	}

	// Convert map to sorted slice
	result := make([]BalanceTimeseriesPointView, 0, len(bucketMap))
	for _, point := range bucketMap {
		result = append(result, *point)
	}

	// Sort by timestamp
	// Note: timestamps are ISO formatted strings, so lexicographic sort works
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Timestamp > result[j].Timestamp {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

func aggregateTotals(series []AccountBalanceSeriesView) BalanceTimeseriesTotalsView {
	if len(series) == 0 {
		return BalanceTimeseriesTotalsView{}
	}

	totals := BalanceTimeseriesTotalsView{}

	for _, s := range series {
		totals.StartBalance += s.Totals.StartBalance
		totals.EndBalance += s.Totals.EndBalance
		totals.TxCount += s.Totals.TxCount

		if s.Totals.MinBalance != nil {
			if totals.MinBalance == nil {
				val := *s.Totals.MinBalance
				totals.MinBalance = &val
			} else {
				*totals.MinBalance += *s.Totals.MinBalance
			}
		}

		if s.Totals.MaxBalance != nil {
			if totals.MaxBalance == nil {
				val := *s.Totals.MaxBalance
				totals.MaxBalance = &val
			} else {
				*totals.MaxBalance += *s.Totals.MaxBalance
			}
		}
	}

	totals.Change = totals.EndBalance - totals.StartBalance

	return totals
}
