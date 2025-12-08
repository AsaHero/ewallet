package ocr_service

import (
	"context"
	"net/http"

	"github.com/AsaHero/e-wallet/internal/inerr"
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
		SetBaseURL(cfg.OCRService.BaseURL).
		SetTimeout(cfg.OCRService.Timeout).
		SetRetryCount(3).
		SetResponseBodyUnlimitedReads(true).
		SetTransport(otelhttp.NewTransport(http.DefaultTransport))

	return &apiClient{
		cfg:        cfg,
		httpClient: httpClient,
	}, nil
}

func (c *apiClient) ImageToText(ctx context.Context, imageURL string) (string, error) {
	data := map[string]any{
		"url": imageURL,
	}

	var response ImageToTextResponse
	var errResponse ErrorResponse
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetResult(&response).
		SetError(&errResponse).
		SetBody(data).
		Post("/ocr/url")
	if err != nil {
		return "", err
	}

	if resp.StatusCode() != http.StatusOK {
		return "", inerr.NewErrHttp(
			resp.StatusCode(),
			resp.Request.Method,
			resp.Request.URL,
			errResponse.Details,
			resp.Bytes(),
		)
	}

	return response.FullText, nil
}
