package ports

import "context"

type OCRProvider interface {
	ImageToText(ctx context.Context, imageURL string) (string, error)
}
