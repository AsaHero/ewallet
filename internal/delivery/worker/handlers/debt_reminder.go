package handlers

import (
	"context"
	"encoding/json"

	"github.com/AsaHero/e-wallet/internal/tasks"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func (h *Handler) DebtReminderCheck(ctx context.Context, task *asynq.Task) error {
	ctx, end := otlp.Start(ctx, otel.Tracer("worker"), "DebtReminderCheck", attribute.String("task_type", task.Type()))
	defer func() { end(nil) }()

	err := h.NotificationUsecase.DebtReminderCheck(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) DebtReminderSend(ctx context.Context, task *asynq.Task) error {
	ctx, end := otlp.Start(ctx, otel.Tracer("worker"), "DebtReminderSend", attribute.String("task_type", task.Type()))
	defer func() { end(nil) }()

	var payload tasks.DebtReminderSendPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	err := h.NotificationUsecase.DebtReminderSend(ctx, payload.DebtID)
	if err != nil {
		return err
	}

	return nil
}
