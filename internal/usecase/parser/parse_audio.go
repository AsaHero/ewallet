package parser

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/inerr"
	"github.com/AsaHero/e-wallet/internal/usecase/ports"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/AsaHero/e-wallet/pkg/utils"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
	"github.com/shogo82148/pointer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type parseAudioUsecase struct {
	contextTimeout  time.Duration
	logger          *logger.Logger
	llmClient       ports.LLMProvider
	usersRepo       entities.UserRepository
	accountsRepo    entities.AccountRepository
	fxRatesProvider ports.FXRatesProvider
}

func NewParseAudioUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	llmClient ports.LLMProvider,
	usersRepo entities.UserRepository,
	accountsRepo entities.AccountRepository,
	fxRatesProvider ports.FXRatesProvider,
) *parseAudioUsecase {
	return &parseAudioUsecase{
		contextTimeout:  timeout,
		logger:          logger,
		llmClient:       llmClient,
		usersRepo:       usersRepo,
		accountsRepo:    accountsRepo,
		fxRatesProvider: fxRatesProvider,
	}
}

type ParseAudioView struct {
	Type             string     `json:"type"`
	Amount           float64    `json:"amount"`
	Currency         string     `json:"currency,omitempty"`
	OriginalAmount   *float64   `json:"original_amount,omitempty"`
	OriginalCurrency *string    `json:"original_currency,omitempty"`
	FxRate           *float64   `json:"fx_rate,omitempty"`
	AccountID        *string    `json:"account_id,omitempty"`
	CategoryID       *int       `json:"category_id,omitempty"`
	Note             string     `json:"note,omitempty"`
	Confidence       float64    `json:"confidence"`
	PerformedAt      *time.Time `json:"performed_at,omitempty"`
}

func (p *parseAudioUsecase) ParseAudio(ctx context.Context, userID string, fileURL string) (_ *ParseAudioView, err error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("parser"), "ParseAudio",
		attribute.String("file_url", fileURL),
	)
	defer func() { end(err) }()

	var input struct {
		userID uuid.UUID
	}
	{
		var err error
		input.userID, err = uuid.Parse(userID)
		if err != nil {
			p.logger.ErrorContext(ctx, "failed to parse user id", err)
			return nil, inerr.NewErrValidation("user_id", "invalud uuid type")
		}
	}

	user, err := p.usersRepo.FindByID(ctx, input.userID)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to get user", err)
		return nil, err
	}

	accounts, err := p.accountsRepo.GetByUserID(ctx, input.userID)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to get accounts", err)
		return nil, err
	}

	resp, err := http.Get(fileURL)
	if err != nil {
		p.logger.ErrorContext(ctx, "Error downloading file", err)
		return
	}
	defer resp.Body.Close()

	// Read file content
	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		p.logger.ErrorContext(ctx, "Error reading file", err)
		return
	}

	// Create temporary file for Whisper API
	tmpFile, err := os.CreateTemp("", "voice-*.oga")
	if err != nil {
		p.logger.ErrorContext(ctx, "Error creating temp file", err)
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err = tmpFile.Write(audioData); err != nil {
		p.logger.ErrorContext(ctx, "Error writing temp file", err)
		return
	}

	// 3) Convert ogg -> mp3 using ffmpeg
	mp3Path, err := utils.ConvertOggToMp3(ctx, tmpFile.Name())
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to convert ogg to mp3", err)
		return nil, err
	}
	defer os.Remove(mp3Path)

	transcriprionText, err := p.llmClient.AudioToText(ctx, mp3Path, user.LanguageCode.String())
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to transcribe audio", err)
		return nil, err
	}

	userPayment := UserPayment{
		Language:    user.LanguageCode.String(),
		Currency:    user.CurrencyCode.String(),
		Timezone:    user.Timezone,
		PaymentText: transcriprionText,
	}

	for _, account := range accounts {
		userPayment.Accounts = append(userPayment.Accounts, UserPaymentAccount{
			ID:   account.ID.String(),
			Name: account.Name,
		})
	}

	userPrompt := NewUserPaymentMessagePrompt(userPayment)
	response, err := p.llmClient.ChatCompletion(ctx, openai.GPT4o, ParserSystemMessage, userPrompt)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to parse text", err)
		return nil, err
	}

	// Clean from starting and ending ``` blocks
	response = utils.CleanMarkdownJSON(response)

	var result ParseAudioView
	err = json.Unmarshal([]byte(response), &result)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to parse text", err)
		return nil, err
	}

	if result.OriginalCurrency != nil {
		fxRate, err := p.fxRatesProvider.GetRate(ctx, *result.OriginalCurrency, user.CurrencyCode.String())
		if err != nil {
			p.logger.ErrorContext(ctx, "failed to get fx rate", err)
			return nil, err
		}
		result.OriginalAmount = pointer.Float64(result.Amount)
		result.Currency = user.CurrencyCode.String()
		result.Amount = result.Amount * fxRate
		result.FxRate = pointer.Float64(fxRate)
	} else {
		result.Currency = user.CurrencyCode.String()
	}

	return &result, nil
}
