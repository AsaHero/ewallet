package currency_api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/AsaHero/e-wallet/internal/inerr"
	"github.com/AsaHero/e-wallet/pkg/redis"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"resty.dev/v3"
)

type apiClient struct {
	httpClient *resty.Client
	redis      *redis.RedisClient
}

func New(redis *redis.RedisClient) (*apiClient, error) {
	httpClient := resty.New().
		SetBaseURL("https://latest.currency-api.pages.dev").
		SetTimeout(30 * time.Second).
		SetRetryCount(3).
		SetResponseBodyUnlimitedReads(true).
		SetTransport(otelhttp.NewTransport(http.DefaultTransport))

	return &apiClient{
		httpClient: httpClient,
		redis:      redis,
	}, nil
}

func (c *apiClient) GetRate(ctx context.Context, fromCurrency string, toCurrency string) (float64, error) {
	from := strings.ToLower(strings.TrimSpace(fromCurrency))
	to := strings.ToLower(strings.TrimSpace(toCurrency))

	if from == "" || to == "" {
		return 0, fmt.Errorf("currency codes must be non-empty")
	}
	if from == to {
		return 1, nil
	}

	// ---- Cache (optional) ----
	cacheKey := "fx:" + from
	if c.redis != nil {
		if b, err := c.redis.GetBytes(ctx, cacheKey); err == nil && len(b) > 0 {
			r, ok, err := extractRateFromPayload(b, from, to)
			if err == nil && ok {
				return r, nil
			}
			// if cache is corrupted/missing target, fall through to fetch
		}
	}

	// ---- Fetch ----
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("accept", "application/json").
		Get("/v1/currencies/" + from + ".json")
	if err != nil {
		return 0, inerr.NewErrHttp(
			http.StatusInternalServerError,
			"GET",
			"/v1/currencies/"+from+".json",
			err.Error(),
			nil,
		)
	}

	if resp.IsError() {
		return 0, inerr.NewErrHttp(
			resp.StatusCode(),
			resp.Request.Method,
			resp.Request.URL,
			"",
			resp.Bytes(),
		)
	}

	body := resp.Bytes()

	rate, ok, err := extractRateFromPayload(body, from, to)
	if err != nil {
		return 0, inerr.NewErrHttp(
			resp.StatusCode(),
			resp.Request.Method,
			resp.Request.URL,
			err.Error(),
			body,
		)
	}
	if !ok {
		return 0, fmt.Errorf("rate not found: from=%s to=%s", from, to)
	}

	// ---- Save cache (optional) ----
	if c.redis != nil {
		_ = c.redis.SetBytes(ctx, cacheKey, body, 12*time.Hour) // ignore cache errors
	}

	return rate, nil
}

func extractRateFromPayload(payload []byte, from string, to string) (rate float64, ok bool, err error) {
	// payload looks like:
	// { "date": "...", "<from>": { "<to>": 0.123, ... } }

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		return 0, false, fmt.Errorf("invalid json payload: %w", err)
	}

	baseBlob, exists := raw[from]
	if !exists {
		return 0, false, fmt.Errorf("payload missing base currency field %q", from)
	}

	var rates map[string]float64
	if err := json.Unmarshal(baseBlob, &rates); err != nil {
		return 0, false, fmt.Errorf("invalid rates object for base %q: %w", from, err)
	}

	r, exists := rates[to]
	if !exists {
		return 0, false, nil
	}
	return r, true, nil
}
