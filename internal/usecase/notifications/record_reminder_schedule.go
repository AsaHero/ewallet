// internal/usecase/notifications/record_reminder_schedule_usecase.go
package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/tasks"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type recordReminderScheduleUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	userRepo       entities.UserRepository
	taskQueue      *asynq.Client
}

func NewRecordReminderScheduleUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	userRepo entities.UserRepository,
	taskQueue *asynq.Client,
) *recordReminderScheduleUsecase {
	return &recordReminderScheduleUsecase{
		contextTimeout: timeout,
		logger:         logger,
		userRepo:       userRepo,
		taskQueue:      taskQueue,
	}
}

func (r *recordReminderScheduleUsecase) RecordReminderSchedule(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, r.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("notifications"), "RecordReminderSchedule")
	defer func() { end(nil) }()

	users, err := r.userRepo.FindAll(ctx)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to get users", err)
		return err
	}

	r.logger.InfoContext(ctx, fmt.Sprintf("Scheduling reminders for %d users", len(users)))

	successCount := 0
	failureCount := 0

	for _, user := range users {
		lang := resolveLang(user)
		loc := resolveLocation(user)
		now := time.Now().In(loc)

		// 12:00 PM (noon)
		noon := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, loc)
		if noon.After(now) {
			delay := time.Until(noon)
			text := recordReminderText(lang, RecordNoon)
			if err := r.enqueueReminder(ctx, user.ID.String(), text, delay); err != nil {
				r.logger.ErrorContext(ctx, fmt.Sprintf("failed to schedule noon reminder for user %s", user.ID.String()), err)
				failureCount++
			} else {
				successCount++
			}
		}

		// 8:00 PM (evening)
		evening := time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, loc)
		if evening.After(now) {
			delay := time.Until(evening)
			text := recordReminderText(lang, RecordEvening)
			if err := r.enqueueReminder(ctx, user.ID.String(), text, delay); err != nil {
				r.logger.ErrorContext(ctx, fmt.Sprintf("failed to schedule evening reminder for user %s", user.ID.String()), err)
				failureCount++
			} else {
				successCount++
			}
		}
	}

	r.logger.InfoContext(ctx, fmt.Sprintf("Record reminder tasks scheduled: %d success, %d failed", successCount, failureCount))

	otlp.Event(ctx, "record_reminder_tasks_scheduled",
		attribute.Int("users", len(users)),
		attribute.Int("success", successCount),
		attribute.Int("failed", failureCount),
	)

	return nil
}

func (r *recordReminderScheduleUsecase) enqueueReminder(ctx context.Context, userID, text string, delay time.Duration) error {
	task, err := tasks.NewRecordReminderSendTask(userID, text)
	if err != nil {
		return err
	}

	_, err = r.taskQueue.Enqueue(task, asynq.Queue("medium"), asynq.ProcessIn(delay))
	if err != nil {
		return err
	}

	return nil
}
