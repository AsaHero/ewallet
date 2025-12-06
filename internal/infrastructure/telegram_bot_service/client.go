package telegram_bot_service

import (
	"context"
	"net/http"

	"github.com/AsaHero/e-wallet/internal/inerr"
	"github.com/AsaHero/e-wallet/internal/usecase/ports"
	"github.com/AsaHero/e-wallet/pkg/config"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"resty.dev/v3"
)

type apiClient struct {
	cfg        *config.Config
	httpClient *resty.Client
}

func New(cfg *config.Config) (*apiClient, error) {
	httpClient := resty.New().
		SetBaseURL(cfg.TelegramBotService.BaseURL).
		SetTimeout(cfg.TelegramBotService.Timeout).
		SetResponseBodyUnlimitedReads(true).
		SetRetryCount(3).
		SetTransport(otelhttp.NewTransport(http.DefaultTransport))

	return &apiClient{
		cfg:        cfg,
		httpClient: httpClient,
	}, nil
}

func (c *apiClient) SendMessage(ctx context.Context, req *ports.SendMessageRequest) error {
	body := map[string]any{
		"userId":    req.UserID,
		"text":      req.Text,
		"parseMode": req.ParseMode,
	}

	var response Response
	httpResponse, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(body).
		SetResult(&response).
		Post("/api/send-message")

	if err != nil {
		return err
	}

	if !httpResponse.IsSuccess() {
		return inerr.NewErrHttp(
			httpResponse.StatusCode(),
			httpResponse.Request.Method,
			httpResponse.Request.URL,
			response.Error,
			httpResponse.Bytes(),
		)
	}
	if !response.Success {
		return inerr.NewErrHttp(
			httpResponse.StatusCode(),
			httpResponse.Request.Method,
			httpResponse.Request.URL,
			response.Error,
			httpResponse.Bytes(),
		)
	}

	return nil
}
