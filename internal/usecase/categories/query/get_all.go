package query

import (
	"context"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"go.opentelemetry.io/otel"
)

type GetAllUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	categoriesRepo entities.CategoryRepository
}

func NewGetAllUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	categoriesRepo entities.CategoryRepository,
) *GetAllUsecase {
	return &GetAllUsecase{
		contextTimeout: timeout,
		logger:         logger,
		categoriesRepo: categoriesRepo,
	}
}

func (u *GetAllUsecase) GetAll(ctx context.Context) (_ []*entities.Category, err error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("categories"), "GetAll")
	defer func() { end(err) }()

	list, err := u.categoriesRepo.FindAll(ctx)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get categories", err)
		return nil, err
	}

	return list, nil
}
