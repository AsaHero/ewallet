package broker

import (
	"context"
)

type Event interface {
	GetID() string
	GetTopic() string
	GetPayload() []byte
	GetHeaders() map[string]string
}

type ConsumerConfig interface {
	GetBrokers() []string
	GetTopic() string
	GetGroupID() string
	GetHandler() func(ctx context.Context, event Event) error
}

type Consumer interface {
	Run(ctx context.Context)
	Subscribe(config ConsumerConfig)
	Stop()
}

type Producer interface {
	Publish(ctx context.Context, event Event) error
	Close()
}
