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

type parseImageUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	llmClient      ports.LLMProvider
	ocrProvider    ports.OCRProvider
}

func NewParseImageUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	llmClient ports.LLMProvider,
	ocrProvider ports.OCRProvider,
) *parseImageUsecase {
	return &parseImageUsecase{
		contextTimeout: timeout,
		logger:         logger,
		llmClient:      llmClient,
		ocrProvider:    ocrProvider,
	}
}

type ParseImageView struct {
	Type        string  `json:"type"`
	Amount      float64 `json:"amount"`
	CategoryID  *int    `json:"category_id,omitempty"`
	Note        string  `json:"note,omitempty"`
	Confidence  float64 `json:"confidence"`
	PerformedAt *string `json:"performed_at,omitempty"`
}

func (p *parseImageUsecase) ParseImage(ctx context.Context, imageURL string) (_ *ParseImageView, err error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("parser"), "ParseImage",
		attribute.String("image_url", imageURL),
	)
	defer func() { end(err) }()

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

	// Parse the extracted text using ChatCompletion
	response, err := p.llmClient.ChatCompletion(ctx, openai.GPT4o, ParserSystemMessage, humanreadableText)
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

	return &result, nil
}
