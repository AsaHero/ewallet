package command

import (
	"context"
	"encoding/json"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/inerr"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"go.opentelemetry.io/otel"
)

type CreateAnonUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	anonRepo       entities.AnonRepository
}

func NewCreateAnonUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	anonRepo entities.AnonRepository,
) *CreateAnonUsecase {
	return &CreateAnonUsecase{
		contextTimeout: timeout,
		logger:         logger,
		anonRepo:       anonRepo,
	}
}

type CreateAnonCommand struct {
	LanguageCode string
	VideoFileID  string
	PhotoFileID  string
	Message      string
	ReplyMarkup  map[string]any
}

func (u *CreateAnonUsecase) CreateAnon(ctx context.Context, cmd *CreateAnonCommand) (*entities.Anon, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("notifications"), "CreateAnon")
	defer func() { end(nil) }()

	var input struct {
		languageCode entities.Language
		replyMarkup  []byte
	}
	{
		if cmd.Message == "" {
			return nil, inerr.NewErrValidation("message", "required")
		}

		input.languageCode = entities.Language(cmd.LanguageCode)
		if input.languageCode == "" {
			input.languageCode = "en"
		}

		if len(cmd.ReplyMarkup) > 0 {
			var err error
			input.replyMarkup, err = json.Marshal(cmd.ReplyMarkup)
			if err != nil {
				u.logger.ErrorContext(ctx, "failed to marshal reply markup", err)
				return nil, inerr.NewErrValidation("reply_markup", "invalid json")
			}
		}
	}

	// Create Anon entity using domain constructor
	anon, err := entities.NewAnon(cmd.Message, input.languageCode, input.replyMarkup)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to create anon entity", err)
		return nil, inerr.NewErrValidation("anon", err.Error())
	}

	// Handle optional fields using domain methods
	anon.SetVideoFileID(cmd.VideoFileID)
	anon.SetPhotoFileID(cmd.PhotoFileID)

	if err := u.anonRepo.Save(ctx, anon); err != nil {
		u.logger.ErrorContext(ctx, "failed to save anon broadcast", err)
		return nil, err
	}

	u.logger.InfoContext(ctx, "anon broadcast created", "id", anon.ID)

	return anon, nil
}
