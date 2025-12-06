package ports

import "context"

type TelegramBotService interface {
	SendMessage(ctx context.Context, req *SendMessageRequest) error
}

type SendMessageRequest struct {
	UserID    int64
	Text      string
	ParseMode string
}
