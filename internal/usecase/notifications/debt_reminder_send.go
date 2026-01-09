// internal/usecase/notifications/debt_reminder_send_usecase.go
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

type debtReminderSendUsecase struct {
	contextTimeout     time.Duration
	logger             *logger.Logger
	userRepo           entities.UserRepository
	debtsRepo          entities.DebtRepository
	telegramBotService ports.TelegramBotService
}

func NewDebtReminderSendUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	userRepo entities.UserRepository,
	debtsRepo entities.DebtRepository,
	telegramBotService ports.TelegramBotService,
) *debtReminderSendUsecase {
	return &debtReminderSendUsecase{
		contextTimeout:     timeout,
		logger:             logger,
		userRepo:           userRepo,
		debtsRepo:          debtsRepo,
		telegramBotService: telegramBotService,
	}
}

func (r *debtReminderSendUsecase) DebtReminderSend(ctx context.Context, debtID string) error {
	ctx, cancel := context.WithTimeout(ctx, r.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("notifications"), "DebtReminderSend",
		attribute.String("debt_id", debtID),
	)
	defer func() { end(nil) }()

	parsedDebtID, err := uuid.Parse(debtID)
	if err != nil {
		return inerr.NewErrValidation("debt_id", err.Error())
	}

	// Get debt details
	debt, err := r.debtsRepo.GetByID(ctx, parsedDebtID)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to get debt", err)
		return err
	}

	// Only send reminders for open debts
	if debt.Status != entities.Open {
		otlp.Event(ctx, "reminder_skipped", attribute.String("reason", "debt_not_open"))
		return nil
	}

	// Get user details
	user, err := r.userRepo.FindByID(ctx, debt.UserID)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to get user", err)
		return err
	}

	lang := resolveLang(user)
	loc := resolveLocation(user)

	text := debtReminderText(lang, DebtMsgParams{
		DebtType: debt.Type,
		Name:     debt.Name,
		Amount:   debt.AmountMajor(),
		Currency: debt.CurrencyCode.String(),
		RemindAt: debt.DueAt,
		Loc:      loc,
	})

	// Inline keyboard actions: Paid / Remind later / Cancel
	replyMarkup := buildDebtInlineKeyboard(lang, debtID)

	// Send notification
	if err := r.telegramBotService.SendMessage(ctx, &ports.SendMessageRequest{
		UserID:      user.TGUserID,
		Text:        text,
		ParseMode:   "HTML",
		ReplyMarkup: replyMarkup,
	}); err != nil {
		r.logger.ErrorContext(ctx, "failed to send telegram message", err)
		return err
	}

	otlp.Event(ctx, "debt_reminder_sent", attribute.String("debt_id", debtID))
	return nil
}
