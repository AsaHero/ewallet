package telegram_bot_service_test

import (
	"testing"

	"github.com/AsaHero/e-wallet/internal/infrastructure/telegram_bot_service"
	"github.com/AsaHero/e-wallet/internal/usecase/ports"
	"github.com/AsaHero/e-wallet/pkg/config"
	"github.com/joho/godotenv"
)

func TestSendMessage(t *testing.T) {
	godotenv.Load("../../../.env")
	cfg, err := config.New()
	if err != nil {
		t.Fatal(err)
	}
	apiClient, err := telegram_bot_service.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	reqeustData := &ports.SendMessageRequest{
		UserID:    131372022,
		Text:      "TestMessage with Test Keybaord with Test <b>parseMode</b>",
		ParseMode: "HTML",
		ReplyMarkup: ports.InlineKeyboardMarkup{
			InlineKeyboard: [][]map[string]any{
				{
					{
						"text":          "Test",
						"callback_data": "test",
					},
				},
			},
		},
	}

	if err := apiClient.SendMessage(t.Context(), reqeustData); err != nil {
		t.Fatal(err)
	}
}
