package command

import (
	"context"
	"errors"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/inerr"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type AuthTelegramUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	usersRepo      entities.UserRepository
}

func NewAuthTelegramUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	usersRepo entities.UserRepository,
) *AuthTelegramUsecase {
	return &AuthTelegramUsecase{
		contextTimeout: timeout,
		usersRepo:      usersRepo,
		logger:         logger,
	}
}

type AuthTelegramCommand struct {
	TelegramUserID int64
	FirstName      string
	LastName       string
	Username       string
}

func (u *AuthTelegramUsecase) AuthTelegram(ctx context.Context, cmd *AuthTelegramCommand) (_ *entities.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("users"), "AuthTelegram",
		attribute.Int64("tg_user_id", cmd.TelegramUserID),
	)
	defer func() { end(err) }()

	user, err := u.usersRepo.FindByTGUserID(ctx, cmd.TelegramUserID)
	if err != nil && !errors.Is(err, inerr.ErrNotFound{}) {
		u.logger.ErrorContext(ctx, "failed to find by tg user id", err)
		return nil, err
	}

	if user != nil {
		return user, nil
	}

	user, err = entities.NewUser(cmd.TelegramUserID, cmd.FirstName, cmd.LastName, cmd.Username)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to create user", err)
		return nil, err
	}

	err = u.usersRepo.Save(ctx, user)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to save user", err)
		return nil, err
	}

	return user, nil
}
