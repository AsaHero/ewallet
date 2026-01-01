package repository

import (
	"context"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/pkg/database/postgres"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func (r *accountBalanceLogRepo) GetBalanceTimeseries(
	ctx context.Context,
	filter *entities.BalanceTimeseriesFilter,
) (map[uuid.UUID][]entities.BalanceTimeseriesPoint, error) {
	db := postgres.FromContext(ctx, r.db)

	// Determine date truncation format based on group_by
	var truncFunc string
	var formatStr string
	switch filter.GroupBy {
	case "day":
		truncFunc = "day"
		formatStr = "YYYY-MM-DD"
	case "week":
		truncFunc = "week"
		formatStr = "YYYY-MM-DD" // first day of week
	case "month":
		truncFunc = "month"
		formatStr = "YYYY-MM"
	default:
		truncFunc = "day"
		formatStr = "YYYY-MM-DD"
	}

	// Build the query with timezone-aware bucket calculation
	query := db.NewSelect().
		ColumnExpr("account_id").
		ColumnExpr("TO_CHAR(DATE_TRUNC(?, occurred_at AT TIME ZONE ?), ?) AS bucket", truncFunc, filter.Timezone, formatStr).
		ColumnExpr("(ARRAY_AGG(balance_before ORDER BY occurred_at ASC))[1] AS balance_open").
		ColumnExpr("(ARRAY_AGG(balance_after  ORDER BY occurred_at DESC))[1] AS balance_close").
		ColumnExpr("MIN(LEAST(balance_before, balance_after)) AS min_balance").
		ColumnExpr("MAX(GREATEST(balance_before, balance_after)) AS max_balance").
		ColumnExpr("COUNT(*) AS tx_count").
		TableExpr("account_balance_log").
		Where("user_id = ?", filter.UserID.String()).
		Where("occurred_at >= ?", filter.From).
		Where("occurred_at <= ?", filter.To)

	// Filter by account IDs if specified
	if len(filter.AccountIDs) > 0 {
		accountIDStrs := make([]string, len(filter.AccountIDs))
		for i, id := range filter.AccountIDs {
			accountIDStrs[i] = id.String()
		}
		query = query.Where("account_id IN (?)", bun.In(accountIDStrs))
	}

	query = query.
		Group("account_id").
		Group("bucket").
		Order("account_id", "bucket")

	// Execute query
	type resultRow struct {
		AccountID    string `bun:"account_id"`
		Bucket       string `bun:"bucket"`
		BalanceOpen  int64  `bun:"balance_open"`
		BalanceClose int64  `bun:"balance_close"`
		MinBalance   int64  `bun:"min_balance"`
		MaxBalance   int64  `bun:"max_balance"`
		TxCount      int    `bun:"tx_count"`
	}

	var rows []resultRow
	err := query.Scan(ctx, &rows)
	if err != nil {
		return nil, postgres.Error(err, AccountBalanceLog{})
	}

	// Group results by account_id and fill in missing buckets
	result := make(map[uuid.UUID][]entities.BalanceTimeseriesPoint)

	// Generate all expected buckets
	buckets := generateBuckets(filter.From, filter.To, filter.GroupBy, filter.Timezone)

	// Group rows by account
	accountRows := make(map[uuid.UUID][]resultRow)
	for _, row := range rows {
		accountID, _ := uuid.Parse(row.AccountID)
		accountRows[accountID] = append(accountRows[accountID], row)
	}

	// Process each account
	for accountID, accountData := range accountRows {
		points := make([]entities.BalanceTimeseriesPoint, 0, len(buckets))

		// Create a map for quick lookup
		bucketData := make(map[string]resultRow)
		for _, row := range accountData {
			bucketData[row.Bucket] = row
		}

		// Fill in all buckets, carrying forward balance for empty buckets
		var lastCloseBalance int64
		for _, bucket := range buckets {
			if data, exists := bucketData[bucket]; exists {
				// Bucket has transactions
				delta := data.BalanceClose - data.BalanceOpen
				points = append(points, entities.BalanceTimeseriesPoint{
					Timestamp:    bucket,
					BalanceOpen:  data.BalanceOpen,
					BalanceClose: data.BalanceClose,
					MinBalance:   data.MinBalance,
					MaxBalance:   data.MaxBalance,
					Delta:        delta,
					TxCount:      data.TxCount,
				})
				lastCloseBalance = data.BalanceClose
			} else {
				// Empty bucket - carry forward last close balance
				points = append(points, entities.BalanceTimeseriesPoint{
					Timestamp:    bucket,
					BalanceOpen:  lastCloseBalance,
					BalanceClose: lastCloseBalance,
					MinBalance:   lastCloseBalance,
					MaxBalance:   lastCloseBalance,
					Delta:        0,
					TxCount:      0,
				})
			}
		}

		result[accountID] = points
	}

	return result, nil
}

// generateBuckets creates all expected bucket timestamps between from and to
func generateBuckets(from, to time.Time, groupBy, timezone string) []string {
	buckets := []string{}

	// Load timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC // fallback to UTC
	}

	// Convert to user timezone
	current := from.In(loc)
	end := to.In(loc)

	switch groupBy {
	case "day":
		for !current.After(end) {
			buckets = append(buckets, current.Format("2006-01-02"))
			current = current.AddDate(0, 0, 1)
		}
	case "week":
		// Start from the beginning of the week
		for current.Weekday() != time.Monday {
			current = current.AddDate(0, 0, -1)
		}
		for !current.After(end) {
			buckets = append(buckets, current.Format("2006-01-02"))
			current = current.AddDate(0, 0, 7)
		}
	case "month":
		// Start from the beginning of the month
		current = time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, loc)
		for !current.After(end) {
			buckets = append(buckets, current.Format("2006-01"))
			current = current.AddDate(0, 1, 0)
		}
	default:
		// Default to day
		for !current.After(end) {
			buckets = append(buckets, current.Format("2006-01-02"))
			current = current.AddDate(0, 0, 1)
		}
	}

	return buckets
}
