package command

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

type CreateTransactionUsecase struct {
	contextTimeout   time.Duration
	logger           *logger.Logger
	usersRepo        entities.UserRepository
	accountsRepo     entities.AccountRepository
	transactionsRepo entities.TransactionRepository
	categoryRepo     entities.CategoryRepository
}

func NewCreateTransactionUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	usersRepo entities.UserRepository,
	accountsRepo entities.AccountRepository,
	transactionsRepo entities.TransactionRepository,
	categoriesRepo entities.CategoryRepository,
) *CreateTransactionUsecase {
	return &CreateTransactionUsecase{
		contextTimeout:   timeout,
		usersRepo:        usersRepo,
		accountsRepo:     accountsRepo,
		transactionsRepo: transactionsRepo,
		categoryRepo:     categoriesRepo,
		logger:           logger,
	}
}

type CreateTransactionCommand struct {
	UserID       string
	AccountID    string
	CategoryID   *int
	Type         string
	Amount       float64
	CurrencyCode string
	Note         string
	PerformedAt  *time.Time
}

func (c *CreateTransactionUsecase) CreateTransaction(ctx context.Context, cmd *CreateTransactionCommand) (_ *entities.Transaction, err error) {
	ctx, cancel := context.WithTimeout(ctx, c.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("transactions"), "CreateTransaction",
		attribute.String("user_id", cmd.UserID),
		attribute.String("account_id", cmd.AccountID),
	)
	defer func() { end(err) }()

	var input struct {
		userID    uuid.UUID
		accountID uuid.UUID
		category  entities.Category
		trnType   entities.TrnType
	}
	{
		var err error
		input.userID, err = uuid.Parse(cmd.UserID)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to parse user id", err)
			return nil, inerr.NewErrValidation("user_id", "invalud uuid type")
		}

		input.accountID, err = uuid.Parse(cmd.AccountID)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to parse account id", err)
			return nil, inerr.NewErrValidation("account_id", "invalud uuid type")
		}

		if cmd.CategoryID != nil {
			category, err := c.categoryRepo.FindByID(ctx, *cmd.CategoryID)
			if err != nil {
				c.logger.ErrorContext(ctx, "failed to get category", err)
				return nil, err
			}

			input.category = *category
		}

		input.trnType = entities.TrnType(cmd.Type)
	}

	user, err := c.usersRepo.FindByID(ctx, input.userID)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to get user", err)
		return nil, err
	}

	account, err := c.accountsRepo.GetByID(ctx, input.accountID)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to get account", err)
		return nil, err
	}

	transaction, err := entities.NewTransaction(
		user.ID,
		account.ID,
		input.category,
		input.trnType,
		0,
		user.CurrencyCode,
		cmd.Note,
	)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to create transaction", err)
		return nil, err
	}

	transaction.SetAmountMajor(cmd.Amount)

	if cmd.PerformedAt != nil {
		transaction.Performed(*cmd.PerformedAt)
	}

	err = c.transactionsRepo.Save(ctx, transaction)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to create transaction", err)
		return nil, err
	}

	return transaction, nil
}
