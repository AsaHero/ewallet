package kafka_consumer

const (
	minBytes = 10e3 // 10KB
	maxBytes = 10e6 // 10MB
)

type Options struct {
	// Broker options
	BootstrapAddrs []string

	// Transport options
	SASLMechanism    string
	SecurityProtocol string
	Username         string
	Password         string
	RootCA           string

	// Reader options
	MaxBytes int
	MinBytes int
}

type Option func(*Options)

func WithBootstrapAddrs(addrs []string) Option {
	return func(o *Options) {
		o.BootstrapAddrs = addrs
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

func WithMaxBytes(bytes int) Option {
	return func(o *Options) {
		o.MaxBytes = bytes
	}
}

func WithMinBytes(bytes int) Option {
	return func(o *Options) {
		o.MinBytes = bytes
	}
}
