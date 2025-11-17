package users

import (
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/usecase/users/command"
	"github.com/AsaHero/e-wallet/internal/usecase/users/query"
	"github.com/AsaHero/e-wallet/pkg/logger"
)

type Commands struct {
	*command.AuthTelegramUsecase
	*command.UpdateUsecase
}

type Query struct {
	*query.GetByTGUserIDUsecase
	*query.GetByIDUsecase
}

type Module struct {
	Command Commands
	Query   Query
}

func NewModule(
	timeout time.Duration,
	logger *logger.Logger,
	usersRepo entities.UserRepository,
) *Module {
	m := &Module{
		Command: Commands{
			AuthTelegramUsecase: command.NewAuthTelegramUsecase(timeout, logger, usersRepo),
			UpdateUsecase:       command.NewUpdateUsecase(timeout, logger, usersRepo),
		},
		Query: Query{
			GetByIDUsecase:       query.NewGetByUserIDUsecase(timeout, logger, usersRepo),
			GetByTGUserIDUsecase: query.NewGetByTGUserIDUsecase(timeout, logger, usersRepo),
		},
	}

	return m
}
