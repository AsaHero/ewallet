package ports

import "context"

type FXRatesProvider interface {
	GetRate(ctx context.Context, baseCurrency string, targetCurrency string) (float64, error)
}
