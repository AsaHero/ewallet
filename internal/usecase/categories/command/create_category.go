package command

import (
	"context"
	"time"

	"github.com/AsaHero/e-wallet/internal/delivery/api/models"
	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/google/uuid"
	"github.com/shogo82148/pointer"
	"go.opentelemetry.io/otel"
)

type CreateCategoryUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	usersRepo      entities.UserRepository
	categoryRepo   entities.CategoryRepository
}

func NewCreateCategoryUsecase(timeout time.Duration, logger *logger.Logger, usersRepo entities.UserRepository, categoryRepo entities.CategoryRepository) *CreateCategoryUsecase {
	return &CreateCategoryUsecase{
		contextTimeout: timeout,
		logger:         logger,
		categoryRepo:   categoryRepo,
		usersRepo:      usersRepo,
	}
}

type CreateCategoryCommand struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Emoji  string `json:"emoji"`
}

func (c *CreateCategoryUsecase) CreateCategory(ctx context.Context, cmd *CreateCategoryCommand) (_ *models.Category, err error) {
	ctx, cancel := context.WithTimeout(ctx, c.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("categories"), "CreateCategory")
	defer func() { end(err) }()

	userID, err := uuid.Parse(cmd.UserID)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to parse user id", err)
		return nil, err
	}

	user, err := c.usersRepo.FindByID(ctx, userID)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to find user", err)
		return nil, err
	}

	category, err := entities.NewUserCategory(userID, cmd.Name, cmd.Emoji)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to create category", err)
		return nil, err
	}

	err = c.categoryRepo.Save(ctx, category)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to save category", err)
		return nil, err
	}

	return &models.Category{
		ID:        category.ID.Int(),
		UserID:    pointer.StringOrNil(user.ID.String()),
		Name:      category.GetName(user.LanguageCode),
		Emoji:     category.Emoji,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}, nil
}
