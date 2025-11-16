package kafka_producer

import (
	"time"

	"github.com/segmentio/kafka-go"
)

type Options struct {
	// Broker options
	BootstrapAddrs         []string
	RequiredAcks           kafka.RequiredAcks
	Async                  bool
	AllowAutoTopicCreation bool

	// Transport options
	SASLMechanism    string
	SecurityProtocol string
	Username         string
	Password         string
	RootCA           string

	// Writer options
	TImeout time.Duration
}

type Option func(*Options)

func WithBootstrapAddrs(addrs []string) Option {
	return func(o *Options) {
		o.BootstrapAddrs = addrs
	}
}

func WithRequiredAcks(requiredAcks kafka.RequiredAcks) Option {
	return func(o *Options) {
		o.RequiredAcks = requiredAcks
	}
}

func WithAsync(async bool) Option {
	return func(o *Options) {
		o.Async = async
	}
}

func WithAllowAutoTopicCreation(allowAutoTopicCreation bool) Option {
	return func(o *Options) {
		o.AllowAutoTopicCreation = allowAutoTopicCreation
	}
}

func WithSASLMechanism(mechanism string) Option {
	return func(o *Options) {
		o.SASLMechanism = mechanism
	}
}

func WithSecurityProtocol(protocol string) Option {
	return func(o *Options) {
		o.SecurityProtocol = protocol
	}
}

func WithTLS(username, password, rootCAPath string) Option {
	return func(o *Options) {
		o.Username = username
		o.Password = password
		o.RootCA = rootCAPath
	}
}
