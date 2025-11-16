package query

import (
	"context"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type GetByTGUserIDUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	usersRepo      entities.UserRepository
}

func NewGetByTGUserIDUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	usersRepo entities.UserRepository,
) *GetByTGUserIDUsecase {
	return &GetByTGUserIDUsecase{
		contextTimeout: timeout,
		usersRepo:      usersRepo,
		logger:         logger,
	}
}

func (g *GetByTGUserIDUsecase) GetByTGUserID(ctx context.Context, tgUserID int64) (_ *entities.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, g.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("users"), "GetByTGUserID",
		attribute.Int64("tg_user_id", tgUserID),
	)
	defer func() { end(err) }()

	user, err := g.usersRepo.FindByTGUserID(ctx, tgUserID)
	if err != nil {
		g.logger.ErrorContext(ctx, "failed to get user", err)
		return nil, err
	}

	return user, nil
}
