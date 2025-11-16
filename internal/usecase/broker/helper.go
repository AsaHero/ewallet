package broker

import (
	"context"
)

type consumerConfig struct {
	brokers []string
	topic   string
	groupID string
	handler func(ctx context.Context, event Event) error
}

func NewConsumerConfig(
	brokers []string,
	topic string,
	groupID string,
	handler func(ctx context.Context, event Event) error,
) ConsumerConfig {
	return &consumerConfig{
		brokers: brokers,
		topic:   topic,
		groupID: groupID,
		handler: handler,
	}
}

func (c *consumerConfig) GetBrokers() []string {
	return c.brokers
}

func (c *consumerConfig) GetTopic() string {
	return c.topic
}

func (c *consumerConfig) GetGroupID() string {
	return c.groupID
}

func (c *consumerConfig) GetHandler() func(ctx context.Context, event Event) error {
	return c.handler
}
