package app

import (
	"context"
	"fmt"

	"github.com/AsaHero/e-wallet/internal/infrastructure/repository"
	"github.com/AsaHero/e-wallet/internal/usecase/jobs"
	"github.com/AsaHero/e-wallet/pkg/app"
	"github.com/AsaHero/e-wallet/pkg/config"
	"github.com/AsaHero/e-wallet/pkg/database/postgres"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/hibiken/asynq"
	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel"
)

type RecordReminderCalculateScheduler struct {
	config       *config.Config
	logger       *logger.Logger
	db           *bun.DB
	taskQueue    *asynq.Client
	shutdownOTLP func(ctx context.Context) error
}

func NewRecordReminderCalculateScheduler(cfg *config.Config) (*RecordReminderCalculateScheduler, error) {
	shutdownOTLP := otlp.InitTracer(
		context.Background(),
		otlp.WithServiceName("record-reminder-calculate-job"),
		otlp.WithEnvironment(cfg.Environment),
		otlp.WithExporterType(otlp.ExporterNameToExporterType[cfg.OTEL.Exporter.Type]),
		otlp.WithEndpoint(cfg.OTEL.Exporter.OTLP.Endpoint),
		otlp.WithExporterProtocol(otlp.ExporterProtocolNameToExporterProtocolType[cfg.OTEL.Exporter.OTLP.Protocol]),
		otlp.WithSamplerType(otlp.SamplerNameToSamplerType[cfg.OTEL.Traces.Sampler]),
		otlp.WithSamplerArg(cfg.OTEL.Traces.SamplerArg),
	)

	logger, err := logger.NewLogger("record-reminder-calculate-job.log", cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	// db config
	db, err := postgres.NewBunDB(
		postgres.WithHost(cfg.DB.Host),
		postgres.WithPort(cfg.DB.Port),
		postgres.WithUser(cfg.DB.User),
		postgres.WithPassword(cfg.DB.Password),
		postgres.WithDB(cfg.DB.Name),
		postgres.WithSSLMode(cfg.DB.Sslmode),
		postgres.WithDebug(cfg.LogLevel == app.Debug),
	)
	if err != nil {
		return nil, fmt.Errorf("error initializing database: %v", err)
	}

	taskQueue := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
		Password: cfg.Redis.Password,
	})

	return &RecordReminderCalculateScheduler{
		config:       cfg,
		logger:       logger,
		db:           db,
		taskQueue:    taskQueue,
		shutdownOTLP: shutdownOTLP,
	}, nil
}

func (a *RecordReminderCalculateScheduler) Run() error {
	// init repository
	usersRepo := repository.NewUsersRepo(a.db)

	// init usecases
	jobsUsecase := jobs.NewModule(a.config.Context.Timeout, a.logger, usersRepo, a.taskQueue)

	ctx, end := otlp.Start(context.Background(), otel.Tracer("RecordReminderCalculate"), "Run")
	defer func() { end(nil) }()

	err := jobsUsecase.RecordReminderCalculateScheduler(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (a *RecordReminderCalculateScheduler) Stop() error {
	if a.db != nil {
		_ = a.db.Close()
	}

	if a.shutdownOTLP != nil {
		_ = a.shutdownOTLP(context.Background())
	}

	if a.logger != nil {
		a.logger.Close()
	}

	if a.taskQueue != nil {
		_ = a.taskQueue.Close()
	}

	return nil
}
