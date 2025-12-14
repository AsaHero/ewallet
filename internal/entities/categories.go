package entities

import "context"

type CategorySlug string

const (
	Other CategorySlug = "other"
)

func (c CategorySlug) String() string {
	return string(c)
}

type Category struct {
	ID       int
	Slug     string
	Position int
	Name     string
}

// Repository
type CategoryRepository interface {
	FindAll(ctx context.Context) ([]*Category, error)
	FindByID(ctx context.Context, id int) (*Category, error)
	FindBySlug(ctx context.Context, slug string) (*Category, error)
}
