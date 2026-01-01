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
	*query.GetTimeseriesStatsUsecase
	*query.GetStatsByCategoryUsecase
	*query.GetStatsBySubcategoryUsecase
	*query.GetStatsByAccountUsecase
	*query.GetStatsCompareUsecase
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
	accountsService *entities.AccountsService,
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
				accountsService,
				transactionsRepo,
				categortiesRepo,
				subcategoriesRepo,
			),
			DeleteTransactionUsecase: command.NewDeleteTransactionUsecase(
				timeout,
				logger,
				txManager,
				accountsRepo,
				accountsService,
				transactionsRepo,
			),
			UpdateTransactionUsecase: command.NewUpdateTransactionUsecase(
				timeout,
				logger,
				txManager,
				usersRepo,
				accountsRepo,
				accountsService,
				transactionsRepo,
				categortiesRepo,
				subcategoriesRepo,
			),
		},
		Query: Query{
			GetByIDUsecase:               query.NewGetByIDUsecase(timeout, logger, transactionsRepo),
			GetByFilterUsecase:           query.NewGetByFilterUsecase(timeout, logger, usersRepo, transactionsRepo),
			GetTimeseriesStatsUsecase:    query.NewGetTimeseriesStatsUsecase(timeout, logger, usersRepo, transactionsRepo),
			GetStatsByCategoryUsecase:    query.NewGetStatsByCategoryUsecase(timeout, logger, usersRepo, transactionsRepo, categortiesRepo),
			GetStatsBySubcategoryUsecase: query.NewGetStatsBySubcategoryUsecase(timeout, logger, usersRepo, transactionsRepo, subcategoriesRepo),
			GetStatsByAccountUsecase:     query.NewGetStatsByAccountUsecase(timeout, logger, usersRepo, transactionsRepo, accountsRepo),
			GetStatsCompareUsecase:       query.NewGetStatsCompareUsecase(timeout, logger, usersRepo, transactionsRepo, categortiesRepo),
		},
	}

	return m
}
