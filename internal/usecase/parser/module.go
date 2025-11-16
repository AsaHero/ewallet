package parser

import (
	"time"

	"github.com/AsaHero/e-wallet/internal/usecase/ports"
	"github.com/AsaHero/e-wallet/pkg/logger"
)

type Command struct {
	*parseTextUsecase
}

type Module struct {
	Command Command
}

func NewModule(timeout time.Duration, logger *logger.Logger, llmClient ports.LLMProvider) *Module {
	return &Module{
		Command: Command{
			parseTextUsecase: NewParseTextUsecase(timeout, logger, llmClient),
		},
	}
}
