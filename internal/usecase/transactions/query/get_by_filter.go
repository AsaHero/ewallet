package query

import (
	"context"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/inerr"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/google/uuid"
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
	UserID string
	Limit  int
	Offset int
}

func (u *GetByFilterUsecase) GetByFilter(ctx context.Context, query *GetByFilterQuery) (_ []*entities.Transaction, _ int, err error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("transactions"), "GetByFilter",
		attribute.String("user_id", query.UserID),
	)
	defer func() { end(err) }()

	var input struct {
		userID uuid.UUID
	}
	{
		var err error
		input.userID, err = uuid.Parse(query.UserID)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to parse transaction id", err)
			return nil, 0, inerr.NewErrValidation("transaction_id", "invalud uuid type")
		}
	}

	trn, total, err := u.transactionsRepo.GetByUserID(ctx, query.Limit, query.Offset, input.userID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get transaction", err)
		return nil, 0, err
	}

	return trn, total, nil
}
