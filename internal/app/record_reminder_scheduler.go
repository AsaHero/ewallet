package app

import (
	"context"
	"fmt"

	"github.com/AsaHero/e-wallet/internal/tasks"
	"github.com/AsaHero/e-wallet/pkg/config"
	"github.com/AsaHero/e-wallet/pkg/database/postgres"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/hibiken/asynq"
	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel"
)

type RecordReminderScheduler struct {
	config       *config.Config
	logger       *logger.Logger
	db           *bun.DB
	scheduler    *asynq.Scheduler
	taskQueue    *asynq.Client
	shutdownOTLP func(ctx context.Context) error
}

func NewRecordReminderScheduler(cfg *config.Config) (*RecordReminderScheduler, error) {
	shutdownOTLP := otlp.InitTracer(
		context.Background(),
		otlp.WithServiceName("record-reminder-scheduler"),
		otlp.WithEnvironment(cfg.Environment),
		otlp.WithExporterType(otlp.ExporterNameToExporterType[cfg.OTEL.Exporter.Type]),
		otlp.WithEndpoint(cfg.OTEL.Exporter.OTLP.Endpoint),
		otlp.WithExporterProtocol(otlp.ExporterProtocolNameToExporterProtocolType[cfg.OTEL.Exporter.OTLP.Protocol]),
		otlp.WithSamplerType(otlp.SamplerNameToSamplerType[cfg.OTEL.Traces.Sampler]),
		otlp.WithSamplerArg(cfg.OTEL.Traces.SamplerArg),
	)

	logger, err := logger.NewLogger("record-reminder-scheduler.log", cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	db, err := postgres.NewBunDB(
		postgres.WithHost(cfg.DB.Host),
		postgres.WithPort(cfg.DB.Port),
		postgres.WithUser(cfg.DB.User),
		postgres.WithPassword(cfg.DB.Password),
		postgres.WithDB(cfg.DB.Name),
		postgres.WithSSLMode(cfg.DB.Sslmode),
	)
	if err != nil {
		return nil, fmt.Errorf("error initializing database: %v", err)
	}

	scheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{
			Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
			Password: cfg.Redis.Password,
		},
		&asynq.SchedulerOpts{
			Location: nil,
		},
	)

	taskQueue := asynq.NewClient(
		asynq.RedisClientOpt{
			Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
			Password: cfg.Redis.Password,
		},
	)

	return &RecordReminderScheduler{
		config:       cfg,
		logger:       logger,
		db:           db,
		scheduler:    scheduler,
		taskQueue:    taskQueue,
		shutdownOTLP: shutdownOTLP,
	}, nil
}

func (a *RecordReminderScheduler) Run(runNow bool) (err error) {
	ctx := context.Background()

	ctx, end := otlp.Start(ctx, otel.Tracer("record-reminder-scheduler"), "Run")
	defer func() { end(err) }()

	task, err := tasks.NewRecordReminderScheduleTask()
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to create record reminder schedule task", err)
		return err
	}

	if runNow {
		_, err = a.taskQueue.Enqueue(task, asynq.Queue("medium"))
		if err != nil {
			a.logger.ErrorContext(ctx, "failed to enqueue record reminder schedule task", err)
			return err
		}

		return nil
	}

	_, err = a.scheduler.Register("0 0 * * 1-5", task, asynq.Queue("medium"))
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to register record reminder schedule task", err)
		return err
	}

	a.logger.InfoContext(ctx, "record reminder schedule task registered")
	a.logger.InfoContext(ctx, "Schdules record remiders at 00:00 every day from Monday to Friday")

	return a.scheduler.Run()
}

func (a *RecordReminderScheduler) Stop() error {
	if a.scheduler != nil {
		a.scheduler.Shutdown()
	}

	if a.db != nil {
		_ = a.db.Close()
	}

	if a.shutdownOTLP != nil {
		_ = a.shutdownOTLP(context.Background())
	}

	if a.logger != nil {
		a.logger.Close()
	}

	return nil
}
