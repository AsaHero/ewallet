package entities

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	TGUserID     int64
	FirstName    string
	LastName     string
	Username     string
	LanguageCode Language
	CurrencyCode Currency
	Timezone     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(tgUserID int64, firstName, lastName, username string) (*User, error) {
	if tgUserID == 0 {
		return nil, errors.New("invalid telegram user id")
	}

	return &User{
		ID:        uuid.New(),
		TGUserID:  tgUserID,
		FirstName: firstName,
		LastName:  lastName,
		Username:  username,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (u *User) UpdateLanguageCode(code Language) {
	u.LanguageCode = code
	u.UpdatedAt = time.Now()
}

func (u *User) UpdateCurrencyCode(code Currency) {
	u.CurrencyCode = code
	u.UpdatedAt = time.Now()
}

func (u *User) UpdateTimezone(timezone string) {
	u.Timezone = timezone
	u.UpdatedAt = time.Now()
}

// Refpository

type UserRepository interface {
	Save(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByTGUserID(ctx context.Context, tgUserID int64) (*User, error)
	FindAll(ctx context.Context) ([]*User, error)
}
