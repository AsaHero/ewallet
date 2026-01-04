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

type GetDebtByIDUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	debtsRepo      entities.DebtRepository
}

func NewGetDebtByIDUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	debtsRepo entities.DebtRepository,
) *GetDebtByIDUsecase {
	return &GetDebtByIDUsecase{
		contextTimeout: timeout,
		debtsRepo:      debtsRepo,
		logger:         logger,
	}
}

func (u *GetDebtByIDUsecase) GetByID(ctx context.Context, id string) (_ *entities.Debt, err error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("debts"), "GetByID",
		attribute.String("debt_id", id),
	)
	defer func() { end(err) }()

	var input struct {
		debtID uuid.UUID
	}
	{
		var err error
		input.debtID, err = uuid.Parse(id)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to parse debt id", err)
			return nil, inerr.NewErrValidation("debt_id", "invalid uuid type")
		}
	}

	debt, err := u.debtsRepo.GetByID(ctx, input.debtID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get debt", err)
		return nil, err
	}

	return debt, nil
}
