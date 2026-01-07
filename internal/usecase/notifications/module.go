package notifications

import (
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/usecase/ports"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/hibiken/asynq"
)

type Module struct {
	*recordReminderScheduleUsecase
	*recordReminderSendUsecase
	*debtReminderCheckUsecase
	*debtReminderSendUsecase
}

func NewModule(
	logger *logger.Logger,
	transactionRepo entities.TransactionRepository,
	userRepo entities.UserRepository,
	debtsRepo entities.DebtRepository,
	anonRepo entities.AnonRepository,
	taskQueue *asynq.Client,
	telegramBotService ports.TelegramBotService,
) *Module {
	return &Module{
		recordReminderScheduleUsecase: NewRecordReminderScheduleUsecase(5*time.Minute, logger, userRepo, taskQueue),
		recordReminderSendUsecase:     NewRecordReminderSendUsecase(30*time.Second, logger, userRepo, transactionRepo, telegramBotService),
		debtReminderCheckUsecase:      NewDebtReminderCheckUsecase(5*time.Minute, logger, debtsRepo, taskQueue),
		debtReminderSendUsecase:       NewDebtReminderSendUsecase(30*time.Second, logger, userRepo, debtsRepo, telegramBotService),
	}
}
