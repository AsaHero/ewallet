package entities

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Anon struct {
	ID           uuid.UUID
	LanguageCode string
	VideoFileID  string
	PhotoFileID  string
	Message      string
	ReplyMarkup  []byte
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewAnon(message string, languageCode Language, replyMarkup []byte) (*Anon, error) {
	if message == "" {
		return nil, errors.New("message is required")
	}

	if languageCode == "" {
		return nil, errors.New("language code is required")
	}

	return &Anon{
		ID:           uuid.New(),
		LanguageCode: languageCode.String(),
		Message:      message,
		ReplyMarkup:  replyMarkup,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

func (a *Anon) SetVideoFileID(videoFileID string) {
	if videoFileID != "" {
		a.VideoFileID = videoFileID
		a.UpdatedAt = time.Now()
	}
}

func (a *Anon) SetPhotoFileID(photoFileID string) {
	if photoFileID != "" {
		a.PhotoFileID = photoFileID
		a.UpdatedAt = time.Now()
	}
}

func (a *Anon) SetMessage(message string) {
	if message != "" {
		a.Message = message
		a.UpdatedAt = time.Now()
	}
}

type AnonRepository interface {
	Save(ctx context.Context, anon *Anon) error
	FindByID(ctx context.Context, id uuid.UUID) (*Anon, error)
	FindAll(ctx context.Context) ([]*Anon, error)
}
