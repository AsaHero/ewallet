package debts

import (
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/usecase/debts/command"
	"github.com/AsaHero/e-wallet/internal/usecase/debts/query"
	"github.com/AsaHero/e-wallet/pkg/database/postgres"
	"github.com/AsaHero/e-wallet/pkg/logger"
)

type Commands struct {
	*command.CreateDebtUsecase
	*command.UpdateDebtUsecase
	*command.PayDebtUsecase
	*command.CancelDebtUsecase
}

type Query struct {
	*query.GetDebtByIDUsecase
	*query.GetDebtsByFilterUsecase
}

type Module struct {
	Command Commands
	Query   Query
}

func NewModule(
	timeout time.Duration,
	logger *logger.Logger,
	txManager postgres.TxManager,
	debtsRepo entities.DebtRepository,
	transactionsRepo entities.TransactionRepository,
) *Module {
	m := &Module{
		Command: Commands{
			CreateDebtUsecase: command.NewCreateDebtUsecase(
				timeout,
				logger,
				debtsRepo,
				transactionsRepo,
			),
			UpdateDebtUsecase: command.NewUpdateDebtUsecase(
				timeout,
				logger,
				txManager,
				debtsRepo,
			),
			PayDebtUsecase: command.NewPayDebtUsecase(
				timeout,
				logger,
				txManager,
				debtsRepo,
			),
			CancelDebtUsecase: command.NewCancelDebtUsecase(
				timeout,
				logger,
				txManager,
				debtsRepo,
			),
		},
		Query: Query{
			GetDebtByIDUsecase:      query.NewGetDebtByIDUsecase(timeout, logger, debtsRepo),
			GetDebtsByFilterUsecase: query.NewGetDebtsByFilterUsecase(timeout, logger, debtsRepo),
		},
	}

	return m
}
