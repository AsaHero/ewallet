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

type UpdateUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	usersRepo      entities.UserRepository
}

func NewUpdateUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	usersRepo entities.UserRepository,
) *UpdateUsecase {
	return &UpdateUsecase{
		contextTimeout: timeout,
		usersRepo:      usersRepo,
		logger:         logger,
	}
}

type UpdateCommand struct {
	UserID       string
	LanguageCode *string
	CurrencyCode *string
}

func (u *UpdateUsecase) Update(ctx context.Context, cmd *UpdateCommand) (_ *entities.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("users"), "Update",
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

	if cmd.LanguageCode != nil {
		user.UpdateLanguageCode(entities.Language(*cmd.LanguageCode))
	}

	if cmd.CurrencyCode != nil {
		user.UpdateCurrencyCode(entities.Currency(*cmd.CurrencyCode))
	}

	err = u.usersRepo.Save(ctx, user)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to save user", err)
		return nil, err
	}

	return user, nil
}
