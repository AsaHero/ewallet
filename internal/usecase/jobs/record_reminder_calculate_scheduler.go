package jobs

import (
	"context"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/tasks"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type recordReminderCalculateSchedulerUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	usersRepo      entities.UserRepository
	taskQueue      *asynq.Client
}

func NewRecordReminderCalculateSchedulerUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	usersRepo entities.UserRepository,
	taskQueue *asynq.Client,
) *recordReminderCalculateSchedulerUsecase {
	return &recordReminderCalculateSchedulerUsecase{
		contextTimeout: timeout,
		logger:         logger,
		usersRepo:      usersRepo,
		taskQueue:      taskQueue,
	}
}

func (r *recordReminderCalculateSchedulerUsecase) RecordReminderCalculateScheduler(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, r.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("jobs"), "RecordReminderCalculateScheduler")
	defer func() { end(nil) }()

	users, err := r.usersRepo.FindAll(ctx)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to get users", err)
		return err
	}

	// Store some stats to send to otlp
	var totalUsers int = len(users)
	var totalTasksCreated int = 0
	for _, user := range users {
		task, err := tasks.NewRecordReminderCalculateTask(user.ID.String())
		if err != nil {
			r.logger.ErrorContext(ctx, "failed to create task", err)
			return err
		}

		if _, err := r.taskQueue.Enqueue(task); err != nil {
			r.logger.ErrorContext(ctx, "failed to enqueue task", err)
			return err
		}
		totalTasksCreated++
	}

	otlp.Annotate(ctx,
		attribute.Int("total_users", totalUsers),
		attribute.Int("total_tasks_created", totalTasksCreated))

	return nil
}
