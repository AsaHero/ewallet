package handlers

import (
	"net/http"

	"github.com/AsaHero/e-wallet/internal/delivery/api/apierr"
	"github.com/AsaHero/e-wallet/internal/delivery/api/middleware"
	"github.com/AsaHero/e-wallet/internal/delivery/api/models"
	"github.com/AsaHero/e-wallet/internal/usecase/transactions/command"
	"github.com/AsaHero/e-wallet/internal/usecase/transactions/query"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/form/v4"
	"github.com/shogo82148/pointer"
)

// CreateTransaction godoc
// @Summary      Creates a new transaction
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.CreateTransactionRequest true "request"
// @Success      201 {object} models.Transaction
// @Failure      400 {object} apierr.Response
// @Failure      401 {object} apierr.Response
// @Router       /transactions [post]
func (h *Handlers) CreateTransaction(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	var req models.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "invalid request payload", err.Error())
		return
	}

	trn, err := h.TransactionsUsecase.Command.CreateTransaction(ctx, &command.CreateTransactionCommand{
		UserID:               userID,
		AccountID:            req.AccountID,
		CategoryID:           req.CategoryID,
		Type:                 req.Type,
		Amount:               req.Amount,
		CurrencyCode:         req.CurrencyCode,
		OriginalAmount:       req.OriginalAmount,
		OriginalCurrencyCode: req.OriginalCurrencyCode,
		FxRate:               req.FxRate,
		Note:                 req.Note,
		PerformedAt:          req.PerformedAt,
	})
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	transaction := models.Transaction{
		ID:                   trn.ID.String(),
		UserID:               trn.UserID.String(),
		AccountID:            trn.AccountID.String(),
		Type:                 trn.Type.String(),
		Status:               trn.Status.String(),
		Amount:               trn.AmountMajor(),
		CurrencyCode:         trn.CurrencyCode.String(),
		OriginalAmount:       pointer.Float64(trn.OriginalAmountMajor()),
		OriginalCurrencyCode: pointer.String(trn.OriginalCurrencyCode.String()),
		FxRate:               pointer.Float64(trn.FxRate),
		Note:                 trn.RowText,
		PerformedAt:          pointer.TimeOrNil(trn.PerformedAt),
		RejectedAt:           pointer.TimeOrNil(trn.RejectedAt),
		CreatedAt:            trn.CreatedAt,
	}

	if trn.Category != nil {
		transaction.CategoryID = pointer.IntOrNil(trn.Category.ID.Int())
	}

	if trn.Subcategory != nil {
		transaction.SubcategoryID = pointer.IntOrNil(trn.Subcategory.ID)
	}

	c.JSON(http.StatusCreated, transaction)
}

// GetTransactions godoc
// @Summary      Lists transactions with pagination
// @Tags         Transactions
// @Produce      json
// @Security     BearerAuth
// @Param        limit        query    int false "limit"
// @Param        offset       query    int false "offset"
// @Param        from         query    string false "from date (ISO 8601)"
// @Param        to           query    string false "to date (ISO 8601)"
// @Param        type         query    string false "transaction type"
// @Param        category_ids query    []int false "category ids"
// @Param        account_ids  query    []string false "account ids"
// @Param        min_amount   query    int false "min amount"
// @Param        max_amount   query    int false "max amount"
// @Param        search       query    string false "search term"
// @Success      200 {object} models.TransactionsResponse
// @Failure      400 {object} apierr.Response
// @Failure      401 {object} apierr.Response
// @Router       /transactions [get]
func (h *Handlers) GetTransactions(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	var req query.GetByFilterQuery
	if err := form.NewDecoder().Decode(&req, c.Request.URL.Query()); err != nil {
		apierr.BadRequest(c, "invalid request form", err.Error())
		return
	}

	if err := h.Validator.Validate(&req); err != nil {
		apierr.BadRequest(c, "invalid request form", err.Error())
		return
	}

	var response *models.TransactionsResponse
	response, err := h.TransactionsUsecase.Query.GetByFilter(ctx, userID, &req)
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetTransaction godoc
// @Summary      Returns transaction by ID
// @Tags         Transactions
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "transaction id"
// @Success      200 {object} models.Transaction
// @Failure      401 {object} apierr.Response
// @Router       /transactions/{id} [get]
func (h *Handlers) GetTransaction(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	trnID := c.Param("id")
	if trnID == "" {
		apierr.BadRequest(c, "transaction id is missing")
		return
	}

	trn, err := h.TransactionsUsecase.Query.GetByID(ctx, trnID)
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	transaction := models.Transaction{
		ID:                   trn.ID.String(),
		UserID:               trn.UserID.String(),
		AccountID:            trn.AccountID.String(),
		CategoryID:           pointer.IntOrNil(trn.Category.ID.Int()),
		SubcategoryID:        pointer.IntOrNil(trn.Subcategory.ID),
		Type:                 trn.Type.String(),
		Status:               trn.Status.String(),
		Amount:               trn.AmountMajor(),
		CurrencyCode:         trn.CurrencyCode.String(),
		OriginalAmount:       pointer.Float64(trn.OriginalAmountMajor()),
		OriginalCurrencyCode: pointer.String(trn.OriginalCurrencyCode.String()),
		FxRate:               pointer.Float64(trn.FxRate),
		Note:                 trn.RowText,
		PerformedAt:          pointer.TimeOrNil(trn.PerformedAt),
		RejectedAt:           pointer.TimeOrNil(trn.RejectedAt),
		CreatedAt:            trn.CreatedAt,
	}

	if trn.Category != nil {
		transaction.CategoryID = pointer.IntOrNil(trn.Category.ID.Int())
	}

	if trn.Subcategory != nil {
		transaction.SubcategoryID = pointer.IntOrNil(trn.Subcategory.ID)
	}

	c.JSON(http.StatusOK, transaction)
}

// DeleteTransaction godoc
// @Summary      Deletes a transaction
// @Tags         Transactions
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "transaction id"
// @Success      204
// @Failure      400 {object} apierr.Response
// @Failure      401 {object} apierr.Response
// @Router       /transactions/{id} [delete]
func (h *Handlers) DeleteTransaction(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	trnID := c.Param("id")
	if trnID == "" {
		apierr.BadRequest(c, "transaction id is missing")
		return
	}

	err := h.TransactionsUsecase.Command.DeleteTransaction(ctx, &command.DeleteTransactionCommand{
		UserID:        userID,
		TransactionID: trnID,
	})
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdateTransaction godoc
// @Summary      Updates a transaction
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "transaction id"
// @Param        request body models.UpdateTransactionRequest true "request"
// @Success      200 {object} models.Transaction
// @Failure      400 {object} apierr.Response
// @Failure      401 {object} apierr.Response
// @Router       /transactions/{id} [put]
func (h *Handlers) UpdateTransaction(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	trnID := c.Param("id")
	if trnID == "" {
		apierr.BadRequest(c, "transaction id is missing")
		return
	}

	var req models.UpdateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "invalid request payload", err.Error())
		return
	}

	trn, err := h.TransactionsUsecase.Command.UpdateTransaction(ctx, &command.UpdateTransactionCommand{
		UserID:               userID,
		TransactionID:        trnID,
		CategoryID:           req.CategoryID,
		SubcategoryID:        req.SubcategoryID,
		Type:                 req.Type,
		Amount:               req.Amount,
		CurrencyCode:         req.CurrencyCode,
		OriginalAmount:       req.OriginalAmount,
		OriginalCurrencyCode: req.OriginalCurrencyCode,
		FxRate:               req.FxRate,
		Note:                 req.Note,
		PerformedAt:          req.PerformedAt,
	})
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	transaction := models.Transaction{
		ID:                   trn.ID.String(),
		UserID:               trn.UserID.String(),
		AccountID:            trn.AccountID.String(),
		Type:                 trn.Type.String(),
		Status:               trn.Status.String(),
		Amount:               trn.AmountMajor(),
		CurrencyCode:         trn.CurrencyCode.String(),
		OriginalAmount:       pointer.Float64(trn.OriginalAmountMajor()),
		OriginalCurrencyCode: pointer.String(trn.OriginalCurrencyCode.String()),
		FxRate:               pointer.Float64(trn.FxRate),
		Note:                 trn.RowText,
		PerformedAt:          pointer.TimeOrNil(trn.PerformedAt),
		RejectedAt:           pointer.TimeOrNil(trn.RejectedAt),
		CreatedAt:            trn.CreatedAt,
	}

	if trn.Category != nil {
		transaction.CategoryID = pointer.IntOrNil(trn.Category.ID.Int())
	}

	if trn.Subcategory != nil {
		transaction.SubcategoryID = pointer.IntOrNil(trn.Subcategory.ID)
	}

	c.JSON(http.StatusOK, transaction)
}
