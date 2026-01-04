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

type UpdateTransactionUsecase struct {
	contextTimeout    time.Duration
	logger            *logger.Logger
	txManager         postgres.TxManager
	transactionsRepo  entities.TransactionRepository
	categoryRepo      entities.CategoryRepository
	subcategoriesRepo entities.SubcategoryRepository
}

func NewUpdateTransactionUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	txManager postgres.TxManager,
	transactionsRepo entities.TransactionRepository,
	categoriesRepo entities.CategoryRepository,
	subcategoriesRepo entities.SubcategoryRepository,
) *UpdateTransactionUsecase {
	return &UpdateTransactionUsecase{
		contextTimeout:    timeout,
		transactionsRepo:  transactionsRepo,
		categoryRepo:      categoriesRepo,
		subcategoriesRepo: subcategoriesRepo,
		logger:            logger,
		txManager:         txManager,
	}
}

type UpdateTransactionCommand struct {
	UserID        string
	TransactionID string
	CategoryID    *int
	SubcategoryID *int
	Note          string
	PerformedAt   *time.Time
}

func (c *UpdateTransactionUsecase) UpdateTransaction(ctx context.Context, cmd *UpdateTransactionCommand) (_ *entities.Transaction, err error) {
	ctx, cancel := context.WithTimeout(ctx, c.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("transactions"), "UpdateTransaction",
		attribute.String("user_id", cmd.UserID),
		attribute.String("transaction_id", cmd.TransactionID),
	)
	defer func() { end(err) }()

	var input struct {
		userID        uuid.UUID
		transactionID uuid.UUID
		category      *entities.Category
		subcategory   *entities.Subcategory
	}
	{
		input.userID, err = uuid.Parse(cmd.UserID)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to parse user id", err)
			return nil, inerr.NewErrValidation("user_id", "invalid uuid type")
		}

		input.transactionID, err = uuid.Parse(cmd.TransactionID)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to parse transaction id", err)
			return nil, inerr.NewErrValidation("transaction_id", "invalid uuid type")
		}

		if cmd.CategoryID != nil {
			category, err := c.categoryRepo.FindByID(ctx, *cmd.CategoryID)
			if err != nil {
				c.logger.ErrorContext(ctx, "failed to get category", err)
				return nil, err
			}

			input.category = category
		}

		if cmd.SubcategoryID != nil {
			subcategory, err := c.subcategoriesRepo.FindByID(ctx, *cmd.SubcategoryID)
			if err != nil {
				c.logger.ErrorContext(ctx, "failed to get subcategory", err)
				return nil, err
			}

			input.subcategory = subcategory
		}
	}

	var transaction *entities.Transaction
	err = c.txManager.WithTx(ctx, func(ctx context.Context) error {
		// 1. Get existing transaction
		transaction, err = c.transactionsRepo.GetByID(ctx, input.transactionID)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to get transaction", err)
			return err
		}

		if transaction == nil {
			return inerr.NewErrNotFound("transaction")
		}

		if transaction.UserID != input.userID {
			return inerr.NewErrNotFound("transaction")
		}

		// 2. Update metadata fields only
		err = transaction.Update(
			input.category,
			input.subcategory,
			cmd.Note,
			cmd.PerformedAt,
		)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to update transaction", err)
			return err
		}

		// 3. Save transaction
		err = c.transactionsRepo.Save(ctx, transaction)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to save transaction", err)
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return transaction, nil
}
