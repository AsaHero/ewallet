package notifications

import (
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/usecase/ports"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/hibiken/asynq"
)

type Module struct {
	*recordReminderCalculateUsecase
	*recordReminderSendUsecase
}

func NewModule(
	logger *logger.Logger,
	transactionRepo entities.TransactionRepository,
	userRepo entities.UserRepository,
	taskQueue *asynq.Client,
	telegramBotService ports.TelegramBotService,
) *Module {
	return &Module{
		recordReminderCalculateUsecase: NewRecordReminderCalculateUsecase(5*time.Minute, logger, transactionRepo, userRepo, taskQueue),
		recordReminderSendUsecase:      NewRecordReminderSendUsecase(30*time.Second, logger, userRepo, transactionRepo, telegramBotService),
	}
}
