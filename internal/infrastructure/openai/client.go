package openai

import (
	"context"
	"fmt"
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

func (c *apiClient) AudioToText(ctx context.Context, filePath string, language string) (string, error) {
	req := openai.AudioRequest{
		Model:    "gpt-4o-transcribe",
		Prompt:   fmt.Sprintf("The audio might be in %s language.", language),
		FilePath: filePath,
	}

	transcript, err := c.client.CreateTranscription(ctx, req)
	if err != nil {
		return "", err
	}

	return transcript.Text, nil
}

func (c *apiClient) ImageToText(ctx context.Context, imageURL string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model: openai.GPT4VisionPreview,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: "Extract all text and transaction-related information from this image. Include amounts, dates, categories, and any notes you can identify.",
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL: imageURL,
						},
					},
				},
			},
		},
		MaxTokens: 500,
	}

	completion, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	return completion.Choices[0].Message.Content, nil
}
