package command

import (
	"context"
	"fmt"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/inerr"
	"github.com/AsaHero/e-wallet/internal/tasks"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type TriggerAnonBroadcastUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	taskQueue      *asynq.Client
	userRepo       entities.UserRepository
	anonRepo       entities.AnonRepository
}

func NewTriggerAnonBroadcastUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	taskQueue *asynq.Client,
	userRepo entities.UserRepository,
	anonRepo entities.AnonRepository,
) *TriggerAnonBroadcastUsecase {
	return &TriggerAnonBroadcastUsecase{
		contextTimeout: timeout,
		logger:         logger,
		taskQueue:      taskQueue,
		userRepo:       userRepo,
		anonRepo:       anonRepo,
	}
}

type TriggerAnonBroadcastCommand struct {
	AnonID        string
	UserIDs       []string
	LanguageCodes []string
}

func (u *TriggerAnonBroadcastUsecase) TriggerAnonBroadcast(ctx context.Context, cmd *TriggerAnonBroadcastCommand) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("notifications"), "TriggerAnonBroadcast")
	defer func() { end(nil) }()

	var input struct {
		anonID uuid.UUID
	}
	{
		var err error
		input.anonID, err = uuid.Parse(cmd.AnonID)
		if err != nil {
			u.logger.ErrorContext(ctx, "invalid anon id", err)
			return inerr.NewErrValidation("anon_id", "invalid uuid")
		}
	}

	// Check if anon exists
	_, err := u.anonRepo.FindByID(ctx, input.anonID)
	if err != nil {
		u.logger.ErrorContext(ctx, "anon broadcast not found", err)
		return err // Could map to Not Found error
	}

	// Get users by filter
	users, err := u.userRepo.FindByFilter(ctx, cmd.UserIDs, cmd.LanguageCodes)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get users by filter", err)
		return err
	}

	successCount := 0
	failureCount := 0

	// Create a task for each user
	for _, user := range users {
		task, err := tasks.NewAnonsSendTask(user.ID.String(), cmd.AnonID)
		if err != nil {
			u.logger.ErrorContext(ctx, fmt.Sprintf("failed to create task for user %s", user.ID.String()), err)
			failureCount++
			continue
		}

		_, err = u.taskQueue.Enqueue(task)
		if err != nil {
			u.logger.ErrorContext(ctx, fmt.Sprintf("failed to enqueue task for user %s", user.ID.String()), err)
			failureCount++
			continue
		}

		successCount++
	}

	u.logger.InfoContext(ctx, fmt.Sprintf("Anon broadcast send tasks enqueued: %d success, %d failed", successCount, failureCount))

	otlp.Event(ctx, "anon_broadcast_tasks_enqueued",
		attribute.Int("total", len(users)),
		attribute.Int("success", successCount),
		attribute.Int("failed", failureCount),
	)
	u.logger.InfoContext(ctx, "anon broadcast triggered and send tasks enqueued")

	return nil
}
