package query

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/AsaHero/e-wallet/internal/delivery/api/models"
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
	usersRepo        entities.UserRepository
	transactionsRepo entities.TransactionRepository
}

func NewGetByFilterUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	usersRepo entities.UserRepository,
	transactionsRepo entities.TransactionRepository,
) *GetByFilterUsecase {
	return &GetByFilterUsecase{
		contextTimeout:   timeout,
		usersRepo:        usersRepo,
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
	CategoryIDs []string `form:"category_ids"`
	AccountIDs  []string `form:"account_ids"`
	MinAmount   *int64   `form:"min_amount"`
	MaxAmount   *int64   `form:"max_amount"`
	Search      string   `form:"search"`
}

func (u *GetByFilterUsecase) GetByFilter(ctx context.Context, userID string, query *GetByFilterQuery) (_ *models.TransactionsResponse, err error) {
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
			return nil, inerr.NewErrValidation("user_id", "invalid uuid type")
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
				return nil, inerr.NewErrValidation("from", "invalid date format")
			}
			input.from = &from
		}

		if query.To != "" {
			to, err := time.Parse(time.RFC3339, query.To)
			if err != nil {
				u.logger.ErrorContext(ctx, "failed to parse to", err)
				return nil, inerr.NewErrValidation("to", "invalid date format")
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
					return nil, inerr.NewErrValidation("account_id", "invalid uuid type")
				}
				input.accountIDs = append(input.accountIDs, uid)
			}
		}

		if len(query.CategoryIDs) > 0 {
			input.categoryIDs = make([]int, 0, len(query.CategoryIDs))
			for _, id := range query.CategoryIDs {
				i, err := strconv.Atoi(id)
				if err != nil {
					u.logger.ErrorContext(ctx, "failed to parse category id", err)
					return nil, inerr.NewErrValidation("category_id", "invalid int type")
				}
				input.categoryIDs = append(input.categoryIDs, i)
			}
		}

	}

	user, err := u.usersRepo.FindByID(ctx, input.userID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get user", err)
		return nil, err
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

	var (
		transactions []*entities.Transaction
		total        int
		totals       map[entities.TrnType]int64
		errTrn       error
		errTotals    error
		wg           sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		transactions, total, errTrn = u.transactionsRepo.GetByFilter(ctx, filter)
	}()

	go func() {
		defer wg.Done()
		totals, errTotals = u.transactionsRepo.GetFilterTotals(ctx, filter)
	}()

	wg.Wait()

	if errTrn != nil {
		u.logger.ErrorContext(ctx, "failed to get transaction", errTrn)
		return nil, errTrn
	}

	resp := &models.TransactionsResponse{
		Items: make([]models.Transaction, 0, len(transactions)),
		Pagination: models.PaginationResponse{
			Limit:  query.Limit,
			Offset: query.Offset,
			Total:  int64(total),
		},
	}

	if errTotals != nil {
		u.logger.ErrorContext(ctx, "failed to get transaction totals", errTotals)
	} else {
		resp.TotalIncome = entities.MajorFromMinor(totals[entities.Deposit], user.CurrencyCode.Scale())
		resp.TotalExpense = entities.MajorFromMinor(totals[entities.Withdrawal], user.CurrencyCode.Scale())
		resp.NetBalance = resp.TotalIncome - resp.TotalExpense
	}

	for _, trn := range transactions {
		item := models.Transaction{
			ID:                   trn.ID.String(),
			UserID:               trn.UserID.String(),
			AccountID:            trn.AccountID.String(),
			Type:                 trn.Type.String(),
			Status:               trn.Status.String(),
			Amount:               trn.AmountMajor(),
			CurrencyCode:         trn.CurrencyCode.String(),
			OriginalAmount:       pointer.Float64(trn.OriginalAmountMajor()),
			OriginalCurrencyCode: pointer.String(trn.OriginalCurrencyCode.String()),
			FxRate:               pointer.Float64(trn.FxRate),
			Note:                 trn.RowText,
			PerformedAt:          pointer.TimeOrNil(trn.PerformedAt),
			RejectedAt:           pointer.TimeOrNil(trn.RejectedAt),
			CreatedAt:            trn.CreatedAt,
		}

		if trn.Category != nil {
			item.CategoryID = pointer.IntOrNil(trn.Category.ID.Int())
		}

		if trn.Subcategory != nil {
			item.SubcategoryID = pointer.IntOrNil(trn.Subcategory.ID)
		}

		resp.Items = append(resp.Items, item)
	}

	return resp, nil
}
