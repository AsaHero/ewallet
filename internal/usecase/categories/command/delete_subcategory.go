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

type DeleteSubcategoryUsecase struct {
	contextTimeout  time.Duration
	logger          *logger.Logger
	usersRepo       entities.UserRepository
	subcategoryRepo entities.SubcategoryRepository
}

func NewDeleteSubcategoryUsecase(timeout time.Duration, logger *logger.Logger, usersRepo entities.UserRepository, subcategoryRepo entities.SubcategoryRepository) *DeleteSubcategoryUsecase {
	return &DeleteSubcategoryUsecase{
		contextTimeout:  timeout,
		logger:          logger,
		subcategoryRepo: subcategoryRepo,
		usersRepo:       usersRepo,
	}
}

type DeleteSubcategoryCommand struct {
	UserID        string `json:"user_id"`
	SubcategoryID int    `json:"subcategory_id"`
}

func (c *DeleteSubcategoryUsecase) DeleteSubcategory(ctx context.Context, cmd *DeleteSubcategoryCommand) (err error) {
	ctx, cancel := context.WithTimeout(ctx, c.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("categories"), "DeleteSubcategory")
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

	err = c.subcategoryRepo.Delete(ctx, userID, cmd.SubcategoryID)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to delete subcategory", err)
		return err
	}

	return nil
}
