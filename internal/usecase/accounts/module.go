package accounts

import (
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/usecase/accounts/command"
	"github.com/AsaHero/e-wallet/internal/usecase/accounts/query"

	"github.com/AsaHero/e-wallet/pkg/logger"
)

type Commands struct {
	*command.CreateAccountUsecase
	*command.UpdateAccountUsecase
	*command.DeleteAccountUsecase
}

type Query struct {
	*query.GetAccountsByUserIDUsecase
}

type Module struct {
	Command Commands
	Query   Query
}

func NewModule(
	timeout time.Duration,
	logger *logger.Logger,
	usersRepo entities.UserRepository,
	accountsRepo entities.AccountRepository,
) *Module {
	m := &Module{
		Command: Commands{
			CreateAccountUsecase: command.NewCreateAccountUsecase(timeout, logger, usersRepo, accountsRepo),
			UpdateAccountUsecase: command.NewUpdateAccountUsecase(timeout, logger, usersRepo, accountsRepo),
			DeleteAccountUsecase: command.NewDeleteAccountUsecase(timeout, logger, usersRepo, accountsRepo),
		},
		Query: Query{
			GetAccountsByUserIDUsecase: query.NewGetAccountsByUserIDUsecase(timeout, logger, accountsRepo),
		},
	}

	return m
}
