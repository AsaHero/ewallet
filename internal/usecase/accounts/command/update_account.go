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

type UpdateAccountUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	usersRepo      entities.UserRepository
	accountsRepo   entities.AccountRepository
}

func NewUpdateAccountUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	usersRepo entities.UserRepository,
	accountsRepo entities.AccountRepository,
) *UpdateAccountUsecase {
	return &UpdateAccountUsecase{
		contextTimeout: timeout,
		usersRepo:      usersRepo,
		accountsRepo:   accountsRepo,
		logger:         logger,
	}
}

type UpdateAccounCommand struct {
	UserID    string
	AccountID string
	Name      *string
	IsDefault *bool
}

func (u *UpdateAccountUsecase) UpdateAccount(ctx context.Context, cmd *UpdateAccounCommand) (_ *entities.Account, err error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("accounts"), "UpdateAccount",
		attribute.String("user_id", cmd.UserID),
	)
	defer func() { end(err) }()

	var input struct {
		userID    uuid.UUID
		accountID uuid.UUID
	}
	{
		var err error
		input.userID, err = uuid.Parse(cmd.UserID)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to parse user id", err)
			return nil, inerr.NewErrValidation("user_id", "invalud uuid type")
		}

		input.accountID, err = uuid.Parse(cmd.AccountID)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to parse account id", err)
			return nil, inerr.NewErrValidation("account_id", "invalud uuid type")
		}
	}

	_, err = u.usersRepo.FindByID(ctx, input.userID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get user", err)
		return nil, err
	}

	account, err := u.accountsRepo.GetByID(ctx, input.accountID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get account", err)
		return nil, err
	}

	if cmd.Name != nil {
		account.UpdateName(*cmd.Name)
	}

	if cmd.IsDefault != nil {
		account.UpdateDefault(*cmd.IsDefault)
	}

	err = u.accountsRepo.Save(ctx, account)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to save account", err)
		return nil, err
	}

	return account, nil
}
