package command

import (
	"context"
	"time"

	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"

	"github.com/AsaHero/e-wallet/internal/entities"
)

type DeleteCategoryUsecase struct {
	contextTimeout time.Duration
	logger         *logger.Logger
	usersRepo      entities.UserRepository
	categoryRepo   entities.CategoryRepository
}

func NewDeleteCategoryUsecase(timeout time.Duration, logger *logger.Logger, usersRepo entities.UserRepository, categoryRepo entities.CategoryRepository) *DeleteCategoryUsecase {
	return &DeleteCategoryUsecase{
		contextTimeout: timeout,
		logger:         logger,
		categoryRepo:   categoryRepo,
		usersRepo:      usersRepo,
	}
}

type DeleteCategoryCommand struct {
	UserID     string `json:"user_id"`
	CategoryID int    `json:"category_id"`
}

func (c *DeleteCategoryUsecase) DeleteCategory(ctx context.Context, cmd *DeleteCategoryCommand) (err error) {
	ctx, cancel := context.WithTimeout(ctx, c.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("categories"), "DeleteCategory")
	defer func() { end(err) }()

	userID, err := uuid.Parse(cmd.UserID)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to parse user id", err)
		return err
	}

	_, err = c.usersRepo.FindByID(ctx, userID)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to find user", err)
		return err
	}

	err = c.categoryRepo.Delete(ctx, userID, cmd.CategoryID)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to delete category", err)
		return err
	}

	return nil
}
