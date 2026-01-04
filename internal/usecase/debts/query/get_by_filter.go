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

type GetDebtsByFilterQuery struct {
	Limit          int      `form:"limit"`
	Offset         int      `form:"offset"`
	TransactionIDs []string `form:"transaction_ids"`
	Types          []string `form:"types"`
	Statuses       []string `form:"statuses"`
}

type GetDebtsByFilterUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	debtsRepo      entities.DebtRepository
}

func NewGetDebtsByFilterUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	debtsRepo entities.DebtRepository,
) *GetDebtsByFilterUsecase {
	return &GetDebtsByFilterUsecase{
		contextTimeout: timeout,
		debtsRepo:      debtsRepo,
		logger:         logger,
	}
}

func (u *GetDebtsByFilterUsecase) GetByFilter(ctx context.Context, userID string, query *GetDebtsByFilterQuery) (_ []*entities.Debt, _ int, err error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("debts"), "GetByFilter",
		attribute.String("user_id", userID),
	)
	defer func() { end(err) }()

	var input struct {
		userID uuid.UUID
		filter *entities.DebtFilter
	}
	{
		var err error
		input.userID, err = uuid.Parse(userID)
		if err != nil {
			u.logger.ErrorContext(ctx, "failed to parse user id", err)
			return nil, 0, inerr.NewErrValidation("user_id", "invalid uuid type")
		}

		input.filter = &entities.DebtFilter{
			UserID: input.userID,
			Limit:  query.Limit,
			Offset: query.Offset,
		}

		// Parse transaction IDs
		if len(query.TransactionIDs) > 0 {
			transactionIDs := make([]uuid.UUID, 0, len(query.TransactionIDs))
			for _, idStr := range query.TransactionIDs {
				id, err := uuid.Parse(idStr)
				if err != nil {
					u.logger.ErrorContext(ctx, "failed to parse transaction id", err)
					return nil, 0, inerr.NewErrValidation("transaction_ids", "invalid uuid type")
				}
				transactionIDs = append(transactionIDs, id)
			}
			input.filter.TransactionIDs = transactionIDs
		}

		// Parse types
		if len(query.Types) > 0 {
			types := make([]entities.DebtType, 0, len(query.Types))
			for _, typeStr := range query.Types {
				types = append(types, entities.DebtType(typeStr))
			}
			input.filter.Types = types
		}

		// Parse statuses
		if len(query.Statuses) > 0 {
			statuses := make([]entities.DebtStatus, 0, len(query.Statuses))
			for _, statusStr := range query.Statuses {
				statuses = append(statuses, entities.DebtStatus(statusStr))
			}
			input.filter.Statuses = statuses
		}
	}

	debts, count, err := u.debtsRepo.GetByFilter(ctx, input.filter)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get debts", err)
		return nil, 0, err
	}

	return debts, count, nil
}
