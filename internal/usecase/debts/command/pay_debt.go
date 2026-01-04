package command

import (
	"context"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/inerr"
	"github.com/AsaHero/e-wallet/pkg/database/postgres"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type PayDebtUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	txManager      postgres.TxManager
	debtsRepo      entities.DebtRepository
}

func NewPayDebtUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	txManager postgres.TxManager,
	debtsRepo entities.DebtRepository,
) *PayDebtUsecase {
	return &PayDebtUsecase{
		contextTimeout: timeout,
		debtsRepo:      debtsRepo,
		logger:         logger,
		txManager:      txManager,
	}
}

type PayDebtCommand struct {
	UserID string
	DebtID string
	PaidAt *time.Time
}

func (c *PayDebtUsecase) PayDebt(ctx context.Context, cmd *PayDebtCommand) (_ *entities.Debt, err error) {
	ctx, cancel := context.WithTimeout(ctx, c.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("debts"), "PayDebt",
		attribute.String("user_id", cmd.UserID),
		attribute.String("debt_id", cmd.DebtID),
	)
	defer func() { end(err) }()

	var input struct {
		userID uuid.UUID
		debtID uuid.UUID
	}
	{
		input.userID, err = uuid.Parse(cmd.UserID)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to parse user id", err)
			return nil, inerr.NewErrValidation("user_id", "invalid uuid type")
		}

		input.debtID, err = uuid.Parse(cmd.DebtID)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to parse debt id", err)
			return nil, inerr.NewErrValidation("debt_id", "invalid uuid type")
		}
	}

	var debt *entities.Debt
	err = c.txManager.WithTx(ctx, func(ctx context.Context) error {
		// 1. Get existing debt
		debt, err = c.debtsRepo.GetByID(ctx, input.debtID)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to get debt", err)
			return err
		}

		if debt == nil {
			return inerr.NewErrNotFound("debt")
		}

		if debt.UserID != input.userID {
			return inerr.NewErrNotFound("debt")
		}

		// 2. Check if debt is already paid or cancelled
		if debt.Status == entities.Paid {
			return inerr.NewErrValidation("debt", "already paid")
		}

		if debt.Status == entities.Cancelled {
			return inerr.NewErrValidation("debt", "cannot pay cancelled debt")
		}

		// 3. Mark debt as paid
		paidAt := time.Now()
		if cmd.PaidAt != nil {
			paidAt = *cmd.PaidAt
		}
		debt.MarkAsPaid(paidAt)

		// 4. Save debt
		err = c.debtsRepo.Save(ctx, debt)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to save debt", err)
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return debt, nil
}
