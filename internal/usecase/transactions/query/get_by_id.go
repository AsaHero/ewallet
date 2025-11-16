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

type GetByIDUsecase struct {
	contextTimeout   time.Duration
	logger           *logger.Logger
	transactionsRepo entities.TransactionRepository
}

func NewGetByIDUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	transactionsRepo entities.TransactionRepository,
) *GetByIDUsecase {
	return &GetByIDUsecase{
		contextTimeout:   timeout,
		transactionsRepo: transactionsRepo,
		logger:           logger,
	}
}

func (u *GetByIDUsecase) GetByID(ctx context.Context, id string) (_ *entities.Transaction, err error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("transactions"), "GetByID",
		attribute.String("transactions_id", id),
	)
	defer func() { end(err) }()

	var input struct {
		trnID uuid.UUID
	}
	{
		var err error
		input.trnID, err = uuid.Parse(id)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to parse transaction id", err)
			return nil, inerr.NewErrValidation("transaction_id", "invalud uuid type")
		}
	}

	trn, err := u.transactionsRepo.GetByID(ctx, input.trnID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get transaction", err)
		return nil, err
	}

	return trn, nil
}
