package parser

import (
	"time"

	"github.com/AsaHero/e-wallet/internal/usecase/ports"
	"github.com/AsaHero/e-wallet/pkg/logger"
)

type Command struct {
	*parseTextUsecase
	*parseAudioUsecase
	*parseImageUsecase
}

type Module struct {
	Command Command
}

func NewModule(timeout time.Duration, logger *logger.Logger, llmClient ports.LLMProvider, ocrProvider ports.OCRProvider) *Module {
	return &Module{
		Command: Command{
			parseTextUsecase:  NewParseTextUsecase(timeout, logger, llmClient),
			parseAudioUsecase: NewParseAudioUsecase(timeout, logger, llmClient),
			parseImageUsecase: NewParseImageUsecase(timeout, logger, llmClient, ocrProvider),
		},
	}
}
