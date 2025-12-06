package repository

import (
	"context"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/pkg/database/postgres"
	"github.com/google/uuid"
	"github.com/shogo82148/pointer"
	"github.com/uptrace/bun"
)

type Users struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID           string     `bun:"id,type:uuid,pk"`
	TGUserID     int64      `bun:"tg_user_id"`
	FirstName    *string    `bun:"first_name,nullzero"`
	LastName     *string    `bun:"last_name,nullzero"`
	Username     *string    `bun:"username,nullzero"`
	LanguageCode string     `bun:"language_code,nullzero"`
	CurrencyCode string     `bun:"currency_code,nullzero"`
	CreatedAt    time.Time  `bun:"created_at,default:current_timestamp"`
	UpdatedAt    *time.Time `bun:"updated_at,nullzero"`
}

type usersRepo struct {
	db bun.IDB
}

func NewUsersRepo(db bun.IDB) entities.UserRepository {
	return &usersRepo{
		db: db,
	}
}

func (r *usersRepo) Save(ctx context.Context, user *entities.User) error {
	db := postgres.FromContext(ctx, r.db)
	var model = r.ToModel(user)

	_, err := db.NewInsert().Model(model).
		On("CONFLICT (id) DO UPDATE").
		Set("tg_user_id = EXCLUDED.tg_user_id").
		Set("first_name = EXCLUDED.first_name").
		Set("last_name = EXCLUDED.last_name").
		Set("username = EXCLUDED.username").
		Set("language_code = EXCLUDED.language_code").
		Set("currency_code = EXCLUDED.currency_code").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	if err != nil {
		return postgres.Error(err, model)
	}

	return err
}

func (r *usersRepo) FindByID(ctx context.Context, userID uuid.UUID) (*entities.User, error) {
	db := postgres.FromContext(ctx, r.db)

	var model Users
	err := db.NewSelect().Model(&model).
		Where("id = ?", userID.String()).
		Scan(ctx)
	if err != nil {
		return nil, postgres.Error(err, model)
	}

	return r.ToEntity(&model), nil
}

func (r *usersRepo) FindByTGUserID(ctx context.Context, tgUserID int64) (*entities.User, error) {
	db := postgres.FromContext(ctx, r.db)

	var model Users
	err := db.NewSelect().Model(&model).
		Where("tg_user_id = ?", tgUserID).
		Scan(ctx)
	if err != nil {
		return nil, postgres.Error(err, model)
	}

	return r.ToEntity(&model), nil
}

func (r *usersRepo) FindAll(ctx context.Context) ([]*entities.User, error) {
	db := postgres.FromContext(ctx, r.db)

	var models []Users
	err := db.NewSelect().Model(&models).
		Scan(ctx)
	if err != nil {
		return nil, postgres.Error(err, models)
	}

	var users []*entities.User = make([]*entities.User, len(models))
	for i, model := range models {
		users[i] = r.ToEntity(&model)
	}

	return users, nil
}

func (r *usersRepo) ToModel(e *entities.User) *Users {
	if e == nil {
		return nil
	}

	users := &Users{
		ID:           e.ID.String(),
		TGUserID:     e.TGUserID,
		FirstName:    pointer.StringOrNil(e.FirstName),
		LastName:     pointer.StringOrNil(e.LastName),
		Username:     pointer.StringOrNil(e.Username),
		LanguageCode: e.LanguageCode.String(),
		CurrencyCode: e.CurrencyCode.String(),
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    pointer.TimeOrNil(e.UpdatedAt),
	}

	return users
}

func (r *usersRepo) ToEntity(m *Users) *entities.User {
	if m == nil {
		return nil
	}

	id, _ := uuid.Parse(m.ID)

	e := &entities.User{
		ID:           id,
		TGUserID:     m.TGUserID,
		FirstName:    pointer.StringValue(m.FirstName),
		LastName:     pointer.StringValue(m.LastName),
		Username:     pointer.StringValue(m.Username),
		LanguageCode: entities.Language(m.LanguageCode),
		CurrencyCode: entities.Currency(m.CurrencyCode),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    pointer.TimeValue(m.UpdatedAt),
	}

	return e
}
