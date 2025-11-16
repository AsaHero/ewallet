package models

import "time"

// Account represents a user's financial account
type Account struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Name      string     `json:"name"`
	Balance   float64    `json:"balance"`
	IsDefault bool       `json:"is_default"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type CreateAccountRequest struct {
	Name      string  `json:"name" binding:"required"`
	Balance   float64 `json:"balance"`
	IsDefault bool    `json:"is_default"`
}

type UpdateAccountRequest struct {
	Name      *string `json:"name"`
	IsDefault *bool   `json:"is_default"`
}
