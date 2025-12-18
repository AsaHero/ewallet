package parser

import (
	"context"
	"encoding/json"
	"sync"
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
	AccountID        *string    `json:"account_id,omitempty"`
	Type             string     `json:"type"`
	Amount           float64    `json:"amount"`
	Currency         string     `json:"currency,omitempty"`
	OriginalAmount   *float64   `json:"original_amount,omitempty"`
	OriginalCurrency *string    `json:"original_currency,omitempty"`
	FxRate           *float64   `json:"fx_rate,omitempty"`
	CategoryID       *int       `json:"category_id,omitempty"`
	SubcategoryID    *int       `json:"subcategory_id,omitempty"`
	Note             string     `json:"note,omitempty"`
	PerformedAt      *time.Time `json:"performed_at,omitempty"`
	Confidence       float64    `json:"confidence"`
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

	var categoryResult CategoryClassificationResult
	var detailsResult TransactionDetailsResult
	var wg sync.WaitGroup
	var errChan = make(chan error, 2)

	// 1. Category Classification
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Fetch categories and subcategories
		categories, err := p.categoriesRepo.FindAll(ctx, input.userID)
		if err != nil {
			p.logger.ErrorContext(ctx, "failed to get categories", err)
			errChan <- err
			return
		}

		subcategories, err := p.subcategoriesRepo.FindAll(ctx, input.userID)
		if err != nil {
			p.logger.ErrorContext(ctx, "failed to get subcategories", err)
			errChan <- err
			return
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
			p.logger.ErrorContext(ctx, "failed to get categories", err)
			errChan <- err
			return
		}

		resp = utils.CleanMarkdownJSON(resp)
		if err := json.Unmarshal([]byte(resp), &categoryResult); err != nil {
			p.logger.ErrorContext(ctx, "failed to parse categories", err)
			errChan <- err
			return
		}
	}()

	// 2. Details Extraction
	wg.Add(1)
	go func() {
		defer wg.Done()
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
			p.logger.ErrorContext(ctx, "failed to get details", err)
			errChan <- err
			return
		}

		resp = utils.CleanMarkdownJSON(resp)
		if err := json.Unmarshal([]byte(resp), &detailsResult); err != nil {
			p.logger.ErrorContext(ctx, "failed to parse details", err)
			errChan <- err
			return
		}
	}()

	wg.Wait()
	if len(errChan) > 0 {
		return nil, <-errChan
	}

	// Merge results
	var result = &ParseImageView{
		Type:          detailsResult.Type,
		AccountID:     detailsResult.AccountID,
		Note:          detailsResult.Note,
		PerformedAt:   detailsResult.PerformedAt,
		CategoryID:    categoryResult.CategoryID,
		SubcategoryID: categoryResult.SubcategoryID,
		Confidence:    (detailsResult.Confidence + categoryResult.Confidence) / 2,
	}

	if detailsResult.Currency != user.CurrencyCode.String() {
		fxRate, err := p.fxRatesProvider.GetRate(ctx, detailsResult.Currency, user.CurrencyCode.String())
		if err != nil {
			p.logger.ErrorContext(ctx, "failed to get fx rate", err)
			return nil, err
		}

		result.Amount = detailsResult.Amount * fxRate
		result.Currency = user.CurrencyCode.String()
		result.OriginalAmount = pointer.Float64(detailsResult.Amount)
		result.OriginalCurrency = pointer.String(detailsResult.Currency)
		result.FxRate = pointer.Float64(fxRate)
	} else {
		result.Amount = detailsResult.Amount
		result.Currency = user.CurrencyCode.String()
	}

	return result, nil
}
