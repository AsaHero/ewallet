package transactions

import (
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/usecase/transactions/command"
	"github.com/AsaHero/e-wallet/internal/usecase/transactions/query"

	"github.com/AsaHero/e-wallet/pkg/database/postgres"
	"github.com/AsaHero/e-wallet/pkg/logger"
)

type Commands struct {
	*command.CreateTransactionUsecase
	*command.DeleteTransactionUsecase
	*command.UpdateTransactionUsecase
}

type Query struct {
	*query.GetByIDUsecase
	*query.GetByFilterUsecase
	*query.GetStatsUsecase
}

type Module struct {
	Command Commands
	Query   Query
}

func NewModule(
	timeout time.Duration,
	logger *logger.Logger,
	txManager postgres.TxManager,
	usersRepo entities.UserRepository,
	accountsRepo entities.AccountRepository,
	transactionsRepo entities.TransactionRepository,
	categortiesRepo entities.CategoryRepository,
	subcategoriesRepo entities.SubcategoryRepository,
) *Module {
	m := &Module{
		Command: Commands{
			CreateTransactionUsecase: command.NewCreateTransactionUsecase(
				timeout,
				logger,
				txManager,
				usersRepo,
				accountsRepo,
				transactionsRepo,
				categortiesRepo,
				subcategoriesRepo,
			),
			DeleteTransactionUsecase: command.NewDeleteTransactionUsecase(
				timeout,
				logger,
				txManager,
				accountsRepo,
				transactionsRepo,
			),
			UpdateTransactionUsecase: command.NewUpdateTransactionUsecase(
				timeout,
				logger,
				txManager,
				usersRepo,
				accountsRepo,
				transactionsRepo,
				categortiesRepo,
				subcategoriesRepo,
			),
		},
		Query: Query{
			GetByIDUsecase:     query.NewGetByIDUsecase(timeout, logger, transactionsRepo),
			GetByFilterUsecase: query.NewGetByFilterUsecase(timeout, logger, transactionsRepo),
			GetStatsUsecase:    query.NewGetStatsUsecase(timeout, logger, usersRepo, accountsRepo, transactionsRepo, categortiesRepo),
		},
	}

	return m
}
