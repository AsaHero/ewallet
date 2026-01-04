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

type debtReminderCheckUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	debtsRepo      entities.DebtRepository
	taskQueue      *asynq.Client
}

func NewDebtReminderCheckUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	debtsRepo entities.DebtRepository,
	taskQueue *asynq.Client,
) *debtReminderCheckUsecase {
	return &debtReminderCheckUsecase{
		contextTimeout: timeout,
		logger:         logger,
		debtsRepo:      debtsRepo,
		taskQueue:      taskQueue,
	}
}

func (r *debtReminderCheckUsecase) DebtReminderCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, r.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("notifications"), "DebtReminderCheck")
	defer func() { end(nil) }()

	// Get all debts with reminders due before now
	debts, err := r.debtsRepo.GetDueReminders(ctx, time.Now())
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to get due reminders", err)
		return err
	}

	r.logger.InfoContext(ctx, fmt.Sprintf("Found %d due debt reminders, enqueuing tasks", len(debts)))

	successCount := 0
	failureCount := 0

	// Create a task for each debt reminder
	for _, debt := range debts {
		task, err := tasks.NewDebtReminderSendTask(debt.ID.String())
		if err != nil {
			r.logger.ErrorContext(ctx, fmt.Sprintf("failed to create task for debt %s", debt.ID.String()), err)
			failureCount++
			continue
		}

		_, err = r.taskQueue.Enqueue(task)
		if err != nil {
			r.logger.ErrorContext(ctx, fmt.Sprintf("failed to enqueue task for debt %s", debt.ID.String()), err)
			failureCount++
			continue
		}

		successCount++
	}

	r.logger.InfoContext(ctx, fmt.Sprintf("Debt reminder tasks enqueued: %d success, %d failed", successCount, failureCount))

	otlp.Event(ctx, "debt_reminder_tasks_enqueued",
		attribute.Int("total", len(debts)),
		attribute.Int("success", successCount),
		attribute.Int("failed", failureCount),
	)

	return nil
}
