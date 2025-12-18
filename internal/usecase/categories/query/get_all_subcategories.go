package query

import (
	"context"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/inerr"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/google/uuid"
)

type GetAllSubcategoriesUsecase struct {
	contextTimeout    time.Duration
	logger            *logger.Logger
	usersRepo         entities.UserRepository
	subcategoriesRepo entities.SubcategoryRepository
}

func NewGetAllSubcategoriesUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	usersRepo entities.UserRepository,
	subcategoriesRepo entities.SubcategoryRepository,
) *GetAllSubcategoriesUsecase {
	return &GetAllSubcategoriesUsecase{
		contextTimeout:    timeout,
		logger:            logger,
		usersRepo:         usersRepo,
		subcategoriesRepo: subcategoriesRepo,
	}
}

type Subcategory struct {
	ID         int    `json:"id"`
	UserID     string `json:"user_id"`
	CategoryID int    `json:"category_id"`
	Position   int    `json:"position"`
	Name       string `json:"name"`
	Emoji      string `json:"emoji"`
}

func (u *GetAllSubcategoriesUsecase) GetAllSubcategories(ctx context.Context, userID string) (_ []Subcategory, err error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	uuid, err := uuid.Parse(userID)
	if err != nil {
		return nil, inerr.NewErrValidation("user_id", "invalid user id")
	}

	user, err := u.usersRepo.FindByID(ctx, uuid)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get user", err)
		return nil, err
	}

	list, err := u.subcategoriesRepo.FindAll(ctx, user.ID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to get subcategories", err)
		return nil, err
	}

	var subcategories []Subcategory
	for _, subcategory := range list {
		subcategories = append(subcategories, Subcategory{
			ID:         subcategory.ID,
			UserID:     user.ID.String(),
			CategoryID: subcategory.CategoryID,
			Position:   subcategory.Position,
			Name:       subcategory.GetName(user.LanguageCode),
			Emoji:      subcategory.Emoji,
		})
	}

	return subcategories, nil
}
