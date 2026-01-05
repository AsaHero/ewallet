package ports

import "context"

type TelegramBotService interface {
	SendMessage(ctx context.Context, req *SendMessageRequest) error
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]map[string]any `json:"inline_keyboard"`
}

type ReplyKeyboardMarkup struct {
	Keyboard [][]map[string]any `json:"keyboard"`
}

type ReplyKeyboardRemove struct {
	RemoveKeyboard bool `json:"remove_keyboard"`
}

type SendMessageRequest struct {
	UserID      int64
	Text        string
	ParseMode   string
	ReplyMarkup any
}
