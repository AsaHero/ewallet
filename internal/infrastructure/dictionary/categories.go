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

type Categories struct {
	bun.BaseModel `bun:"table:categories,alias:c"`

	ID        int        `bun:"id,pk"`
	UserID    *string    `bun:"user_id"`
	Position  int        `bun:"position"`
	NameEN    string     `bun:"name_en"`
	NameRU    string     `bun:"name_ru"`
	NameUZ    string     `bun:"name_uz"`
	Emoji     string     `bun:"emoji"`
	CreatedAt time.Time  `bun:"created_at"`
	UpdatedAt *time.Time `bun:"updated_at,nullzero"`
}

func (c Categories) Key() int {
	return c.ID
}

func (c Categories) Code() string {
	return fmt.Sprintf("%d", c.ID)
}

type categoriesDict struct {
	*postgres.BaseDictionary[int, *Categories]
	db bun.IDB
}

func NewCategoriesDict(db bun.IDB) entities.CategoryRepository {
	return &categoriesDict{
		BaseDictionary: postgres.NewDictionary(db, postgres.WithOrderBy[int, *Categories]("position", "asc")),
		db:             db,
	}
}

func (d *categoriesDict) FindAll(ctx context.Context, userID uuid.UUID) ([]*entities.Category, error) {
	items, err := d.BaseDictionary.Values(ctx)
	if err != nil {
		return nil, err
	}

	var categories []*entities.Category
	for _, item := range items {
		if item.UserID != nil && *item.UserID != userID.String() {
			continue
		}
		categories = append(categories, d.ToEntity(item))
	}

	return categories, nil
}
func (d *categoriesDict) FindByID(ctx context.Context, id int) (*entities.Category, error) {
	item, err := d.BaseDictionary.GetByKey(ctx, id)
	if err != nil {
		return nil, err
	}

	return d.ToEntity(item), nil
}

func (d *categoriesDict) Delete(ctx context.Context, userID uuid.UUID, id int) error {
	db := postgres.FromContext(ctx, d.db)

	_, err := db.NewDelete().Model(&Categories{}).
		Where("user_id = ?", userID.String()).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return postgres.Error(err, Categories{})
	}

	d.BaseDictionary.Load(ctx)

	return nil
}

func (d *categoriesDict) ToEntity(c *Categories) *entities.Category {
	var userID uuid.UUID
	if c.UserID != nil {
		userID, _ = uuid.Parse(*c.UserID)
	}

	return &entities.Category{
		ID:        entities.CategoryID(c.ID),
		UserID:    userID,
		Position:  c.Position,
		NameEN:    c.NameEN,
		NameRU:    c.NameRU,
		NameUZ:    c.NameUZ,
		Emoji:     c.Emoji,
		CreatedAt: c.CreatedAt,
		UpdatedAt: pointer.TimeValue(c.UpdatedAt),
	}
}

func (d *categoriesDict) ToModel(c *entities.Category) *Categories {
	return &Categories{
		ID:        c.ID.Int(),
		UserID:    pointer.String(c.UserID.String()),
		Position:  c.Position,
		NameEN:    c.NameEN,
		NameRU:    c.NameRU,
		NameUZ:    c.NameUZ,
		Emoji:     c.Emoji,
		CreatedAt: c.CreatedAt,
		UpdatedAt: pointer.Time(c.UpdatedAt),
	}
}
