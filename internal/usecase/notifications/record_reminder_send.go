package notifications

import (
	"context"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/inerr"
	"github.com/AsaHero/e-wallet/internal/usecase/ports"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type recordReminderSendUsecase struct {
	contextTimeout     time.Duration
	logger             *logger.Logger
	userRepo           entities.UserRepository
	telegramBotService ports.TelegramBotService
}

func NewRecordReminderSendUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	userRepo entities.UserRepository,
	telegramBotService ports.TelegramBotService,
) *recordReminderSendUsecase {
	return &recordReminderSendUsecase{
		contextTimeout:     timeout,
		logger:             logger,
		userRepo:           userRepo,
		telegramBotService: telegramBotService,
	}
}

func (r *recordReminderSendUsecase) RecordReminderSend(ctx context.Context, userID string, text string) error {
	ctx, cancel := context.WithTimeout(ctx, r.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("notifications"), "RecordReminderSend",
		attribute.String("user_id", userID),
	)
	defer func() { end(nil) }()

	var input struct {
		userID uuid.UUID
	}
	{
		var err error

		input.userID, err = uuid.Parse(userID)
		if err != nil {
			return inerr.NewErrValidation("user_id", err.Error())
		}
	}

	user, err := r.userRepo.FindByID(ctx, input.userID)
	if err != nil {
		return err
	}

	if err := r.telegramBotService.SendMessage(ctx, &ports.SendMessageRequest{
		UserID:    user.TGUserID,
		Text:      text,
		ParseMode: "HTML",
	}); err != nil {
		return err
	}

	return nil
}
