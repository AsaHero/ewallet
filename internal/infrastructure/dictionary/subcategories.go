package dictionary

import (
	"context"
	"fmt"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/pkg/database/postgres"
	"github.com/google/uuid"
	"github.com/shogo82148/pointer"
	"github.com/uptrace/bun"
)

type Subcategories struct {
	bun.BaseModel `bun:"table:subcategories,alias:sb"`

	ID         int        `bun:"id,pk"`
	CategoryID int        `bun:"category_id"`
	UserID     *string    `bun:"user_id,nullzero"`
	Position   int        `bun:"position"`
	NameEN     string     `bun:"name_en"`
	NameRU     string     `bun:"name_ru"`
	NameUZ     string     `bun:"name_uz"`
	Emoji      string     `bun:"emoji"`
	CreatedAt  time.Time  `bun:"created_at"`
	UpdatedAt  *time.Time `bun:"updated_at,nullzero"`
}

func (s Subcategories) Key() int {
	return s.ID
}

func (s Subcategories) Code() string {
	return fmt.Sprintf("%d", s.ID)
}

type subcategoriesDict struct {
	*postgres.BaseDictionary[int, *Subcategories]
	db bun.IDB
}

func NewSubcategoriesDict(db bun.IDB) entities.SubcategoryRepository {
	return &subcategoriesDict{
		BaseDictionary: postgres.NewDictionary(db, postgres.WithOrderBy[int, *Subcategories]("position", "asc")),
		db:             db,
	}
}

func (d *subcategoriesDict) FindAll(ctx context.Context, userID uuid.UUID) ([]*entities.Subcategory, error) {
	items, err := d.BaseDictionary.Values(ctx)
	if err != nil {
		return nil, err
	}

	var subcategories []*entities.Subcategory
	for _, item := range items {
		if item.UserID != nil && *item.UserID != userID.String() {
			continue
		}
		subcategories = append(subcategories, d.ToEntity(item))
	}

	return subcategories, nil
}

func (d *subcategoriesDict) FindByID(ctx context.Context, id int) (*entities.Subcategory, error) {
	item, err := d.BaseDictionary.GetByKey(ctx, id)
	if err != nil {
		return nil, err
	}

	return d.ToEntity(item), nil
}

func (d *subcategoriesDict) FindByCategoryID(ctx context.Context, categoryID int, userID uuid.UUID) ([]*entities.Subcategory, error) {
	items, err := d.BaseDictionary.Values(ctx)
	if err != nil {
		return nil, err
	}

	var subcategories []*entities.Subcategory
	for _, item := range items {
		if item.CategoryID != categoryID {
			continue
		}
		if item.UserID != nil && *item.UserID != userID.String() {
			continue
		}
		subcategories = append(subcategories, d.ToEntity(item))
	}

	return subcategories, nil
}

func (d *subcategoriesDict) Delete(ctx context.Context, userID uuid.UUID, id int) error {
	db := postgres.FromContext(ctx, d.db)

	_, err := db.NewDelete().Model(&Subcategories{}).
		Where("user_id = ?", userID.String()).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return postgres.Error(err, Subcategories{})
	}

	d.BaseDictionary.Load(ctx)

	return nil
}

func (d *subcategoriesDict) ToEntity(s *Subcategories) *entities.Subcategory {

	var userID uuid.UUID
	if s.UserID != nil {
		userID, _ = uuid.Parse(*s.UserID)
	}

	return &entities.Subcategory{
		ID:         s.ID,
		CategoryID: s.CategoryID,
		UserID:     userID,
		Position:   s.Position,
		NameEN:     s.NameEN,
		NameRU:     s.NameRU,
		NameUZ:     s.NameUZ,
		Emoji:      s.Emoji,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  pointer.TimeValue(s.UpdatedAt),
	}
}

func (d *subcategoriesDict) ToModel(s *entities.Subcategory) *Subcategories {
	return &Subcategories{
		ID:         s.ID,
		CategoryID: s.CategoryID,
		UserID:     pointer.String(s.UserID.String()),
		Position:   s.Position,
		NameEN:     s.NameEN,
		NameRU:     s.NameRU,
		NameUZ:     s.NameUZ,
		Emoji:      s.Emoji,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  pointer.Time(s.UpdatedAt),
	}
}
