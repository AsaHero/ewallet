package entities

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type CategoryID int

const (
	OtherCategory CategoryID = 25
)

func (c CategoryID) Int() int {
	return int(c)
}

type Category struct {
	ID        CategoryID
	UserID    uuid.UUID
	Position  int
	NameEN    string
	NameRU    string
	NameUZ    string
	Emoji     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUserCategory(
	userID uuid.UUID,
	name string,
	emoji string) (*Category, error) {
	if userID == uuid.Nil {
		return nil, errors.New("invalid user id")
	}

	if name == "" {
		return nil, errors.New("invalid name")
	}

	return &Category{
		UserID: userID,
		NameEN: name,
		NameRU: name,
		NameUZ: name,
		Emoji:  emoji,
	}, nil
}

func (c *Category) GetName(lang Language) string {
	switch lang {
	case EN:
		return c.NameEN
	case RU:
		return c.NameRU
	case UZ:
		return c.NameUZ
	default:
		return c.NameEN
	}
}

// Repository
type CategoryRepository interface {
	Save(ctx context.Context, category *Category) error
	FindAll(ctx context.Context, userID uuid.UUID) ([]*Category, error)
	FindByID(ctx context.Context, id int) (*Category, error)
	Delete(ctx context.Context, userID uuid.UUID, id int) error
}
