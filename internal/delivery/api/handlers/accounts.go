package handlers

import (
	"net/http"
	"time"

	"github.com/AsaHero/e-wallet/internal/delivery/api/apierr"
	"github.com/AsaHero/e-wallet/internal/delivery/api/middleware"
	"github.com/AsaHero/e-wallet/internal/delivery/api/models"
	"github.com/AsaHero/e-wallet/internal/entities"
	"github.com/AsaHero/e-wallet/internal/usecase/accounts/command"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shogo82148/pointer"
)

// GetAccounts godoc
// @Summary      Lists accounts for the authenticated user
// @Tags         Accounts
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} models.Account
// @Failure      401 {object} apierr.Response
// @Router       /accounts [get]
func (h *Handlers) GetAccounts(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	now := time.Now()
	accounts := []models.Account{
		{
			ID:        uuid.NewString(),
			UserID:    userID,
			Name:      "Default wallet",
			Balance:   125000,
			IsDefault: true,
			CreatedAt: now.Add(-48 * time.Hour),
			UpdatedAt: &now,
		},
	}

	c.JSON(http.StatusOK, accounts)
}

// CreateAccount godoc
// @Summary      Creates a new account
// @Tags         Accounts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.CreateAccountRequest true "request"
// @Success      201 {object} models.Account
// @Failure      400 {object} apierr.Response
// @Failure      401 {object} apierr.Response
// @Router       /accounts [post]
func (h *Handlers) CreateAccount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "invalid request payload", err.Error())
		return
	}

	account, err := h.AccountsUsecase.Command.CreateAccount(c, &command.CreateAccountCommand{
		UserID:    userID,
		Name:      req.Name,
		Balance:   req.Balance,
		IsDefault: req.IsDefault,
	})
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	response := models.Account{
		ID:        account.ID.String(),
		UserID:    account.UserID.String(),
		Name:      account.Name,
		Balance:   account.AmountMajor(entities.UZS),
		IsDefault: account.IsDefault,
		CreatedAt: account.CreatedAt,
		UpdatedAt: pointer.TimeOrNil(account.UpdatedAt),
	}

	c.JSON(http.StatusCreated, response)
}

// UpdateAccount godoc
// @Summary      Updates account information
// @Tags         Accounts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "account id"
// @Param        request body models.UpdateAccountRequest true "request"
// @Success      200 {object} models.Account
// @Failure      400 {object} apierr.Response
// @Failure      401 {object} apierr.Response
// @Router       /accounts/{id} [patch]
func (h *Handlers) UpdateAccount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	accountID := c.Param("id")
	if accountID == "" {
		apierr.BadRequest(c, "account id is missing")
		return
	}

	var req models.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "invalid request payload", err.Error())
		return
	}

	account, err := h.AccountsUsecase.Command.UpdateAccount(c, &command.UpdateAccounCommand{
		UserID:    userID,
		AccountID: accountID,
		Name:      req.Name,
		IsDefault: req.IsDefault,
	})
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	response := models.Account{
		ID:        account.ID.String(),
		UserID:    account.UserID.String(),
		Name:      account.Name,
		Balance:   account.AmountMajor(entities.UZS),
		IsDefault: account.IsDefault,
		CreatedAt: account.CreatedAt,
		UpdatedAt: pointer.TimeOrNil(account.UpdatedAt),
	}

	c.JSON(http.StatusOK, response)
}

// DeleteAccount godoc
// @Summary      Deletes an account
// @Tags         Accounts
// @Security     BearerAuth
// @Param        id path string true "account id"
// @Success      204
// @Failure      401 {object} apierr.Response
// @Router       /accounts/{id} [delete]
func (h *Handlers) DeleteAccount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	accountID := c.Param("id")
	if accountID == "" {
		apierr.BadRequest(c, "account id is missing")
		return
	}

	err := h.AccountsUsecase.Command.DeleteAccount(c, &command.DeleteAccountCommand{
		UserID:    userID,
		AccountID: accountID,
	})
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
