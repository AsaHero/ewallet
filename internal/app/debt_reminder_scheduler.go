package app

import (
	"log"

	"github.com/AsaHero/e-wallet/internal/tasks"
	"github.com/AsaHero/e-wallet/pkg/config"
	"github.com/hibiken/asynq"
)

type DebtReminderScheduler struct {
	config    *config.Config
	scheduler *asynq.Scheduler
	taskQueue *asynq.Client
}

func NewDebtReminderScheduler(cfg *config.Config) (*DebtReminderScheduler, error) {
	scheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{
			Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
			Password: cfg.Redis.Password,
		},
		&asynq.SchedulerOpts{
			Location: nil,
		},
	)

	taskQueue := asynq.NewClient(
		asynq.RedisClientOpt{
			Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
			Password: cfg.Redis.Password,
		},
	)

	return &DebtReminderScheduler{
		config:    cfg,
		scheduler: scheduler,
		taskQueue: taskQueue,
	}, nil
}

func (a *DebtReminderScheduler) Run(runNow bool) error {
	// Check for due debt reminders every 15 minutes
	task, err := tasks.NewDebtReminderCheckTask()
	if err != nil {
		return err
	}

	if runNow {
		_, err = a.taskQueue.Enqueue(task, asynq.Queue("high"))
		if err != nil {
			return err
		}

		return nil
	}

	_, err = a.scheduler.Register("*/15 * * * *", task, asynq.Queue("high"))
	if err != nil {
		return err
	}

	log.Println("Starting debt reminder scheduler...")
	log.Println("Checking for due debt reminders every 15 minutes")

	return a.scheduler.Run()
}

func (a *DebtReminderScheduler) Stop() error {
	if a.scheduler != nil {
		a.scheduler.Shutdown()
	}

	return nil
}
