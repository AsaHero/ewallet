package query

import (
	"context"
	"math"
	"sort"
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

type GetStatsCompareUsecase struct {
	contextTimeout   time.Duration
	logger           *logger.Logger
	usersRepo        entities.UserRepository
	transactionsRepo entities.TransactionRepository
	categoriesRepo   entities.CategoryRepository
}

func NewGetStatsCompareUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	usersRepo entities.UserRepository,
	transactionsRepo entities.TransactionRepository,
	categoriesRepo entities.CategoryRepository,
) *GetStatsCompareUsecase {
	return &GetStatsCompareUsecase{
		contextTimeout:   timeout,
		transactionsRepo: transactionsRepo,
		usersRepo:        usersRepo,
		categoriesRepo:   categoriesRepo,
		logger:           logger,
	}
}

type GetStatsCompareQuery struct {
	Period      string   `form:"period"`
	BaseFrom    string   `form:"base_from"`
	BaseTo      string   `form:"base_to"`
	CompareFrom string   `form:"compare_from"`
	CompareTo   string   `form:"compare_to"`
	AccountIDs  []string `form:"account_ids"`
	Type        string   `form:"type"`
	TopLimit    int      `form:"top_limit"`
}

type StatsCompareView struct {
	Base       PeriodStats `json:"base"`
	Compare    PeriodStats `json:"compare"`
	Delta      DeltaStats  `json:"delta"`
	TopChanges TopChanges  `json:"top_changes"`
}

type PeriodStats struct {
	From    string  `json:"from"`
	To      string  `json:"to"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Net     float64 `json:"net"`
	Count   int     `json:"count"`
}

type DeltaStats struct {
	Income  DeltaValue `json:"income"`
	Expense DeltaValue `json:"expense"`
	Net     DeltaValue `json:"net"`
}

type DeltaValue struct {
	Abs float64 `json:"abs"`
	Pct float64 `json:"pct"`
}

type TopChanges struct {
	ExpenseByCategory []CategoryChange `json:"expense_by_category"`
}

type CategoryChange struct {
	CategoryID int     `json:"category_id"`
	Name       string  `json:"name"`
	Emoji      string  `json:"emoji"`
	Base       float64 `json:"base"`
	Compare    float64 `json:"compare"`
	DeltaAbs   float64 `json:"delta_abs"`
	DeltaPct   float64 `json:"delta_pct"`
}

func (u *GetStatsCompareUsecase) GetStatsCompare(
	ctx context.Context,
	userID string,
	query *GetStatsCompareQuery,
) (_ *StatsCompareView, err error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("transactions"), "GetStatsCompare",
		attribute.String("user_id", userID),
		attribute.String("period", query.Period),
	)
	defer func() { end(err) }()

	var input struct {
		userID      uuid.UUID
		baseFrom    time.Time
		baseTo      time.Time
		compareFrom time.Time
		compareTo   time.Time
		accountIDs  []uuid.UUID
		trnType     *entities.TrnType
		topLimit    int
	}

	// Parse and validate inputs
	{
		var err error
		input.userID, err = uuid.Parse(userID)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to parse user id", err)
			return nil, inerr.NewErrValidation("user_id", "invalid uuid type")
		}

		// Determine periods
		if query.Period != "" {
			input.baseFrom, input.baseTo, input.compareFrom, input.compareTo = u.calculatePeriods(query.Period)
		} else {
			// Custom periods
			if query.BaseFrom == "" || query.BaseTo == "" || query.CompareFrom == "" || query.CompareTo == "" {
				return nil, inerr.NewErrValidation("period", "either period or all custom dates (base_from, base_to, compare_from, compare_to) required")
			}

			input.baseFrom, err = time.Parse(time.DateOnly, query.BaseFrom)
			if err != nil {
				return nil, inerr.NewErrValidation("base_from", "invalid date format, use YYYY-MM-DD")
			}

			input.baseTo, err = time.Parse(time.DateOnly, query.BaseTo)
			if err != nil {
				return nil, inerr.NewErrValidation("base_to", "invalid date format, use YYYY-MM-DD")
			}
			input.baseTo = utils.EndOfDate(input.baseTo)

			input.compareFrom, err = time.Parse(time.DateOnly, query.CompareFrom)
			if err != nil {
				return nil, inerr.NewErrValidation("compare_from", "invalid date format, use YYYY-MM-DD")
			}

			input.compareTo, err = time.Parse(time.DateOnly, query.CompareTo)
			if err != nil {
				return nil, inerr.NewErrValidation("compare_to", "invalid date format, use YYYY-MM-DD")
			}
			input.compareTo = utils.EndOfDate(input.compareTo)
		}

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
			if t != entities.Deposit && t != entities.Withdrawal {
				return nil, inerr.NewErrValidation("type", "invalid transaction type, must be deposit or withdrawal")
			}
			input.trnType = &t
		}

		// Set top limit
		if query.TopLimit <= 0 {
			query.TopLimit = 5
		}
		input.topLimit = query.TopLimit
	}

	// Get user for currency conversion
	user, err := u.usersRepo.FindByID(ctx, input.userID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get user", err)
		return nil, err
	}

	scale := user.CurrencyCode.Scale()

	// Fetch base period stats
	baseStats, err := u.transactionsRepo.GetStatsByCategory(ctx, input.userID, input.baseFrom, input.baseTo, input.accountIDs, input.trnType)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get base stats", err)
		return nil, err
	}

	// Fetch compare period stats
	compareStats, err := u.transactionsRepo.GetStatsByCategory(ctx, input.userID, input.compareFrom, input.compareTo, input.accountIDs, input.trnType)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get compare stats", err)
		return nil, err
	}

	// Build category maps for easy lookup
	baseMap := make(map[int]entities.CategoryStatsItem)
	for _, item := range baseStats {
		baseMap[item.CategoryID] = item
	}

	compareMap := make(map[int]entities.CategoryStatsItem)
	for _, item := range compareStats {
		compareMap[item.CategoryID] = item
	}

	// Calculate period totals
	basePeriod := u.calculatePeriodStats(baseStats, scale, input.baseFrom, input.baseTo)
	comparePeriod := u.calculatePeriodStats(compareStats, scale, input.compareFrom, input.compareTo)

	// Calculate deltas
	delta := DeltaStats{
		Income:  u.calculateDelta(basePeriod.Income, comparePeriod.Income),
		Expense: u.calculateDelta(basePeriod.Expense, comparePeriod.Expense),
		Net:     u.calculateDelta(basePeriod.Net, comparePeriod.Net),
	}

	// Calculate top changes by category
	topChanges := u.calculateTopChanges(ctx, baseMap, compareMap, scale, user.LanguageCode, input.topLimit)

	response := &StatsCompareView{
		Base:       basePeriod,
		Compare:    comparePeriod,
		Delta:      delta,
		TopChanges: topChanges,
	}

	return response, nil
}

func (u *GetStatsCompareUsecase) calculatePeriods(period string) (baseFrom, baseTo, compareFrom, compareTo time.Time) {
	now := time.Now()

	switch period {
	case "this_month_vs_last_month":
		// Compare: this month (from 1st to today)
		compareFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		compareTo = now

		// Base: last month (entire month)
		baseFrom = compareFrom.AddDate(0, -1, 0)
		baseTo = compareFrom.Add(-time.Second)

	case "last_7_days_vs_previous_7_days":
		compareTo = now
		compareFrom = now.AddDate(0, 0, -7)
		baseTo = compareFrom.Add(-time.Second)
		baseFrom = baseTo.AddDate(0, 0, -7)

	case "this_year_vs_last_year":
		// Compare: this year (from Jan 1 to today)
		compareFrom = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		compareTo = now

		// Base: same period last year
		baseFrom = compareFrom.AddDate(-1, 0, 0)
		baseTo = compareTo.AddDate(-1, 0, 0)

	default:
		// Default to last 30 days vs previous 30 days
		compareTo = now
		compareFrom = now.AddDate(0, 0, -30)
		baseTo = compareFrom.Add(-time.Second)
		baseFrom = baseTo.AddDate(0, 0, -30)
	}

	return baseFrom, baseTo, compareFrom, compareTo
}

func (u *GetStatsCompareUsecase) calculatePeriodStats(stats []entities.CategoryStatsItem, scale int, from, to time.Time) PeriodStats {
	var totalIncome, totalExpense int64
	var count int

	for _, item := range stats {
		if item.Total >= 0 {
			totalIncome += item.Total
		} else {
			totalExpense += -item.Total
		}
		count += item.Count
	}

	return PeriodStats{
		From:    from.Format(time.DateOnly),
		To:      to.Format(time.DateOnly),
		Income:  entities.MajorFromMinor(totalIncome, scale),
		Expense: entities.MajorFromMinor(totalExpense, scale),
		Net:     entities.MajorFromMinor(totalIncome-totalExpense, scale),
		Count:   count,
	}
}

func (u *GetStatsCompareUsecase) calculateDelta(base, compare float64) DeltaValue {
	abs := compare - base
	pct := 0.0

	if base != 0 {
		pct = abs / base
	} else if compare != 0 {
		// If base is 0 but compare is not, show as 100% increase or special value
		pct = 1.0
	}

	return DeltaValue{
		Abs: abs,
		Pct: pct,
	}
}

func (u *GetStatsCompareUsecase) calculateTopChanges(
	ctx context.Context,
	baseMap, compareMap map[int]entities.CategoryStatsItem,
	scale int,
	lang entities.Language,
	topLimit int,
) TopChanges {
	// Collect all unique category IDs
	categoryIDs := make(map[int]bool)
	for id := range baseMap {
		categoryIDs[id] = true
	}
	for id := range compareMap {
		categoryIDs[id] = true
	}

	// Calculate changes for each category
	var changes []CategoryChange

	for categoryID := range categoryIDs {
		baseItem, hasBase := baseMap[categoryID]
		compareItem, hasCompare := compareMap[categoryID]

		// Use absolute values for expenses (they come as negative)
		baseTotal := 0.0
		if hasBase {
			baseTotal = math.Abs(entities.MajorFromMinor(baseItem.Total, scale))
		}

		compareTotal := 0.0
		if hasCompare {
			compareTotal = math.Abs(entities.MajorFromMinor(compareItem.Total, scale))
		}

		deltaAbs := compareTotal - baseTotal

		// Only include if there's a meaningful change
		if math.Abs(deltaAbs) < 0.01 {
			continue
		}

		// Get category details
		category, err := u.categoriesRepo.FindByID(ctx, categoryID)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to get category", err)
			continue
		}

		deltaPct := 0.0
		if baseTotal != 0 {
			deltaPct = deltaAbs / baseTotal
		} else if compareTotal != 0 {
			deltaPct = 1.0
		}

		changes = append(changes, CategoryChange{
			CategoryID: categoryID,
			Name:       category.GetName(lang),
			Emoji:      category.Emoji,
			Base:       baseTotal,
			Compare:    compareTotal,
			DeltaAbs:   deltaAbs,
			DeltaPct:   deltaPct,
		})
	}

	// Sort by absolute delta (descending)
	sort.Slice(changes, func(i, j int) bool {
		return math.Abs(changes[i].DeltaAbs) > math.Abs(changes[j].DeltaAbs)
	})

	// Take top N
	if len(changes) > topLimit {
		changes = changes[:topLimit]
	}

	return TopChanges{
		ExpenseByCategory: changes,
	}
}
