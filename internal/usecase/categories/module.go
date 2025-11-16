package categories

import (
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/usecase/categories/query"
	"github.com/AsaHero/e-wallet/pkg/logger"
)

type Command struct{}
type Query struct {
	*query.GetAllUsecase
}
type Module struct {
	Command Command
	Query   Query
}

func NewModule(timeout time.Duration, logger *logger.Logger, categoriesRepo entities.CategoryRepository) *Module {
	m := &Module{
		Query: Query{
			GetAllUsecase: query.NewGetAllUsecase(timeout, logger, categoriesRepo),
		},
	}

	return m
}
