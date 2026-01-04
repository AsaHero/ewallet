package models

import "time"

// Debt represents a debt/loan record
type Debt struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	TransactionID string     `json:"transaction_id"`
	Type          string     `json:"type"`   // borrow, lend
	Status        string     `json:"status"` // open, paid, cancelled
	Amount        float64    `json:"amount"`
	CurrencyCode  string     `json:"currency_code"`
	RemindAt      *time.Time `json:"remind_at,omitempty"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}

type UpdateDebtRequest struct {
	Amount       *float64   `json:"amount"`
	CurrencyCode *string    `json:"currency_code"`
	RemindAt     *time.Time `json:"remind_at"`
}

type PayDebtRequest struct {
	PaidAt *time.Time `json:"paid_at"`
}

type DebtsResponse struct {
	Items      []Debt             `json:"items"`
	Pagination PaginationResponse `json:"pagination"`
}
