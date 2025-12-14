package models

import "time"

type User struct {
	ID           string     `json:"id"`
	TgUserID     int64      `json:"tg_user_id"`
	FirstName    string     `json:"first_name,omitempty"`
	LastName     string     `json:"last_name,omitempty"`
	Username     string     `json:"username,omitempty"`
	LanguageCode string     `json:"language_code,omitempty"`
	CurrencyCode string     `json:"currency_code,omitempty"`
	Timezone     string     `json:"timezone,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

type UpdateUserRequest struct {
	CurrencyCode *string `json:"currency_code"`
	LanguageCode *string `json:"language_code"`
	Timezone     *string `json:"timezone"`
}
