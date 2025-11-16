package entities

import "context"

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
