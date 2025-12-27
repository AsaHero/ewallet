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
	"github.com/shogo82148/pointer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type GetByFilterUsecase struct {
	contextTimeout   time.Duration
	logger           *logger.Logger
	transactionsRepo entities.TransactionRepository
}

func NewGetByFilterUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	transactionsRepo entities.TransactionRepository,
) *GetByFilterUsecase {
	return &GetByFilterUsecase{
		contextTimeout:   timeout,
		transactionsRepo: transactionsRepo,
		logger:           logger,
	}
}

type GetByFilterQuery struct {
	UserID      string   `form:"user_id"`
	Limit       uint64   `form:"limit"`
	Offset      uint64   `form:"offset"`
	From        string   `form:"from"`
	To          string   `form:"to"`
	Type        string   `form:"type"`
	CategoryIDs []int    `form:"category_ids"`
	AccountIDs  []string `form:"account_ids"`
	MinAmount   *int64   `form:"min_amount"`
	MaxAmount   *int64   `form:"max_amount"`
	Search      string   `form:"search"`
}

func (u *GetByFilterUsecase) GetByFilter(ctx context.Context, userID string, query *GetByFilterQuery) (_ []*entities.Transaction, _ int, err error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("transactions"), "GetByFilter",
		attribute.String("user_id", query.UserID),
	)
	defer func() { end(err) }()

	var input struct {
		userID      uuid.UUID
		limit       int
		offset      int
		from        *time.Time
		to          *time.Time
		trnType     []entities.TrnType
		categoryIDs []int
		accountIDs  []uuid.UUID
	}
	{
		var err error
		input.userID, err = uuid.Parse(userID)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to parse user id", err)
			return nil, 0, inerr.NewErrValidation("user_id", "invalid uuid type")
		}

		if query.Limit == 0 {
			query.Limit = 20
		}
		if query.Limit > 100 {
			query.Limit = 100
		}
		input.limit = int(query.Limit)
		input.offset = int(query.Offset)

		if query.From != "" {
			from, err := time.Parse(time.RFC3339, query.From)
			if err != nil {
				u.logger.ErrorContext(ctx, "failed to parse from", err)
				return nil, 0, inerr.NewErrValidation("from", "invalid date format")
			}
			input.from = &from
		}

		if query.To != "" {
			to, err := time.Parse(time.RFC3339, query.To)
			if err != nil {
				u.logger.ErrorContext(ctx, "failed to parse to", err)
				return nil, 0, inerr.NewErrValidation("to", "invalid date format")
			}
			to = utils.StartOfDate(to).AddDate(0, 0, 1)
			input.to = &to
		}

		if query.Type != "" {
			t := entities.TrnType(query.Type)
			input.trnType = append(input.trnType, t)
		}

		if len(query.AccountIDs) > 0 {
			input.accountIDs = make([]uuid.UUID, 0, len(query.AccountIDs))
			for _, id := range query.AccountIDs {
				uid, err := uuid.Parse(id)
				if err != nil {
					u.logger.ErrorContext(ctx, "failed to parse account id", err)
					return nil, 0, inerr.NewErrValidation("account_id", "invalid uuid type")
				}
				input.accountIDs = append(input.accountIDs, uid)
			}
		}

		if len(query.CategoryIDs) > 0 {
			input.categoryIDs = make([]int, 0, len(query.CategoryIDs))
			for _, id := range query.CategoryIDs {
				input.categoryIDs = append(input.categoryIDs, id)
			}
		}

	}

	filter := &entities.TransactionFilter{
		UserID:      input.userID,
		Limit:       input.limit,
		Offset:      input.offset,
		CategoryIDs: input.categoryIDs,
		AccountIDs:  input.accountIDs,
		From:        input.from,
		To:          input.to,
		Types:       input.trnType,
		Search:      pointer.StringOrNil(query.Search),
		MinAmount:   query.MinAmount,
		MaxAmount:   query.MaxAmount,
	}

	trn, total, err := u.transactionsRepo.GetByFilter(ctx, filter)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get transaction", err)
		return nil, 0, err
	}

	return trn, total, nil
}
