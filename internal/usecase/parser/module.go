package parser

import (
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
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

func NewModule(
	logger *logger.Logger,
	llmClient ports.LLMProvider,
	ocrProvider ports.OCRProvider,
	usersRepo entities.UserRepository,
	accountsRepo entities.AccountRepository,
	categoriesRepo entities.CategoryRepository,
	subcategoriesRepo entities.SubcategoryRepository,
	fxRatesProvider ports.FXRatesProvider,
) *Module {
	return &Module{
		Command: Command{
			parseTextUsecase:  NewParseTextUsecase(2*time.Minute, logger, llmClient, usersRepo, accountsRepo, categoriesRepo, subcategoriesRepo, fxRatesProvider),
			parseAudioUsecase: NewParseAudioUsecase(2*time.Minute, logger, llmClient, usersRepo, accountsRepo, categoriesRepo, subcategoriesRepo, fxRatesProvider),
			parseImageUsecase: NewParseImageUsecase(2*time.Minute, logger, llmClient, ocrProvider, usersRepo, accountsRepo, categoriesRepo, subcategoriesRepo, fxRatesProvider),
		},
	}
}
