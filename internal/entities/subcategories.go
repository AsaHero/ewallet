package entities

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Subcategory struct {
	ID         int
	CategoryID int
	UserID     uuid.UUID
	Position   int
	NameEN     string
	NameRU     string
	NameUZ     string
	Emoji      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewUserSubcategory(
	categoryID int,
	userID uuid.UUID,
	name string,
	emoji string,
) (*Subcategory, error) {
	if categoryID == 0 {
		return nil, errors.New("invalid category id")
	}

	if userID == uuid.Nil {
		return nil, errors.New("invalid user id")
	}

	if name == "" {
		return nil, errors.New("invalid name")
	}

	return &Subcategory{
		CategoryID: categoryID,
		UserID:     userID,
		NameEN:     name,
		NameRU:     name,
		NameUZ:     name,
		Emoji:      emoji,
	}, nil
}

func (s *Subcategory) GetName(lang Language) string {
	switch lang {
	case EN:
		return s.NameEN
	case RU:
		return s.NameRU
	case UZ:
		return s.NameUZ
	default:
		return s.NameEN
	}
}

// Repository
type SubcategoryRepository interface {
	Save(ctx context.Context, subcategory *Subcategory) error
	FindAll(ctx context.Context, userID uuid.UUID) ([]*Subcategory, error)
	FindByID(ctx context.Context, id int) (*Subcategory, error)
	FindByCategoryID(ctx context.Context, categoryID int, userID uuid.UUID) ([]*Subcategory, error)
	Delete(ctx context.Context, userID uuid.UUID, id int) error
}
