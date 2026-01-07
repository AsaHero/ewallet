package command

import (
	"context"
	"fmt"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/usecase/ports"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
)

type AnonSendUsecase struct {
	contextTimeout     time.Duration
	logger             *logger.Logger
	userRepo           entities.UserRepository
	anonRepo           entities.AnonRepository
	telegramBotService ports.TelegramBotService
}

func NewAnonSendUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	userRepo entities.UserRepository,
	anonRepo entities.AnonRepository,
	telegramBotService ports.TelegramBotService,
) *AnonSendUsecase {
	return &AnonSendUsecase{
		contextTimeout:     timeout,
		logger:             logger,
		userRepo:           userRepo,
		anonRepo:           anonRepo,
		telegramBotService: telegramBotService,
	}
}

func (u *AnonSendUsecase) AnonsSend(ctx context.Context, userID string, anonID string) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("notifications"), "AnonsSend")
	defer func() { end(nil) }()

	// Parse IDs
	uID, err := uuid.Parse(userID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to parse user ID", err)
		return err
	}

	aID, err := uuid.Parse(anonID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to parse anon ID", err)
		return err
	}

	// Get user
	user, err := u.userRepo.FindByID(ctx, uID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get user", err)
		return err
	}

	// Get anon broadcast content
	anon, err := u.anonRepo.FindByID(ctx, aID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get anon broadcast", err)
		return err
	}

	// Prepare and send message
	req := &ports.SendMessageRequest{
		UserID:    user.TGUserID,
		Text:      anon.Message,
		ParseMode: "HTML",
	}

	// Add video if present
	if anon.VideoFileID != "" {
		req.VideoID = anon.VideoFileID
	}

	// Add photo if present
	if anon.PhotoFileID != "" {
		req.PhotoID = anon.PhotoFileID
	}

	err = u.telegramBotService.SendMessage(ctx, req)
	if err != nil {
		u.logger.ErrorContext(ctx, fmt.Sprintf("failed to send anon broadcast to user %s", userID), err)
		return err
	}

	u.logger.InfoContext(ctx, fmt.Sprintf("anon broadcast sent to user %s", userID))

	return nil
}
