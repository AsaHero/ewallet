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
	"github.com/shogo82148/pointer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/sync/errgroup"
)

type parseImageUsecase struct {
	contextTimeout    time.Duration
	logger            *logger.Logger
	llmClient         ports.LLMProvider
	ocrProvider       ports.OCRProvider
	usersRepo         entities.UserRepository
	accountsRepo      entities.AccountRepository
	categoriesRepo    entities.CategoryRepository
	subcategoriesRepo entities.SubcategoryRepository
	fxRatesProvider   ports.FXRatesProvider
}

func NewParseImageUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	llmClient ports.LLMProvider,
	ocrProvider ports.OCRProvider,
	usersRepo entities.UserRepository,
	accountsRepo entities.AccountRepository,
	categoriesRepo entities.CategoryRepository,
	subcategoriesRepo entities.SubcategoryRepository,
	fxRatesProvider ports.FXRatesProvider,
) *parseImageUsecase {
	return &parseImageUsecase{
		contextTimeout:    timeout,
		logger:            logger,
		llmClient:         llmClient,
		ocrProvider:       ocrProvider,
		usersRepo:         usersRepo,
		accountsRepo:      accountsRepo,
		categoriesRepo:    categoriesRepo,
		subcategoriesRepo: subcategoriesRepo,
		fxRatesProvider:   fxRatesProvider,
	}
}

type ParseImageView struct {
	Type             string     `json:"type"`
	Amount           float64    `json:"amount"`
	Currency         string     `json:"currency,omitempty"`
	OriginalAmount   *float64   `json:"original_amount,omitempty"`
	OriginalCurrency *string    `json:"original_currency,omitempty"`
	FxRate           *float64   `json:"fx_rate,omitempty"`
	AccountID        *string    `json:"account_id,omitempty"`
	CategoryID       *int       `json:"category_id,omitempty"`
	Note             string     `json:"note,omitempty"`
	Confidence       float64    `json:"confidence"`
	PerformedAt      *time.Time `json:"performed_at,omitempty"`
}

func (p *parseImageUsecase) ParseImage(ctx context.Context, userID string, imageURL string) (_ *ParseImageView, err error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("parser"), "ParseImage",
		attribute.String("image_url", imageURL),
	)
	defer func() { end(err) }()

	var input struct {
		userID uuid.UUID
	}
	{
		var err error
		input.userID, err = uuid.Parse(userID)
		if err != nil {
			p.logger.ErrorContext(ctx, "failed to parse user id", err)
			return nil, inerr.NewErrValidation("user_id", "invalud uuid type")
		}
	}

	user, err := p.usersRepo.FindByID(ctx, input.userID)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to get user", err)
		return nil, err
	}

	accounts, err := p.accountsRepo.GetByUserID(ctx, input.userID)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to get accounts", err)
		return nil, err
	}

	// Extract text from image using Vision API
	extractedText, err := p.ocrProvider.ImageToText(ctx, imageURL)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to extract text from image", err)
		return nil, err
	}

	// Generate human readable text from ocr output
	humanreadableText, err := p.llmClient.ChatCompletion(ctx, openai.GPT4o, "", NewOcrParserMessagePrompt(extractedText))
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to generate human readable text", err)
		return nil, err
	}

	// Parallel execution using errgroup
	g, ctx := errgroup.WithContext(ctx)

	var categoryResult struct {
		CategoryID    int     `json:"category_id"`
		SubcategoryID *int    `json:"subcategory_id"`
		Confidence    float64 `json:"confidence"`
	}

	var detailsResult ParseImageView

	// 1. Category Classification
	g.Go(func() error {
		// Fetch categories and subcategories
		categories, err := p.categoriesRepo.FindAll(ctx, input.userID)
		if err != nil {
			p.logger.ErrorContext(ctx, "failed to get categories", err)
			return err
		}

		subcategories, err := p.subcategoriesRepo.FindAll(ctx, input.userID)
		if err != nil {
			p.logger.ErrorContext(ctx, "failed to get subcategories", err)
			return err
		}

		// Build CategoryInfo list
		var catInfos []CategoryInfo
		for _, cat := range categories {
			info := CategoryInfo{
				ID:   cat.ID.Int(),
				Name: cat.NameEN,
			}
			for _, sub := range subcategories {
				if sub.CategoryID == cat.ID.Int() {
					info.Subcategories = append(info.Subcategories, SubcategoryInfo{
						ID:   sub.ID,
						Name: sub.NameEN,
					})
				}
			}
			catInfos = append(catInfos, info)
		}

		prompt := NewCategoryClassificationPrompt(catInfos, humanreadableText, user.LanguageCode.String())
		resp, err := p.llmClient.ChatCompletion(ctx, openai.GPT4o, CategoryClassificationSystemMessage, prompt)
		if err != nil {
			return err
		}

		resp = utils.CleanMarkdownJSON(resp)
		return json.Unmarshal([]byte(resp), &categoryResult)
	})

	// 2. Details Extraction
	g.Go(func() error {
		userPayment := UserPayment{
			Language:    user.LanguageCode.String(),
			Currency:    user.CurrencyCode.String(),
			Timezone:    user.Timezone,
			PaymentText: humanreadableText,
		}

		for _, account := range accounts {
			userPayment.Accounts = append(userPayment.Accounts, UserPaymentAccount{
				ID:   account.ID.String(),
				Name: account.Name,
			})
		}

		prompt := NewTransactionDetailsPrompt(userPayment)
		resp, err := p.llmClient.ChatCompletion(ctx, openai.GPT4o, TransactionDetailsSystemMessage, prompt)
		if err != nil {
			return err
		}

		resp = utils.CleanMarkdownJSON(resp)
		return json.Unmarshal([]byte(resp), &detailsResult)
	})

	if err := g.Wait(); err != nil {
		p.logger.ErrorContext(ctx, "failed to parse image parallel", err)
		return nil, err
	}

	// Merge results
	detailsResult.CategoryID = pointer.Int(categoryResult.CategoryID)
	detailsResult.Confidence = (detailsResult.Confidence + categoryResult.Confidence) / 2

	if detailsResult.OriginalCurrency != nil {
		fxRate, err := p.fxRatesProvider.GetRate(ctx, *detailsResult.OriginalCurrency, user.CurrencyCode.String())
		if err != nil {
			p.logger.ErrorContext(ctx, "failed to get fx rate", err)
			return nil, err
		}
		detailsResult.OriginalAmount = pointer.Float64(detailsResult.Amount)
		detailsResult.Currency = user.CurrencyCode.String()
		detailsResult.Amount = detailsResult.Amount * fxRate
		detailsResult.FxRate = pointer.Float64(fxRate)
	} else {
		detailsResult.Currency = user.CurrencyCode.String()
	}

	return &detailsResult, nil
}
