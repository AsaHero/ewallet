package query

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

type GetByIDUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	usersRepo      entities.UserRepository
}

func NewGetByUserIDUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	usersRepo entities.UserRepository,
) *GetByIDUsecase {
	return &GetByIDUsecase{
		contextTimeout: timeout,
		usersRepo:      usersRepo,
		logger:         logger,
	}
}

func (g *GetByIDUsecase) GetByID(ctx context.Context, id string) (_ *entities.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, g.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("users"), "GetByID",
		attribute.String("user_id", id),
	)
	defer func() { end(err) }()

	var input struct {
		userID uuid.UUID
	}
	{
		var err error
		input.userID, err = uuid.Parse(id)
		if err != nil {
			g.logger.ErrorContext(ctx, "failed to parse user id", err)
			return nil, inerr.NewErrValidation("user_id", "invalud uuid type")
		}
	}

	user, err := g.usersRepo.FindByID(ctx, input.userID)
	if err != nil {
		g.logger.ErrorContext(ctx, "failed to get user", err)
		return nil, err
	}

	return user, nil
}
