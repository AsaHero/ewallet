package models

import "time"

type Category struct {
	ID        int       `json:"id"`
	UserID    *string   `json:"user_id,omitempty"`
	Name      string    `json:"name"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Subcategory struct {
	ID         int       `json:"id"`
	CategoryID int       `json:"category_id"`
	UserID     *string   `json:"user_id,omitempty"`
	Name       string    `json:"name"`
	Emoji      string    `json:"emoji"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CreateCategoryRequest struct {
	Name  string `json:"name"`
	Emoji string `json:"emoji"`
}

type CreateSubcategoryRequest struct {
	CategoryID int    `json:"category_id"`
	Name       string `json:"name"`
	Emoji      string `json:"emoji"`
}
