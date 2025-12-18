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
	usersRepo         entities.UserRepository
	accountsRepo      entities.AccountRepository
	transactionsRepo  entities.TransactionRepository
	categoryRepo      entities.CategoryRepository
	subcategoriesRepo entities.SubcategoryRepository
}

func NewUpdateTransactionUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	txManager postgres.TxManager,
	usersRepo entities.UserRepository,
	accountsRepo entities.AccountRepository,
	transactionsRepo entities.TransactionRepository,
	categoriesRepo entities.CategoryRepository,
	subcategoriesRepo entities.SubcategoryRepository,
) *UpdateTransactionUsecase {
	return &UpdateTransactionUsecase{
		contextTimeout:    timeout,
		usersRepo:         usersRepo,
		accountsRepo:      accountsRepo,
		transactionsRepo:  transactionsRepo,
		categoryRepo:      categoriesRepo,
		subcategoriesRepo: subcategoriesRepo,
		logger:            logger,
		txManager:         txManager,
	}
}

type UpdateTransactionCommand struct {
	UserID               string
	TransactionID        string
	CategoryID           *int
	SubcategoryID        *int
	Type                 string
	Amount               float64
	CurrencyCode         string
	OriginalAmount       *float64
	OriginalCurrencyCode *string
	FxRate               *float64
	Note                 string
	PerformedAt          *time.Time
}

func (c *UpdateTransactionUsecase) UpdateTransaction(ctx context.Context, cmd *UpdateTransactionCommand) (_ *entities.Transaction, err error) {
	ctx, cancel := context.WithTimeout(ctx, c.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("transactions"), "UpdateTransaction",
		attribute.String("user_id", cmd.UserID),
		attribute.String("transaction_id", cmd.TransactionID),
	)
	defer func() { end(err) }()

	userID, err := uuid.Parse(cmd.UserID)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to parse user id", err)
		return nil, inerr.NewErrValidation("user_id", "invalid uuid type")
	}

	transactionID, err := uuid.Parse(cmd.TransactionID)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to parse transaction id", err)
		return nil, inerr.NewErrValidation("transaction_id", "invalid uuid type")
	}

	var input struct {
		category    *entities.Category
		subcategory *entities.Subcategory
		trnType     entities.TrnType
	}
	{
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

		if cmd.Type == "deposit" {
			input.trnType = entities.Deposit
		} else {
			input.trnType = entities.Withdrawal
		}
	}

	user, err := c.usersRepo.FindByID(ctx, userID)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to get user", err)
		return nil, err
	}

	var transaction *entities.Transaction
	err = c.txManager.WithTx(ctx, func(ctx context.Context) error {
		// 1. Get existing transaction
		transaction, err = c.transactionsRepo.GetByID(ctx, transactionID)
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

		// 2. Get account and revert old transaction
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

		// 3. Update transaction fields
		// 3. Update transaction fields
		err = transaction.Update(
			input.category,
			input.subcategory,
			input.trnType,
			cmd.Amount,
			user.CurrencyCode,
			cmd.OriginalAmount,
			cmd.OriginalCurrencyCode,
			cmd.FxRate,
			cmd.Note,
			cmd.PerformedAt,
		)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to update transaction", err)
			return err
		}

		// 4. Apply new transaction to account
		err = account.ApplyTransaction(transaction)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to apply transaction", err)
			return err
		}

		err = c.accountsRepo.Save(ctx, account)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to save account", err)
			return err
		}

		// 5. Save transaction
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
