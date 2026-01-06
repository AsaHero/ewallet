package command

import (
	"context"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/inerr"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/google/uuid"
	"github.com/shogo82148/pointer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type CreateDebtUsecase struct {
	timeout          time.Duration
	logger           *logger.Logger
	debtsRepo        entities.DebtRepository
	transactionsRepo entities.TransactionRepository
}

func NewCreateDebtUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	debtsRepo entities.DebtRepository,
	transactionsRepo entities.TransactionRepository,
) *CreateDebtUsecase {
	return &CreateDebtUsecase{
		timeout:          timeout,
		logger:           logger,
		debtsRepo:        debtsRepo,
		transactionsRepo: transactionsRepo,
	}
}

type CreateDebtCommand struct {
	TransactionID string
	Name          *string
	Note          *string
	DueAt         *time.Time
}

func (c *CreateDebtUsecase) CreateDebt(ctx context.Context, cmd *CreateDebtCommand) (_ *entities.Debt, err error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("debts"), "CreateDebt",
		attribute.String("transaction_id", cmd.TransactionID),
		attribute.String("due_at", pointer.TimeValue(cmd.DueAt).String()),
		attribute.String("name", pointer.StringValue(cmd.Name)),
		attribute.String("note", pointer.StringValue(cmd.Note)),
	)
	defer func() { end(err) }()

	var input struct {
		transactionID uuid.UUID
	}
	{
		var err error

		input.transactionID, err = uuid.Parse(cmd.TransactionID)
		if err != nil {
			return nil, inerr.NewErrValidation("transaction_id", "invalid uuid type")
		}
	}

	transaction, err := c.transactionsRepo.GetByID(ctx, input.transactionID)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to get transaction", err)
		return nil, err
	}

	var deptType entities.DebtType
	switch transaction.Type {
	case entities.Deposit:
		deptType = entities.Borrow
	case entities.Withdrawal:
		deptType = entities.Lend
	default:
		return nil, inerr.NewErrValidation("transaction_type", "invalid transaction type")
	}

	debt, err := entities.NewDebt(
		transaction.UserID,
		input.transactionID,
		deptType,
	)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to create debt", err)
		return nil, err
	}

	if err := debt.SetAmountMinor(transaction.Amount, transaction.CurrencyCode); err != nil {
		c.logger.ErrorContext(ctx, "failed to set amount minor", err)
		return nil, err
	}

	if cmd.DueAt != nil {
		debt.SetDueAt(*cmd.DueAt)
	}

	if cmd.Name != nil {
		debt.SetName(*cmd.Name)
	}

	if cmd.Note != nil {
		debt.SetNote(*cmd.Note)
	}

	if err := c.debtsRepo.Save(ctx, debt); err != nil {
		c.logger.ErrorContext(ctx, "failed to save debt", err)
		return nil, err
	}

	return debt, nil

}
