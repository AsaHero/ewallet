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

func NewModule(logger *logger.Logger, llmClient ports.LLMProvider, ocrProvider ports.OCRProvider) *Module {
	return &Module{
		Command: Command{
			parseTextUsecase:  NewParseTextUsecase(1*time.Minute, logger, llmClient),
			parseAudioUsecase: NewParseAudioUsecase(1*time.Minute, logger, llmClient),
			parseImageUsecase: NewParseImageUsecase(1*time.Minute, logger, llmClient, ocrProvider),
		},
	}
}
