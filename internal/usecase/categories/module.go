package categories

import (
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/usecase/categories/command"
	"github.com/AsaHero/e-wallet/internal/usecase/categories/query"
	"github.com/AsaHero/e-wallet/pkg/logger"
)

type Command struct {
	*command.CreateCategoryUsecase
	*command.CreateSubcategoryUsecase
}
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
		Command: Command{
			CreateCategoryUsecase:    command.NewCreateCategoryUsecase(timeout, logger, usersRepo, categoriesRepo),
			CreateSubcategoryUsecase: command.NewCreateSubcategoryUsecase(timeout, logger, usersRepo, subcategoriesRepo),
		},
		Query: Query{
			GetAllCategoriesUsecase:    query.NewGetAllCategoriesUsecase(timeout, logger, usersRepo, categoriesRepo),
			GetAllSubcategoriesUsecase: query.NewGetAllSubcategoriesUsecase(timeout, logger, usersRepo, subcategoriesRepo),
		},
	}

	return m
}
