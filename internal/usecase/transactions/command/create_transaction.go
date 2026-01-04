package command

import (
	"context"
	"time"

	"github.com/AsaHero/e-wallet/internal/delivery/api/models"
	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/inerr"
	"github.com/AsaHero/e-wallet/pkg/database/postgres"
	"github.com/AsaHero/e-wallet/pkg/logger"
	"github.com/AsaHero/e-wallet/pkg/otlp"
	"github.com/google/uuid"
	"github.com/shogo82148/pointer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type CreateTransactionUsecase struct {
	contextTimeout    time.Duration
	logger            *logger.Logger
	txManager         postgres.TxManager
	usersRepo         entities.UserRepository
	accountsRepo      entities.AccountRepository
	accountsService   *entities.AccountsService
	transactionsRepo  entities.TransactionRepository
	categoryRepo      entities.CategoryRepository
	subcategoriesRepo entities.SubcategoryRepository
	debtsRepo         entities.DebtRepository
}

func NewCreateTransactionUsecase(
	timeout time.Duration,
	logger *logger.Logger,
	txManager postgres.TxManager,
	usersRepo entities.UserRepository,
	accountsRepo entities.AccountRepository,
	accountsService *entities.AccountsService,
	transactionsRepo entities.TransactionRepository,
	categoriesRepo entities.CategoryRepository,
	subcategoriesRepo entities.SubcategoryRepository,
	debtsRepo entities.DebtRepository,
) *CreateTransactionUsecase {
	return &CreateTransactionUsecase{
		contextTimeout:    timeout,
		usersRepo:         usersRepo,
		accountsRepo:      accountsRepo,
		accountsService:   accountsService,
		transactionsRepo:  transactionsRepo,
		categoryRepo:      categoriesRepo,
		subcategoriesRepo: subcategoriesRepo,
		debtsRepo:         debtsRepo,
		logger:            logger,
		txManager:         txManager,
	}
}

type CreateTransactionCommand struct {
	UserID               string
	AccountID            string
	Type                 string
	Amount               float64
	CurrencyCode         string
	OriginalAmount       *float64
	OriginalCurrencyCode *string
	FxRate               *float64
	CategoryID           *int
	SubcategoryID        *int
	Note                 string
	PerformedAt          *time.Time
}

func (c *CreateTransactionUsecase) CreateTransaction(ctx context.Context, cmd *CreateTransactionCommand) (_ *models.CreateTransactionResponse, err error) {
	ctx, cancel := context.WithTimeout(ctx, c.contextTimeout)
	defer cancel()

	ctx, end := otlp.Start(ctx, otel.Tracer("transactions"), "CreateTransaction",
		attribute.String("user_id", cmd.UserID),
		attribute.String("account_id", cmd.AccountID),
	)
	defer func() { end(err) }()

	var input struct {
		userID      uuid.UUID
		accountID   uuid.UUID
		category    *entities.Category
		subcategory *entities.Subcategory
		trnType     entities.TrnType
	}
	{
		var err error
		input.userID, err = uuid.Parse(cmd.UserID)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to parse user id", err)
			return nil, inerr.NewErrValidation("user_id", "invalud uuid type")
		}

		input.accountID, err = uuid.Parse(cmd.AccountID)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to parse account id", err)
			return nil, inerr.NewErrValidation("account_id", "invalud uuid type")
		}

		if cmd.CategoryID != nil {
			category, err := c.categoryRepo.FindByID(ctx, *cmd.CategoryID)
			if err != nil {
				c.logger.ErrorContext(ctx, "failed to get category", err)
				return nil, err
			}

			input.category = category
		}

		if cmd.SubcategoryID != nil {
			subcategory, err := c.subcategoriesRepo.FindByID(ctx, *cmd.SubcategoryID)
			if err != nil {
				c.logger.ErrorContext(ctx, "failed to get subcategory", err)
				return nil, err
			}

			input.subcategory = subcategory
		}

		if cmd.Type == "deposit" {
			input.trnType = entities.Deposit
		} else {
			input.trnType = entities.Withdrawal
		}
	}

	user, err := c.usersRepo.FindByID(ctx, input.userID)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to get user", err)
		return nil, err
	}

	var transaction *entities.Transaction
	var debt *entities.Debt
	err = c.txManager.WithTx(ctx, func(ctx context.Context) error {
		account, err := c.accountsRepo.GetByIDForUpdate(ctx, input.accountID)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to get account", err)
			return err
		}

		transaction, err = entities.NewTransaction(
			user.ID,
			account.ID,
			input.trnType,
			cmd.Note,
		)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to create transaction", err)
			return err
		}

		err = transaction.Categorise(input.category, input.subcategory)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to categorise transaction", err)
			return err
		}

		err = transaction.SetAmountMajor(cmd.Amount, user.CurrencyCode)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to set amount major", err)
			return err
		}

		if cmd.OriginalAmount != nil {
			err = transaction.SetOriginalAmountMajor(*cmd.OriginalAmount, entities.Currency(*cmd.OriginalCurrencyCode))
			if err != nil {
				c.logger.ErrorContext(ctx, "failed to set original amount major", err)
				return err
			}

			if cmd.FxRate != nil {
				err = transaction.SetFxRate(*cmd.FxRate)
				if err != nil {
					c.logger.ErrorContext(ctx, "failed to set fx rate", err)
					return err
				}
			}
		}

		if cmd.PerformedAt != nil {
			transaction.Performed(*cmd.PerformedAt)
		} else {
			transaction.Performed(time.Now())
		}

		err = c.transactionsRepo.Save(ctx, transaction)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to create transaction", err)
			return err
		}

		err = c.accountsService.ApplyTransaction(ctx, account, transaction)
		if err != nil {
			c.logger.ErrorContext(ctx, "failed to apply transaction", err)
			return err
		}

		// Create debt if category is "Loans & Debts" (ID 26)
		if input.category != nil && input.category.ID.Int() == 26 {
			// Determine debt type based on transaction type
			var debtType entities.DebtType
			if input.trnType == entities.Withdrawal {
				debtType = entities.Borrow // User borrowed money (withdrawal = spending)
			} else {
				debtType = entities.Lend // User lent money (deposit = receiving)
			}

			debt, err = entities.NewDebt(user.ID, transaction.ID, debtType)
			if err != nil {
				c.logger.ErrorContext(ctx, "failed to create debt", err)
				return err
			}

			// Set debt amount from transaction
			err = debt.SetAmountMinor(transaction.Amount, transaction.CurrencyCode)
			if err != nil {
				c.logger.ErrorContext(ctx, "failed to set debt amount", err)
				return err
			}

			err = c.debtsRepo.Save(ctx, debt)
			if err != nil {
				c.logger.ErrorContext(ctx, "failed to save debt", err)
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	response := &models.CreateTransactionResponse{
		Transaction: models.Transaction{
			ID:                   transaction.ID.String(),
			UserID:               transaction.UserID.String(),
			AccountID:            transaction.AccountID.String(),
			Type:                 transaction.Type.String(),
			Status:               transaction.Status.String(),
			Amount:               transaction.AmountMajor(),
			CurrencyCode:         transaction.CurrencyCode.String(),
			OriginalAmount:       pointer.Float64(transaction.OriginalAmountMajor()),
			OriginalCurrencyCode: pointer.String(transaction.OriginalCurrencyCode.String()),
			FxRate:               pointer.Float64(transaction.FxRate),
			Note:                 transaction.RowText,
			PerformedAt:          pointer.TimeOrNil(transaction.PerformedAt),
			RejectedAt:           pointer.TimeOrNil(transaction.RejectedAt),
			CreatedAt:            transaction.CreatedAt,
		},
	}

	if transaction.Category != nil {
		response.CategoryID = pointer.IntOrNil(transaction.Category.ID.Int())
	}

	if transaction.Subcategory != nil {
		response.SubcategoryID = pointer.IntOrNil(transaction.Subcategory.ID)
	}

	if debt != nil {
		response.Debt = models.Debt{
			ID:            debt.ID.String(),
			UserID:        debt.UserID.String(),
			TransactionID: debt.TransactionID.String(),
			Status:        debt.Status.String(),
			Type:          debt.Type.String(),
			Amount:        debt.AmountMajor(),
			CurrencyCode:  debt.CurrencyCode.String(),
			RemindAt:      pointer.TimeOrNil(debt.RemindAt),
			PaidAt:        pointer.TimeOrNil(debt.PaidAt),
			UpdatedAt:     pointer.TimeOrNil(debt.UpdatedAt),
			CreatedAt:     debt.CreatedAt,
		}
	}

	return response, nil
}
