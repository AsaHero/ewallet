package ports

import "context"

type LLMProvider interface {
	ChatCompletion(ctx context.Context, model, system, message string) (string, error)
	AudioToText(ctx context.Context, filePath string) (string, error)
}
