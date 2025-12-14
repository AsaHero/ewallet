package models

import "time"

// Transaction represents a financial transaction
type Transaction struct {
	ID                   string     `json:"id"`
	UserID               string     `json:"user_id"`
	AccountID            string     `json:"account_id"`
	CategoryID           *int       `json:"category_id,omitempty"`
	Type                 string     `json:"type"`
	Status               string     `json:"status,omitempty"`
	Amount               float64    `json:"amount"`
	CurrencyCode         string     `json:"currency_code"`
	OriginalAmount       *float64   `json:"original_amount,omitempty"`
	OriginalCurrencyCode *string    `json:"original_currency_code,omitempty"`
	FxRate               *float64   `json:"fx_rate,omitempty"`
	Note                 string     `json:"note,omitempty"`
	PerformedAt          *time.Time `json:"performed_at,omitempty"`
	RejectedAt           *time.Time `json:"rejected_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
}

// ParseTransactionRequest represents payload that is parsed by AI
type ParseTextRequest struct {
	Content string `json:"content" binding:"required"`
}

type ParseAudioRequest struct {
	FileURL string `json:"file_url" binding:"required"`
}

type ParseImageRequest struct {
	ImageURL string `json:"image_url" binding:"required"`
}

type CreateTransactionRequest struct {
	AccountID            string     `json:"account_id" binding:"required"`
	CategoryID           *int       `json:"category_id"`
	Type                 string     `json:"type" binding:"required"`
	Amount               float64    `json:"amount" binding:"required"`
	CurrencyCode         string     `json:"currency_code"`
	OriginalAmount       *float64   `json:"original_amount,omitempty"`
	OriginalCurrencyCode *string    `json:"original_currency_code,omitempty"`
	FxRate               *float64   `json:"fx_rate,omitempty"`
	Note                 string     `json:"note"`
	PerformedAt          *time.Time `json:"performed_at"`
}

type UpdateTransactionRequest struct {
	Amount     *float64 `json:"amount"`
	CategoryID *int     `json:"category_id"`
	Note       *string  `json:"note"`
}

type TransactionsResponse struct {
	Items      []Transaction      `json:"items"`
	Pagination PaginationResponse `json:"pagination"`
}
