package worker

import (
	"github.com/AsaHero/e-wallet/internal/delivery"
	"github.com/AsaHero/e-wallet/internal/delivery/worker/handlers"
	"github.com/AsaHero/e-wallet/internal/tasks"
	"github.com/hibiken/asynq"
)

func NewRouter(opts *delivery.Options) *asynq.Server {
	server := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     opts.Config.Redis.Host + ":" + opts.Config.Redis.Port,
			Password: opts.Config.Redis.Password,
		},
		asynq.Config{
			Concurrency: 100,
			Queues: map[string]int{
				"critical": 6,
				"medium":   3,
				"low":      1,
			},
		},
	)

	handler := handlers.Handler{
		NotificationUsecase: opts.NotificationUsecase,
	}

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.RecordReminderCalculateTaskName, handler.RecordReminderCalculate)
	mux.HandleFunc(tasks.RecordReminderSendTaskName, handler.RecordReminderSend)

	return server
}
