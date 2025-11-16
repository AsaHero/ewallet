package models



// Request/Response DTOs
type AuthRequest struct {
	TgUserID     int64  `json:"tg_user_id" binding:"required"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
	CurrencyCode string `json:"currency_code"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
