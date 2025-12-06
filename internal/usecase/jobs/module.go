package jobs

import (
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/hibiken/asynq"
)

type Module struct {
	*recordReminderCalculateSchedulerUsecase
}

func NewModule(timeout time.Duration, logger *logger.Logger, usersRepo entities.UserRepository, taskQueue *asynq.Client) *Module {
	return &Module{
		recordReminderCalculateSchedulerUsecase: NewRecordReminderCalculateSchedulerUsecase(timeout, logger, usersRepo, taskQueue),
	}
}
