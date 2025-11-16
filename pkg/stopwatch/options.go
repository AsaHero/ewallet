package stopwatch

import (
	"time"

	"github.com/AsaHero/e-wallet/pkg/logger"
)

type Options func(*options)

type options struct {
	StartTime time.Time
	Logger    *logger.Logger
	AutoLog   bool
	OnStop    []func(elapsed time.Duration)
}

func WithStartTime(startTime time.Time) Options {
	return func(o *options) {
		o.StartTime = startTime
	}
}

func WithLogger(logger *logger.Logger) Options {
	return func(o *options) {
		o.Logger = logger
	}
}

func WithAutoLog(autoLog bool) Options {
	return func(o *options) {
		o.AutoLog = autoLog
	}
}

func WithOnStop(onStop []func(elapsed time.Duration)) Options {
	return func(o *options) {
		o.OnStop = onStop
	}
}
