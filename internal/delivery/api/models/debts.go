package models

import "time"

// Debt represents a debt/loan record
type Debt struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	TransactionID string     `json:"transaction_id"`
	Type          string     `json:"type"`   // borrow, lend
	Status        string     `json:"status"` // open, paid, cancelled
	Name          string     `json:"name"`
	Amount        float64    `json:"amount"`
	CurrencyCode  string     `json:"currency_code"`
	Note          string     `json:"note"`
	DueAt         *time.Time `json:"due_at,omitempty"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}

type CreateDebtRequest struct {
	TransactionID string     `json:"transaction_id"`
	Type          string     `json:"type"`
	Name          *string    `json:"name"`
	DueAt         *time.Time `json:"due_at"`
	Note          *string    `json:"note"`
}

type UpdateDebtRequest struct {
	Amount *float64   `json:"amount"`
	Name   *string    `json:"name"`
	DueAt  *time.Time `json:"due_at"`
	Note   *string    `json:"note"`
}

type PayDebtRequest struct {
	PaidAt *time.Time `json:"paid_at"`
}

type DebtsResponse struct {
	Items      []Debt             `json:"items"`
	Pagination PaginationResponse `json:"pagination"`
}
