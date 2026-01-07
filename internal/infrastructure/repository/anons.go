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

type Anons struct {
	bun.BaseModel `bun:"table:anons,alias:a"`

	ID           string     `bun:"id,type:uuid,pk"`
	LanguageCode string     `bun:"language_code"`
	VideoFileID  *string    `bun:"video_file_id,nullzero"`
	PhotoFileID  *string    `bun:"photo_file_id,nullzero"`
	Message      string     `bun:"message,nullzero"`
	ReplyMarkup  []byte     `bun:"reply_markup,type:jsonb"`
	CreatedAt    time.Time  `bun:"created_at,default:current_timestamp"`
	UpdatedAt    *time.Time `bun:"updated_at,default:current_timestamp"`
}

type anonsRepo struct {
	db bun.IDB
}

func NewAnonsRepo(db bun.IDB) entities.AnonRepository {
	return &anonsRepo{
		db: db,
	}
}

func (r *anonsRepo) Save(ctx context.Context, anon *entities.Anon) error {
	db := postgres.FromContext(ctx, r.db)
	var model = r.ToModel(anon)

	_, err := db.NewInsert().Model(model).
		On("CONFLICT (id) DO UPDATE").
		Set("language_code = EXCLUDED.language_code").
		Set("video_file_id = EXCLUDED.video_file_id").
		Set("photo_file_id = EXCLUDED.photo_file_id").
		Set("message = EXCLUDED.message").
		Set("reply_markup = EXCLUDED.reply_markup").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	if err != nil {
		return postgres.Error(err, model)
	}

	return nil
}

func (r *anonsRepo) FindByID(ctx context.Context, id uuid.UUID) (*entities.Anon, error) {
	db := postgres.FromContext(ctx, r.db)

	var model Anons
	err := db.NewSelect().Model(&model).
		Where("id = ?", id.String()).
		Scan(ctx)
	if err != nil {
		return nil, postgres.Error(err, model)
	}

	return r.ToEntity(&model), nil
}

func (r *anonsRepo) FindAll(ctx context.Context) ([]*entities.Anon, error) {
	db := postgres.FromContext(ctx, r.db)

	var models []Anons
	err := db.NewSelect().Model(&models).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, postgres.Error(err, models)
	}

	anons := make([]*entities.Anon, len(models))
	for i, model := range models {
		anons[i] = r.ToEntity(&model)
	}

	return anons, nil
}

func (r *anonsRepo) ToModel(e *entities.Anon) *Anons {
	if e == nil {
		return nil
	}

	return &Anons{
		ID:           e.ID.String(),
		LanguageCode: e.LanguageCode,
		VideoFileID:  pointer.StringOrNil(e.VideoFileID),
		PhotoFileID:  pointer.StringOrNil(e.PhotoFileID),
		Message:      e.Message,
		ReplyMarkup:  e.ReplyMarkup,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    pointer.TimeOrNil(e.UpdatedAt),
	}
}

func (r *anonsRepo) ToEntity(m *Anons) *entities.Anon {
	if m == nil {
		return nil
	}

	id, _ := uuid.Parse(m.ID)
	return &entities.Anon{
		ID:           id,
		LanguageCode: m.LanguageCode,
		VideoFileID:  pointer.StringValue(m.VideoFileID),
		PhotoFileID:  pointer.StringValue(m.PhotoFileID),
		Message:      m.Message,
		ReplyMarkup:  m.ReplyMarkup,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    pointer.TimeValue(m.UpdatedAt),
	}
}
