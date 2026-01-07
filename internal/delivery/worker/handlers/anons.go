package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AsaHero/e-wallet/internal/tasks"
	"github.com/hibiken/asynq"
)

func (h *Handler) AnonsSend(ctx context.Context, t *asynq.Task) error {
	var payload tasks.AnonsSendPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	return h.AnonsUsecase.Command.AnonsSend(ctx, payload.UserID, payload.AnonID)
}
