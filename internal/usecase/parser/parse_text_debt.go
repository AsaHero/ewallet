package parser

import (
	"context"
	"encoding/json"
	"time"

	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/inerr"
	"github.com/AsaHero/e-wallet/internal/usecase/ports"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/AsaHero/e-wallet/pkg/utils"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type parseTextDebtUsecase struct {
	contextTimeout   time.Duration
	logger           *logger.Logger
	llmClient        ports.LLMProvider
	usersRepo        entities.UserRepository
	transactionsRepo entities.TransactionRepository
}

func NewParseTextDebtUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	llmClient ports.LLMProvider,
	usersRepo entities.UserRepository,
	transactionsRepo entities.TransactionRepository,
) *parseTextDebtUsecase {
	return &parseTextDebtUsecase{
		contextTimeout:   timeout,
		logger:           logger,
		llmClient:        llmClient,
		usersRepo:        usersRepo,
		transactionsRepo: transactionsRepo,
	}
}

type ParseTextDebtView struct {
	Type             string  `json:"type"`
	Amount           float64 `json:"amount"`
	Currency         string  `json:"currency"`
	CounterpartyName string  `json:"counterparty_name"`
	Note             string  `json:"note"`
	DueDate          string  `json:"due_date"`
	Confidence       float64 `json:"confidence"`
}

type DebtDetailsResult struct {
	CounterpartyName string  `json:"counterparty_name"`
	DueDate          string  `json:"due_date"`
	Confidence       float64 `json:"confidence"`
}

func (p *parseTextDebtUsecase) ParseTextDebt(ctx context.Context, userID string, transactionID string, text string) (_ *ParseTextDebtView, err error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("parser"), "ParseTextDebt",
		attribute.String("user_id", userID),
		attribute.String("text", text),
	)
	defer func() { end(err) }()

	var input struct {
		userID        uuid.UUID
		transactionID uuid.UUID
	}
	{
		var err error
		input.userID, err = uuid.Parse(userID)
		if err != nil {
			p.logger.ErrorContext(ctx, "failed to parse user id", err)
			return nil, inerr.NewErrValidation("user_id", "invalud uuid type")
		}

		input.transactionID, err = uuid.Parse(transactionID)
		if err != nil {
			p.logger.ErrorContext(ctx, "failed to parse transaction id", err)
			return nil, inerr.NewErrValidation("transaction_id", "invalud uuid type")
		}
	}

	user, err := p.usersRepo.FindByID(ctx, input.userID)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to get user", err)
		return nil, err
	}

	transaction, err := p.transactionsRepo.GetByID(ctx, input.transactionID)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to get transaction", err)
		return nil, err
	}

	var detailsResult DebtDetailsResult
	resp, err := p.llmClient.ChatCompletion(ctx, openai.GPT4o, DebtDetailsSystemMessage, NewDebtDetailsPrompt(text, user.LanguageCode.String(), user.Timezone, time.Now().UTC()))
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to get debt details", err)
		return nil, err
	}

	resp = utils.CleanMarkdownJSON(resp)
	if err := json.Unmarshal([]byte(resp), &detailsResult); err != nil {
		p.logger.ErrorContext(ctx, "failed to parse debt details", err)
		return nil, err
	}

	return &ParseTextDebtView{
		Type:             utils.If(transaction.Type == entities.Deposit, entities.Borrow.String(), entities.Lend.String()),
		Amount:           entities.MajorFromMinor(transaction.Amount, transaction.CurrencyCode.Scale()),
		Currency:         transaction.CurrencyCode.String(),
		CounterpartyName: detailsResult.CounterpartyName,
		Note:             transaction.RowText,
		DueDate:          detailsResult.DueDate,
		Confidence:       detailsResult.Confidence,
	}, nil
}
