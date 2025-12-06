package telegram_bot_service

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error"`
}
