package parser

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/AsaHero/e-wallet/internal/usecase/ports"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/sashabaranov/go-openai"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type parseAudioUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	llmClient      ports.LLMProvider
}

func NewParseAudioUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	llmClient ports.LLMProvider,
) *parseAudioUsecase {
	return &parseAudioUsecase{
		contextTimeout: timeout,
		logger:         logger,
		llmClient:      llmClient,
	}
}

type ParseAudioView struct {
	Type        string     `json:"type"`
	Amount      float64    `json:"amount"`
	CategoryID  *int       `json:"category_id,omitempty"`
	Note        string     `json:"note,omitempty"`
	Confidence  float64    `json:"confidence"`
	PerformedAt *time.Time `json:"performed_at,omitempty"`
}

func (p *parseAudioUsecase) ParseAudio(ctx context.Context, fileURL string) (_ *ParseAudioView, err error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("parser"), "ParseAudio",
		attribute.String("file_url", fileURL),
	)
	defer func() { end(err) }()

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

	transcriprion, err := p.llmClient.AudioToText(ctx, tmpFile.Name())
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to transcribe audio", err)
		return nil, err
	}

	response, err := p.llmClient.ChatCompletion(ctx, openai.GPT4o, ParserSystemMessage, transcriprion)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to parse text", err)
		return nil, err
	}

	var result ParseAudioView
	err = json.Unmarshal([]byte(response), &result)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to parse text", err)
		return nil, err
	}

	return &result, nil
}
