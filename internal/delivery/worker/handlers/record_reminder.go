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

func (h *Handler) RecordReminderSchedule(ctx context.Context, task *asynq.Task) error {
	ctx, end := otlp.Start(ctx, otel.Tracer("worker"), "RecordReminderSchedule", attribute.String("task_type", task.Type()))
	defer func() { end(nil) }()

	err := h.NotificationUsecase.RecordReminderSchedule(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) RecordReminderSend(ctx context.Context, task *asynq.Task) error {
	ctx, end := otlp.Start(ctx, otel.Tracer("worker"), "RecordReminderSend", attribute.String("task_type", task.Type()))
	defer func() { end(nil) }()

	var payload tasks.RecordReminderSendPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	err := h.NotificationUsecase.RecordReminderSend(ctx, payload.UserID, payload.Text)
	if err != nil {
		return err
	}

	return nil
}
