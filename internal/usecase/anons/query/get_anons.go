package query

import (
	"context"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"go.opentelemetry.io/otel"
)

type GetAnonsUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	anonRepo       entities.AnonRepository
}

func NewGetAnonsUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	anonRepo entities.AnonRepository,
) *GetAnonsUsecase {
	return &GetAnonsUsecase{
		contextTimeout: timeout,
		logger:         logger,
		anonRepo:       anonRepo,
	}
}

type GetAnonsQuery struct {
	// Add pagination or filters if needed later
}

func (u *GetAnonsUsecase) GetAnons(ctx context.Context, query *GetAnonsQuery) ([]*entities.Anon, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("notifications"), "GetAnons")
	defer func() { end(nil) }()

	anons, err := u.anonRepo.FindAll(ctx)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to list anons", err)
		return nil, err
	}

	return anons, nil
}
