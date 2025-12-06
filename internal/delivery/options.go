package delivery

import (
	"github.com/AsaHero/e-wallet/internal/delivery/api/validation"
	"github.com/AsaHero/e-wallet/internal/usecase/accounts"
	"github.com/AsaHero/e-wallet/internal/usecase/categories"
	"github.com/AsaHero/e-wallet/internal/usecase/notifications"
	"github.com/AsaHero/e-wallet/internal/usecase/parser"
	"github.com/AsaHero/e-wallet/internal/usecase/transactions"
	"github.com/AsaHero/e-wallet/internal/usecase/users"
	"github.com/AsaHero/e-wallet/pkg/config"
	"github.com/AsaHero/e-wallet/pkg/logger"
)

type Options struct {
	Config              *config.Config
	Validator           *validation.Validator
	Logger              *logger.Logger
	UsersUsecase        *users.Module
	AccountsUsecase     *accounts.Module
	TransactionsUsecase *transactions.Module
	CategoriesUsecase   *categories.Module
	ParserUsecase       *parser.Module
	NotificationUsecase *notifications.Module
}
