package models

import "time"

type CreateAnonBroadcastRequest struct {
	Language    string         `json:"language"`
	VideoFileID string         `json:"video_file_id"`
	PhotoFileID string         `json:"photo_file_id"`
	Message     string         `json:"message"`
	ReplyMarkup map[string]any `json:"reply_markup"`
}

type AnonBroadcastFilters struct {
	UserIDs       []string `json:"user_ids"`
	LanguageCodes []string `json:"language_codes"`
}

type AnonBroadcastResponse struct {
	ID          string         `json:"id"`
	VideoFileID string         `json:"video_file_id"`
	PhotoFileID string         `json:"photo_file_id"`
	Message     string         `json:"message"`
	ReplyMarkup map[string]any `json:"reply_markup"`
	CreatedAt   time.Time      `json:"created_at"`
}
