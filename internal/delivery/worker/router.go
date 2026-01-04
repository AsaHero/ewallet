package worker

import (
	"github.com/AsaHero/e-wallet/internal/delivery"
	"github.com/AsaHero/e-wallet/internal/delivery/worker/handlers"
	"github.com/AsaHero/e-wallet/internal/tasks"
	"github.com/hibiken/asynq"
)

func NewRouter(opts *delivery.Options) *asynq.ServeMux {
	handler := handlers.Handler{
		NotificationUsecase: opts.NotificationUsecase,
	}

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.RecordReminderScheduleTaskName, handler.RecordReminderSchedule)
	mux.HandleFunc(tasks.RecordReminderSendTaskName, handler.RecordReminderSend)
	mux.HandleFunc(tasks.DebtReminderCheckTaskName, handler.DebtReminderCheck)
	mux.HandleFunc(tasks.DebtReminderSendTaskName, handler.DebtReminderSend)

	return mux
}
