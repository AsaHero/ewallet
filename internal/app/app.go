package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AsaHero/e-wallet/internal/delivery"
	"github.com/AsaHero/e-wallet/internal/delivery/api"
	"github.com/AsaHero/e-wallet/internal/delivery/api/validation"
	"github.com/AsaHero/e-wallet/internal/delivery/worker"
	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/infrastructure/currency_api"
	"github.com/AsaHero/e-wallet/internal/infrastructure/dictionary"
	"github.com/AsaHero/e-wallet/internal/infrastructure/ocr_service"
	"github.com/AsaHero/e-wallet/internal/infrastructure/openai"
	"github.com/AsaHero/e-wallet/internal/infrastructure/repository"
	"github.com/AsaHero/e-wallet/internal/infrastructure/telegram_bot_service"
	"github.com/AsaHero/e-wallet/internal/usecase/accounts"
	"github.com/AsaHero/e-wallet/internal/usecase/categories"
	"github.com/AsaHero/e-wallet/internal/usecase/notifications"
	"github.com/AsaHero/e-wallet/internal/usecase/parser"
	"github.com/AsaHero/e-wallet/internal/usecase/transactions"
	"github.com/AsaHero/e-wallet/internal/usecase/users"
	"github.com/AsaHero/e-wallet/pkg/app"
	"github.com/AsaHero/e-wallet/pkg/config"
	"github.com/AsaHero/e-wallet/pkg/database/postgres"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/AsaHero/e-wallet/pkg/redis"
	"github.com/hibiken/asynq"
	"github.com/uptrace/bun"
)

type App struct {
	config       *config.Config
	logger       *logger.Logger
	server       *http.Server
	db           *bun.DB
	taskWorker   *asynq.Server
	taskQueue    *asynq.Client
	redis        *redis.RedisClient
	shutdownOTLP func(ctx context.Context) error
}

func New(cfg *config.Config) (*App, error) {
	shutdownOTLP := otlp.InitTracer(
		context.Background(),
		otlp.WithServiceName(cfg.OTEL.ServiceName),
		otlp.WithEnvironment(cfg.Environment),
		otlp.WithExporterType(otlp.ExporterNameToExporterType[cfg.OTEL.Exporter.Type]),
		otlp.WithEndpoint(cfg.OTEL.Exporter.OTLP.Endpoint),
		otlp.WithExporterProtocol(otlp.ExporterProtocolNameToExporterProtocolType[cfg.OTEL.Exporter.OTLP.Protocol]),
		otlp.WithSamplerType(otlp.SamplerNameToSamplerType[cfg.OTEL.Traces.Sampler]),
		otlp.WithSamplerArg(cfg.OTEL.Traces.SamplerArg),
	)

	logger, err := logger.NewLogger(cfg.APP+".log", cfg.LogLevel)
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

	reids, err := redis.New(
		redis.WithAddress(cfg.Redis.Host+":"+cfg.Redis.Port),
		redis.WithPassword(cfg.Redis.Password),
	)
	if err != nil {
		return nil, fmt.Errorf("error initializing redis: %v", err)
	}

	taskQueue := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
		Password: cfg.Redis.Password,
	})

	return &App{
		config:       cfg,
		logger:       logger,
		db:           db,
		taskQueue:    taskQueue,
		redis:        reids,
		shutdownOTLP: shutdownOTLP,
	}, nil
}

func (a *App) Run() error {
	txManager := postgres.NewTxManager(a.db)

	// init provider
	openaiProvider, err := openai.New(a.config)
	if err != nil {
		return fmt.Errorf("failed to create openai provider: %w", err)
	}

	telegramBotService, err := telegram_bot_service.New(a.config)
	if err != nil {
		return fmt.Errorf("failed to create telegram bot service: %w", err)
	}

	ocrProvider, err := ocr_service.New(a.config)
	if err != nil {
		return fmt.Errorf("failed to create ocr provider: %w", err)
	}

	currencyApiClient, err := currency_api.New(a.redis)
	if err != nil {
		return fmt.Errorf("failed to create currency api client: %w", err)
	}

	// init dictionary
	categoriesDict := dictionary.NewCategoriesDict(a.db)
	subcategoriesDict := dictionary.NewSubcategoriesDict(a.db)

	// init repository
	usersRepo := repository.NewUsersRepo(a.db)
	accountsRepo := repository.NewAccountsRepo(a.db)
	transactionsRepo := repository.NewTransactionsRepo(a.db, categoriesDict, subcategoriesDict)

	// domain services
	accountsDomainService := entities.NewAccountsService(accountsRepo)

	// init usecases
	usersUsecase := users.NewModule(a.config.Context.Timeout, a.logger, usersRepo)
	accountsUsecase := accounts.NewModule(a.config.Context.Timeout, a.logger, usersRepo, accountsRepo, accountsDomainService, transactionsRepo, categoriesDict)
	transactionsUsecase := transactions.NewModule(a.config.Context.Timeout, a.logger, txManager, usersRepo, accountsRepo, transactionsRepo, categoriesDict, subcategoriesDict)
	categoriesUsecase := categories.NewModule(a.config.Context.Timeout, a.logger, categoriesDict, subcategoriesDict, usersRepo)
	parserUsecase := parser.NewModule(a.logger, openaiProvider, ocrProvider, usersRepo, accountsRepo, currencyApiClient)
	notificationsUsecase := notifications.NewModule(a.logger, transactionsRepo, usersRepo, a.taskQueue, telegramBotService)

	// init handlers
	opts := &delivery.Options{
		Config:              a.config,
		Validator:           validation.NewValidator(),
		Logger:              a.logger,
		UsersUsecase:        usersUsecase,
		AccountsUsecase:     accountsUsecase,
		TransactionsUsecase: transactionsUsecase,
		CategoriesUsecase:   categoriesUsecase,
		ParserUsecase:       parserUsecase,
		NotificationUsecase: notificationsUsecase,
	}

	mux := worker.NewRouter(opts)
	a.taskWorker = worker.NewWorker(a.config)
	go a.taskWorker.Run(mux)

	router := api.NewRouter(opts)
	a.server = api.NewServer(a.config, router)

	a.logger.Info("Listen http server:", "address", a.config.Server.Host+":"+a.config.Server.Port)
	return a.server.ListenAndServe()
}

func (a *App) Stop() error {
	if a.server != nil {
		_ = a.server.Shutdown(context.Background())
	}

	if a.db != nil {
		a.db.Close()
	}

	if a.taskWorker != nil {
		a.taskWorker.Stop()
	}

	if a.taskQueue != nil {
		_ = a.taskQueue.Close()
	}

	if a.shutdownOTLP != nil {
		a.shutdownOTLP(context.Background())
	}

	if a.logger != nil {
		a.logger.Close()
	}

	return nil
}
