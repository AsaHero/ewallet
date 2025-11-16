package parser

import (
	"context"
	"encoding/json"
	"time"

	"github.com/AsaHero/e-wallet/internal/usecase/ports"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/sashabaranov/go-openai"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type parseTextUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	llmClient      ports.LLMProvider
}

func NewParseTextUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	llmClient ports.LLMProvider,
) *parseTextUsecase {
	return &parseTextUsecase{
		contextTimeout: timeout,
		logger:         logger,
		llmClient:      llmClient,
	}
}

type ParseTextView struct {
	Type        string     `json:"type"`
	Amount      float64    `json:"amount"`
	CategoryID  *int       `json:"category_id,omitempty"`
	Note        string     `json:"note,omitempty"`
	Confidence  float64    `json:"confidence"`
	PerformedAt *time.Time `json:"performed_at,omitempty"`
}

func (p *parseTextUsecase) ParseText(ctx context.Context, text string) (_ *ParseTextView, err error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("parser"), "ParseText",
		attribute.String("text", text),
	)
	defer func() { end(err) }()

	response, err := p.llmClient.ChatCompletion(ctx, openai.GPT4o, ParserSystemMessage, text)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to parse text", err)
		return nil, err
	}

	var result ParseTextView
	err = json.Unmarshal([]byte(response), &result)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to parse text", err)
		return nil, err
	}

	return &result, nil
}
