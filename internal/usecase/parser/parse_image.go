package parser

import (
	"context"
	"encoding/json"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/inerr"
	"github.com/AsaHero/e-wallet/internal/usecase/ports"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
	"github.com/shogo82148/pointer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type parseImageUsecase struct {
	contextTimeout  time.Duration
	logger          *logger.Logger
	llmClient       ports.LLMProvider
	ocrProvider     ports.OCRProvider
	usersRepo       entities.UserRepository
	accountsRepo    entities.AccountRepository
	fxRatesProvider ports.FXRatesProvider
}

func NewParseImageUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	llmClient ports.LLMProvider,
	ocrProvider ports.OCRProvider,
	usersRepo entities.UserRepository,
	accountsRepo entities.AccountRepository,
	fxRatesProvider ports.FXRatesProvider,
) *parseImageUsecase {
	return &parseImageUsecase{
		contextTimeout:  timeout,
		logger:          logger,
		llmClient:       llmClient,
		ocrProvider:     ocrProvider,
		usersRepo:       usersRepo,
		accountsRepo:    accountsRepo,
		fxRatesProvider: fxRatesProvider,
	}
}

type ParseImageView struct {
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

func (p *parseImageUsecase) ParseImage(ctx context.Context, userID string, imageURL string) (_ *ParseImageView, err error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("parser"), "ParseImage",
		attribute.String("image_url", imageURL),
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

	// Extract text from image using Vision API
	extractedText, err := p.ocrProvider.ImageToText(ctx, imageURL)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to extract text from image", err)
		return nil, err
	}

	// Generate human readable text from ocr output
	humanreadableText, err := p.llmClient.ChatCompletion(ctx, openai.GPT4o, "", NewOcrParserMessagePrompt(extractedText))
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to generate human readable text", err)
		return nil, err
	}

	userPayment := UserPayment{
		Language:    user.LanguageCode.String(),
		Currency:    user.CurrencyCode.String(),
		Timezone:    user.Timezone,
		PaymentText: humanreadableText,
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

	var result ParseImageView
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
