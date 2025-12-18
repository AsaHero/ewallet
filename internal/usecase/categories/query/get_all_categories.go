package query

import (
	"context"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/inerr"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
)

type GetAllCategoriesUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	usersRepo      entities.UserRepository
	categoriesRepo entities.CategoryRepository
}

func NewGetAllCategoriesUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	usersRepo entities.UserRepository,
	categoriesRepo entities.CategoryRepository,
) *GetAllCategoriesUsecase {
	return &GetAllCategoriesUsecase{
		contextTimeout: timeout,
		logger:         logger,
		usersRepo:      usersRepo,
		categoriesRepo: categoriesRepo,
	}
}

// Category represents a transaction category
type Category struct {
	ID       int    `json:"id"`
	UserID   string `json:"user_id"`
	Position int    `json:"position"`
	Name     string `json:"name"`
	Emoji    string `json:"emoji"`
}

func (u *GetAllCategoriesUsecase) GetAllCategories(ctx context.Context, userID string) (_ []Category, err error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("categories"), "GetAll")
	defer func() { end(err) }()

	uuid, err := uuid.Parse(userID)
	if err != nil {
		return nil, inerr.NewErrValidation("user_id", "invalid user id")
	}

	user, err := u.usersRepo.FindByID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	list, err := u.categoriesRepo.FindAll(ctx, uuid)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get categories", err)
		return nil, err
	}

	var categories []Category
	for _, category := range list {
		categories = append(categories, Category{
			ID:       category.ID.Int(),
			UserID:   user.ID.String(),
			Position: category.Position,
			Name:     category.GetName(user.LanguageCode),
			Emoji:    category.Emoji,
		})
	}

	return categories, nil
}
