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

	return &DebtReminderScheduler{
		config:    cfg,
		scheduler: scheduler,
	}, nil
}

func (a *DebtReminderScheduler) Run() error {
	// Check for due debt reminders every 15 minutes
	task, err := tasks.NewDebtReminderCheckTask()
	if err != nil {
		return err
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
