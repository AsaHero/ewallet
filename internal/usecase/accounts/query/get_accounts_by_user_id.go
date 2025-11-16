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

type GetAccountsByUserIDUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	accountsRepo   entities.AccountRepository
}

func NewGetAccountsByUserIDUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	accountsRepo entities.AccountRepository,
) *GetAccountsByUserIDUsecase {
	return &GetAccountsByUserIDUsecase{
		contextTimeout: timeout,
		accountsRepo:   accountsRepo,
		logger:         logger,
	}
}

func (u *GetAccountsByUserIDUsecase) GetAccountsByUserID(ctx context.Context, userID string) (_ []*entities.Account, err error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("accounts"), "GetAccountsByUserID",
		attribute.String("user_id", userID),
	)
	defer func() { end(err) }()

	var input struct {
		userID uuid.UUID
	}
	{
		var err error
		input.userID, err = uuid.Parse(userID)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to parse user id", err)
			return nil, inerr.NewErrValidation("user_id", "invalud uuid type")
		}
	}

	accounts, err := u.accountsRepo.GetByUserID(ctx, input.userID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get accounts", err)
		return nil, err
	}

	return accounts, nil
}
