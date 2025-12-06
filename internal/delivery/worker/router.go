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
	mux.HandleFunc(tasks.RecordReminderCalculateTaskName, handler.RecordReminderCalculate)
	mux.HandleFunc(tasks.RecordReminderSendTaskName, handler.RecordReminderSend)

	return mux
}
