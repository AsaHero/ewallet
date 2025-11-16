package dictionary

import (
	"context"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/pkg/database/postgres"
	"github.com/uptrace/bun"
)

type Categories struct {
	bun.BaseModel `bun:"table:categories,alias:c"`

	ID       int    `bun:"id,pk"`
	Slug     string `bun:"slug"`
	Position int    `bun:"position"`
	Name     string `bun:"name"`
}

func (c Categories) Key() int {
	return c.ID
}

func (c Categories) Code() string {
	return c.Slug
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

func (d *categoriesDict) FindAll(ctx context.Context) ([]*entities.Category, error) {
	items, err := d.BaseDictionary.Values(ctx)
	if err != nil {
		return nil, err
	}

	var categories []*entities.Category
	for _, item := range items {
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
func (d *categoriesDict) FindBySlug(ctx context.Context, slug string) (*entities.Category, error) {
	item, err := d.BaseDictionary.GetByCode(ctx, slug)
	if err != nil {
		return nil, err
	}

	return d.ToEntity(item), nil
}

func (d *categoriesDict) ToEntity(c *Categories) *entities.Category {
	return &entities.Category{
		ID:       c.ID,
		Slug:     c.Slug,
		Position: c.Position,
		Name:     c.Name,
	}
}
