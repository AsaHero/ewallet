package categories

import (
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/usecase/categories/query"
	"github.com/AsaHero/e-wallet/pkg/logger"
)

type Command struct{}
type Query struct {
	*query.GetAllCategoriesUsecase
	*query.GetAllSubcategoriesUsecase
}
type Module struct {
	Command Command
	Query   Query
}

func NewModule(timeout time.Duration, logger *logger.Logger, categoriesRepo entities.CategoryRepository, subcategoriesRepo entities.SubcategoryRepository, usersRepo entities.UserRepository) *Module {
	m := &Module{
		Query: Query{
			GetAllCategoriesUsecase:    query.NewGetAllCategoriesUsecase(timeout, logger, usersRepo, categoriesRepo),
			GetAllSubcategoriesUsecase: query.NewGetAllSubcategoriesUsecase(timeout, logger, usersRepo, subcategoriesRepo),
		},
	}

	return m
}
