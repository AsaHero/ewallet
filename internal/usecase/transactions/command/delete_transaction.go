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

type DeleteTransactionUsecase struct {
	contextTimeout   time.Duration
	logger           *logger.Logger
	txManager        postgres.TxManager
	accountsRepo     entities.AccountRepository
	transactionsRepo entities.TransactionRepository
}

func NewDeleteTransactionUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	txManager postgres.TxManager,
	accountsRepo entities.AccountRepository,
	transactionsRepo entities.TransactionRepository,
) *DeleteTransactionUsecase {
	return &DeleteTransactionUsecase{
		contextTimeout:   timeout,
		logger:           logger,
		txManager:        txManager,
		accountsRepo:     accountsRepo,
		transactionsRepo: transactionsRepo,
	}
}

type DeleteTransactionCommand struct {
	UserID        string
	TransactionID string
}

func (c *DeleteTransactionUsecase) DeleteTransaction(ctx context.Context, cmd *DeleteTransactionCommand) error {
	ctx, cancel := context.WithTimeout(ctx, c.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("transactions"), "DeleteTransaction",
		attribute.String("user_id", cmd.UserID),
		attribute.String("transaction_id", cmd.TransactionID),
	)
	defer func() { end(nil) }()

	userID, err := uuid.Parse(cmd.UserID)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to parse user id", err)
		return inerr.NewErrValidation("user_id", "invalid uuid type")
	}

	transactionID, err := uuid.Parse(cmd.TransactionID)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to parse transaction id", err)
		return inerr.NewErrValidation("transaction_id", "invalid uuid type")
	}

	err = c.txManager.WithTx(ctx, func(ctx context.Context) error {
		transaction, err := c.transactionsRepo.GetByID(ctx, transactionID)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to get transaction", err)
			return err
		}

		if transaction == nil {
			return inerr.NewErrNotFound("transaction")
		}

		if transaction.UserID != userID {
			return inerr.NewErrNotFound("transaction")
		}

		account, err := c.accountsRepo.GetByIDForUpdate(ctx, transaction.AccountID)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to get account", err)
			return err
		}

		err = account.RevertTransaction(transaction)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to revert transaction", err)
			return err
		}

		err = c.accountsRepo.Save(ctx, account)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to save account", err)
			return err
		}

		err = c.transactionsRepo.Delete(ctx, transactionID)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to delete transaction", err)
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
