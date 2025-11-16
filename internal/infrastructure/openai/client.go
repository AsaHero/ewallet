package openai

import (
	"context"
	"net/http"

	"github.com/AsaHero/e-wallet/internal/usecase/ports"
	"github.com/AsaHero/e-wallet/pkg/config"
	"github.com/AsaHero/e-wallet/pkg/utils"
	"github.com/sashabaranov/go-openai"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type apiClient struct {
	client *openai.Client
}

func New(cfg *config.Config) (ports.LLMProvider, error) {
	config := openai.DefaultConfig(cfg.OpenAI.APIKey)
	config.HTTPClient = &http.Client{
		Transport: otelhttp.NewTransport(utils.DefaultInsecureTransport()),
	}

	client := openai.NewClientWithConfig(config)

	return &apiClient{
		client: client,
	}, nil
}

func (c *apiClient) ChatCompletion(ctx context.Context, model string, system string, message string) (string, error) {
	chatCompletionMessages := []openai.ChatCompletionMessage{}

	if system != "" {
		chatCompletionMessages = append(chatCompletionMessages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: system,
		})
	}

	chatCompletionMessages = append(chatCompletionMessages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	})

	req := openai.ChatCompletionRequest{
		Model:    model,
		Messages: chatCompletionMessages,
	}

	completion, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	response := completion.Choices[0].Message.Content

	return response, nil
}
