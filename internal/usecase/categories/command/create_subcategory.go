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

type CreateSubcategoryUsecase struct {
	contextTimeout  time.Duration
	logger          *logger.Logger
	usersRepo       entities.UserRepository
	subcategoryRepo entities.SubcategoryRepository
}

func NewCreateSubcategoryUsecase(timeout time.Duration, logger *logger.Logger, usersRepo entities.UserRepository, subcategoryRepo entities.SubcategoryRepository) *CreateSubcategoryUsecase {
	return &CreateSubcategoryUsecase{
		contextTimeout:  timeout,
		logger:          logger,
		usersRepo:       usersRepo,
		subcategoryRepo: subcategoryRepo,
	}
}

type CreateSubcategoryCommand struct {
	UserID     string `json:"user_id"`
	CategoryID int    `json:"category_id"`
	Name       string `json:"name"`
	Emoji      string `json:"emoji"`
}

func (c *CreateSubcategoryUsecase) CreateSubcategory(ctx context.Context, cmd *CreateSubcategoryCommand) (_ *models.Subcategory, err error) {
	ctx, cancel := context.WithTimeout(ctx, c.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("categories"), "CreateSubcategory")
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

	subcategory, err := entities.NewUserSubcategory(cmd.CategoryID, userID, cmd.Name, cmd.Emoji)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to create category", err)
		return nil, err
	}

	err = c.subcategoryRepo.Save(ctx, subcategory)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to save category", err)
		return nil, err
	}

	return &models.Subcategory{
		ID:         subcategory.ID,
		CategoryID: subcategory.CategoryID,
		UserID:     pointer.StringOrNil(user.ID.String()),
		Name:       subcategory.GetName(user.LanguageCode),
		Emoji:      subcategory.Emoji,
		CreatedAt:  subcategory.CreatedAt,
		UpdatedAt:  subcategory.UpdatedAt,
	}, nil
}
