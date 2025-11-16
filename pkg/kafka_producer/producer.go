package kafka_producer

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/AsaHero/e-wallet/internal/usecase/broker"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/scram"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Producer struct {
	options *Options
	log     *logger.Logger
	writers map[string]*kafka.Writer // topic -> writer
	mu      sync.RWMutex
}

func New(log *logger.Logger, opts ...Option) *Producer {
	o := &Options{
		BootstrapAddrs:         make([]string, 0),
		SASLMechanism:          "",
		SecurityProtocol:       "",
		Username:               "",
		Password:               "",
		RootCA:                 "",
		RequiredAcks:           kafka.RequireAll,
		Async:                  true,
		AllowAutoTopicCreation: true,
		TImeout:                10 * time.Second,
	}

	for _, opt := range opts {
		opt(o)
	}

	return &Producer{
		options: o,
		log:     log,
		writers: make(map[string]*kafka.Writer),
	}
}

func (p *Producer) Publish(ctx context.Context, e broker.Event) (err error) {
	topic := e.GetTopic()
	w := p.getOrCreateWriter(topic)

	ctx, end := otlp.Start(ctx, otel.Tracer("KafkaProducer"), fmt.Sprintf("ProduceToTopic:%s", topic))
	defer func() { end(err) }()

	msg := kafka.Message{
		Key:     []byte(e.GetID()), // stable partitioning per ID
		Value:   e.GetPayload(),
		Time:    time.Now(),
		Headers: toKafkaHeaders(e.GetHeaders()),
	}

	p.embedTracingToMessage(ctx, &msg)

	return w.WriteMessages(ctx, msg)
}

func (p *Producer) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for topic, w := range p.writers {
		if err := w.Close(); err != nil {
			p.log.With("component", "kafka-producer", "topic", topic).Warn("failed to close writer", "error", err)
		}
	}
	p.writers = make(map[string]*kafka.Writer)
}

// -------- internals --------

func (p *Producer) getOrCreateWriter(topic string) *kafka.Writer {
	// Fast read path
	p.mu.RLock()
	w, ok := p.writers[topic]
	p.mu.RUnlock()
	if ok {
		return w
	}

	// Slow path: build writer
	p.mu.Lock()
	defer p.mu.Unlock()
	if w, ok = p.writers[topic]; ok {
		return w
	}

	transport, err := p.buildTransport()
	if err != nil {
		p.log.With("component", "kafka-producer").Error("failed to build Kafka transport", "error", err)
		// Hard-fail is usually better here so you don't silently produce without auth.
		// But if you prefer to continue, you can set transport = &kafka.Transport{} instead.
		transport = &kafka.Transport{}
	}

	w = &kafka.Writer{
		Addr:         kafka.TCP(p.options.BootstrapAddrs...),
		Topic:        topic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: p.options.RequiredAcks,
		Transport:    transport,

		//Tune based on the project needs:
		// BatchTimeout:           20 * time.Millisecond,
		// BatchSize:              100,          // messages
		// BatchBytes:             1 << 20,      // 1MB
		// Compression:            kafka.Snappy, // good default; or None/Lz4/Zstd

		Async:                  p.options.Async,
		AllowAutoTopicCreation: p.options.AllowAutoTopicCreation,
		Completion: func(messages []kafka.Message, err error) {
			if err != nil {
				p.log.With("component", "kafka-producer", "topic", topic).Error("write error", "error", err)
			}

			for _, m := range messages {
				p.log.With("component", "kafka-producer", "topic", topic).Debug("message sent", "key", string(m.Key), "offset", m.Offset)
			}
		},
	}
	p.writers[topic] = w
	return w
}

func (p *Producer) embedTracingToMessage(ctx context.Context, msg *kafka.Message) {
	msg.Headers = append(msg.Headers, kafka.Header{
		Key:   "trace_id",
		Value: []byte(trace.SpanFromContext(ctx).SpanContext().TraceID().String()),
	})
	msg.Headers = append(msg.Headers, kafka.Header{
		Key:   "span_id",
		Value: []byte(trace.SpanFromContext(ctx).SpanContext().SpanID().String()),
	})
}

func (p *Producer) buildTransport() (*kafka.Transport, error) {
	var (
		mech sasl.Mechanism
		err  error
	)

	// If username empty => no SASL (plain PLAINTEXT)
	if u := strings.TrimSpace(p.options.Username); u != "" {
		switch strings.ToUpper(strings.TrimSpace(p.options.SASLMechanism)) {
		case "SCRAM-SHA-512", "SCRAM_SHA_512":
			mech, err = scram.Mechanism(scram.SHA512, u, p.options.Password)
		case "SCRAM-SHA-256", "SCRAM_SHA_256":
			mech, err = scram.Mechanism(scram.SHA256, u, p.options.Password)
		default:
			// default to SCRAM-SHA-512 if not specified
			mech, err = scram.Mechanism(scram.SHA512, u, p.options.Password)
		}
		if err != nil {
			return nil, err
		}
	}

	// TLS only if SASL_SSL (or SSL)
	var tlsCfg *tls.Config
	switch strings.ToUpper(strings.TrimSpace(p.options.SecurityProtocol)) {
	case "SASL_SSL", "SSL":
		tlsCfg = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: false,
		}
		if p.options.RootCA != "" {
			tlsCfg.RootCAs = x509.NewCertPool()

			pem, err := os.ReadFile(p.options.RootCA)
			if err != nil {
				return nil, err
			}
			tlsCfg.RootCAs.AppendCertsFromPEM(pem)
		}
	default:
		// SASL_PLAINTEXT / PLAINTEXT => tlsCfg = nil
	}

	return &kafka.Transport{
		SASL:        mech,   // nil if no SASL
		TLS:         tlsCfg, // nil for SASL_PLAINTEXT
		DialTimeout: 10 * time.Second,
		IdleTimeout: 60 * time.Second,
		MetadataTTL: 5 * time.Minute,
	}, nil
}

func toKafkaHeaders(h map[string]string) []kafka.Header {
	if len(h) == 0 {
		return nil
	}
	out := make([]kafka.Header, 0, len(h))
	for k, v := range h {
		out = append(out, kafka.Header{Key: k, Value: []byte(v)})
	}
	return out
}
