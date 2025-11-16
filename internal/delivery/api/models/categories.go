package models

// Category represents a transaction category
type Category struct {
	ID       int    `json:"id"`
	Slug     string `json:"slug"`
	Position int    `json:"position"`
	Name     string `json:"name"`
}
