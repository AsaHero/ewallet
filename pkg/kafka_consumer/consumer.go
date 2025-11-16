package kafka_consumer

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/AsaHero/e-wallet/internal/usecase/broker"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/AsaHero/e-wallet/pkg/retry"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/scram"

	"go.opentelemetry.io/otel"
)

type consumer struct {
	options        *Options
	logger         *logger.Logger
	consumerConfig []broker.ConsumerConfig
	readers        []*kafka.Reader
	wg             sync.WaitGroup

	retryCfg retry.RetryConfig
	cancel   context.CancelFunc
}

func New(logger *logger.Logger, opts ...Option) *consumer {
	o := &Options{
		BootstrapAddrs:   make([]string, 0),
		SASLMechanism:    "",
		SecurityProtocol: "",
		Username:         "",
		Password:         "",
		RootCA:           "",
		MaxBytes:         maxBytes,
		MinBytes:         minBytes,
	}

	for _, opt := range opts {
		opt(o)
	}

	return &consumer{
		options:        o,
		logger:         logger,
		consumerConfig: make([]broker.ConsumerConfig, 0),
		readers:        make([]*kafka.Reader, 0),
		retryCfg:       retry.DefaultConfig(),
	}
}

func (c *consumer) Subscribe(conf broker.ConsumerConfig) {
	c.consumerConfig = append(c.consumerConfig, conf)
}

func (c *consumer) Run(ctx context.Context) {
	// derive cancelable context for all workers
	ctx, c.cancel = context.WithCancel(ctx)

	dialer, err := c.buildDialer()
	if err != nil {
		c.logger.With("component", "kafka-consumer").Error("failed to build kafka dialer", "error", err)
		// You can choose to return here to fail-fast:
		// return
		// Or continue without auth (not recommended):
		dialer = &kafka.Dialer{Timeout: 10 * time.Second, DualStack: true}
	}

	for _, conf := range c.consumerConfig {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:               conf.GetBrokers(),
			GroupID:               conf.GetGroupID(),
			Topic:                 conf.GetTopic(),
			MinBytes:              c.options.MinBytes,
			MaxBytes:              c.options.MaxBytes,
			Dialer:                dialer,
			WatchPartitionChanges: true,
		})
		c.readers = append(c.readers, reader)
		c.worker(ctx, reader, conf)
	}
}

func (c *consumer) Stop() {
	// cancel first so any blocking FetchMessage unblocks
	if c.cancel != nil {
		c.cancel()
	}

	for _, r := range c.readers {
		_ = r.Close()
	}
	c.wg.Wait()
}

func (c *consumer) worker(ctx context.Context, r *kafka.Reader, conf broker.ConsumerConfig) {
	c.wg.Add(1)

	go func() {
		defer c.wg.Done()

		topic := conf.GetTopic()
		handler := conf.GetHandler()
		consumerName := fmt.Sprintf("ConsumeFromTopic:%s", conf.GetTopic())

		for {
			var err error
			var end func(err error)

			msg, err := r.FetchMessage(ctx)
			if err != nil {
				// treat terminal conditions as exit
				if errors.Is(err, context.Canceled) ||
					errors.Is(err, kafka.ErrGroupClosed) ||
					errors.Is(err, io.EOF) {
					c.logger.Info("stopping reader", "reason", err)
					return
				}

				c.logger.Error("fetch error", "error", err)

				// exit if shutting down
				select {
				case <-ctx.Done():
					return
				case <-time.After(300 * time.Millisecond):
					continue
				}
			}

			traceId, spanId, err := getTraceAndSpanId(msg)
			if err != nil {
				c.logger.Error(
					"failed to get span_id or trace_id of a kafka message",
					"error", err,
					slog.String("topic", topic),
					slog.String("value", string(msg.Value)),
				)
			}

			ctxOtlp, _, err := otlp.RestoreTraceContext(traceId, spanId)
			if err != nil {
				c.logger.Error(
					"failed to form context from trace_id and span_id",
					"error", err,
					slog.String("topic", topic),
					slog.String("value", string(msg.Value)),
				)
			} else {
				ctx, end = otlp.Start(ctxOtlp, otel.Tracer("KafkaConsumer"), consumerName)
			}

			evt := toEvent(msg)

			// bounded retries around handler
			_, hErr := retry.Retry(ctx, c.retryCfg, func(ctx context.Context) (struct{}, error) {
				return struct{}{}, handler(ctx, evt)
			})
			if hErr != nil {
				c.logger.Error("handler error",
					slog.Int("partition", msg.Partition),
					slog.Int64("offset", msg.Offset),
					"error", hErr,
				)
				// do not commit on handler error → redelivery
				if end != nil {
					end(hErr)
				}
				continue
			}

			if err := r.CommitMessages(ctx, msg); err != nil {
				// Commit can also surface after Close()/cancel; treat canceled as exit
				if errors.Is(err, context.Canceled) {
					c.logger.Info("stopping reader", "reason", "commit canceled")
					return
				}
				c.logger.Error("commit error",
					slog.Int("partition", msg.Partition),
					slog.Int64("offset", msg.Offset),
					"error", err,
				)
			}

			if end != nil {
				end(err)
			}
		}
	}()
}

func getTraceAndSpanId(msg kafka.Message) (string, string, error) {
	var (
		spanId, traceId string
	)

	for _, header := range msg.Headers {
		switch header.Key {
		case "trace_id":
			traceId = string(header.Value)
		case "span_id":
			spanId = string(header.Value)
		}
	}

	if len(traceId) == 0 {
		return "", "", errors.New("missing trace_id field in kafka message header")
	}

	if len(spanId) == 0 {
		return "", "", errors.New("missing span_id field in kafka message header")
	}

	return traceId, spanId, nil
}

func (c *consumer) buildDialer() (*kafka.Dialer, error) {
	var (
		sasl sasl.Mechanism
		err  error
	)

	user := strings.TrimSpace(c.options.Username)
	pass := c.options.Password

	// SASL mechanism (optional; only if username provided)
	if user != "" {
		switch strings.ToUpper(strings.TrimSpace(c.options.SASLMechanism)) {
		case "SCRAM-SHA-512", "SCRAM_SHA_512":
			sasl, err = scram.Mechanism(scram.SHA512, user, pass)
		case "SCRAM-SHA-256", "SCRAM_SHA_256":
			sasl, err = scram.Mechanism(scram.SHA256, user, pass)
		default:
			// default to SCRAM-SHA-512
			sasl, err = scram.Mechanism(scram.SHA512, user, pass)
		}
		if err != nil {
			return nil, err
		}
	}

	// TLS only if SASL_SSL/SSL
	var tlsCfg *tls.Config
	switch strings.ToUpper(strings.TrimSpace(c.options.SecurityProtocol)) {
	case "SASL_SSL", "SSL":
		tlsCfg = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: false,
		}

		if c.options.RootCA != "" {
			tlsCfg.RootCAs = x509.NewCertPool()

			pem, err := os.ReadFile(c.options.RootCA)
			if err != nil {
				return nil, err
			}
			tlsCfg.RootCAs.AppendCertsFromPEM(pem)
		}
	default:
		// SASL_PLAINTEXT / PLAINTEXT → no TLS
	}

	return &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		SASLMechanism: sasl,   // nil if no SASL
		TLS:           tlsCfg, // nil for PLAINTEXT/SASL_PLAINTEXT
	}, nil
}

// Event mapping
type event struct {
	id      string
	topic   string
	payload []byte
	headers map[string]string
}

func (e event) GetID() string                 { return e.id }
func (e event) GetTopic() string              { return e.topic }
func (e event) GetPayload() []byte            { return e.payload }
func (e event) GetHeaders() map[string]string { return e.headers }

func toEvent(m kafka.Message) broker.Event {
	// map headers
	h := make(map[string]string, len(m.Headers))
	for _, kv := range m.Headers {
		h[kv.Key] = string(kv.Value)
	}

	// derive ID: key → common header → composite fallback
	id := string(m.Key)
	if id == "" {
		if v, ok := h["event-id"]; ok {
			id = v
		} else if v, ok := h["x-event-id"]; ok {
			id = v
		} else if v, ok := h["id"]; ok {
			id = v
		} else {
			id = fmt.Sprintf("%s:%d:%d", m.Topic, m.Partition, m.Offset)
		}
	}

	return event{
		id:      id,
		topic:   m.Topic,
		payload: m.Value,
		headers: h,
	}
}
