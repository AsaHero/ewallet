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

type DeleteAccountUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	usersRepo      entities.UserRepository
	accountsRepo   entities.AccountRepository
}

func NewDeleteAccountUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	usersRepo entities.UserRepository,
	accountsRepo entities.AccountRepository,
) *DeleteAccountUsecase {
	return &DeleteAccountUsecase{
		contextTimeout: timeout,
		usersRepo:      usersRepo,
		accountsRepo:   accountsRepo,
		logger:         logger,
	}
}

type DeleteAccountCommand struct {
	UserID    string
	AccountID string
}

func (u *DeleteAccountUsecase) DeleteAccount(ctx context.Context, cmd *DeleteAccountCommand) (err error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("accounts"), "DeleteAccount",
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
			return inerr.NewErrValidation("user_id", "invalud uuid type")
		}

		input.accountID, err = uuid.Parse(cmd.AccountID)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to parse account id", err)
			return inerr.NewErrValidation("account_id", "invalud uuid type")
		}
	}

	_, err = u.usersRepo.FindByID(ctx, input.userID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get user", err)
		return err
	}

	account, err := u.accountsRepo.GetByID(ctx, input.accountID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get account", err)
		return err
	}

	err = u.accountsRepo.Delete(ctx, account)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to detele account", err)
		return err
	}

	return nil
}
