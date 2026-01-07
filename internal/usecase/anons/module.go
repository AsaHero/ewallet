package anons

import (
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/usecase/anons/command"
	"github.com/AsaHero/e-wallet/internal/usecase/anons/query"
	"github.com/AsaHero/e-wallet/internal/usecase/ports"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/hibiken/asynq"
)

type Commands struct {
	*command.CreateAnonUsecase
	*command.TriggerAnonBroadcastUsecase
	*command.AnonSendUsecase
}

type Query struct {
	*query.GetAnonsUsecase
}

type Module struct {
	Command Commands
	Query   Query
}

func NewModule(
	timeout time.Duration,
	logger *logger.Logger,
	userRepo entities.UserRepository,
	anonRepo entities.AnonRepository,
	taskQueue *asynq.Client,
	telegramBotService ports.TelegramBotService,
) *Module {
	return &Module{
		Command: Commands{
			CreateAnonUsecase: command.NewCreateAnonUsecase(
				timeout,
				logger,
				anonRepo,
			),
			TriggerAnonBroadcastUsecase: command.NewTriggerAnonBroadcastUsecase(
				timeout,
				logger,
				taskQueue,
				userRepo,
				anonRepo,
			),
			AnonSendUsecase: command.NewAnonSendUsecase(
				timeout,
				logger,
				userRepo,
				anonRepo,
				telegramBotService,
			),
		},
		Query: Query{
			GetAnonsUsecase: query.NewGetAnonsUsecase(
				timeout,
				logger,
				anonRepo,
			),
		},
	}
}
