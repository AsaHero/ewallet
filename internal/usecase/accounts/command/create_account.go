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

type CreateAccountUsecase struct {
	contextTimeout   time.Duration
	logger           *logger.Logger
	usersRepo        entities.UserRepository
	accountsRepo     entities.AccountRepository
	transactionsRepo entities.TransactionRepository
	categoryRepo     entities.CategoryRepository
}

func NewCreateAccountUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	usersRepo entities.UserRepository,
	accountsRepo entities.AccountRepository,
	transactionsRepo entities.TransactionRepository,
	categoryRepo entities.CategoryRepository,
) *CreateAccountUsecase {
	return &CreateAccountUsecase{
		contextTimeout:   timeout,
		usersRepo:        usersRepo,
		accountsRepo:     accountsRepo,
		transactionsRepo: transactionsRepo,
		categoryRepo:     categoryRepo,
		logger:           logger,
	}
}

type CreateAccountCommand struct {
	UserID    string
	Name      string
	Balance   float64
	IsDefault bool
}

func (u *CreateAccountUsecase) CreateAccount(ctx context.Context, cmd *CreateAccountCommand) (_ *entities.Account, err error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("accounts"), "CreateAccount",
		attribute.String("user_id", cmd.UserID),
	)
	defer func() { end(err) }()

	var input struct {
		userID uuid.UUID
	}
	{
		var err error
		input.userID, err = uuid.Parse(cmd.UserID)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to parse user id", err)
			return nil, inerr.NewErrValidation("user_id", "invalud uuid type")
		}
	}

	user, err := u.usersRepo.FindByID(ctx, input.userID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get user", err)
		return nil, err
	}

	account, err := entities.NewAccount(user.ID, cmd.Name)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to create account", err)
		return nil, err
	}
	account.SetAmountMajor(cmd.Balance, user.CurrencyCode)
	account.UpdateDefault(cmd.IsDefault)

	err = u.accountsRepo.Save(ctx, account)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to save account", err)
		return nil, err
	}

	if account.Balance > 0 {
		otherCategory, err := u.categoryRepo.FindByID(ctx, entities.OtherCategory.Int())
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to get other category", err)
			return nil, err
		}

		transaction, err := entities.NewTransaction(
			account.UserID,
			account.ID,
			entities.Deposit,
			"Баланс счета",
		)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to create transaction", err)
			return nil, err
		}

		err = transaction.Categorise(otherCategory, nil)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to categorise transaction", err)
			return nil, err
		}

		err = transaction.SetAmountMajor(cmd.Balance, user.CurrencyCode)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to set amount major", err)
			return nil, err
		}

		err = u.transactionsRepo.Save(ctx, transaction)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to save transaction", err)
			return nil, err
		}
	}

	return account, nil
}
