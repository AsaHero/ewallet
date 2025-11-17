package handlers

import (
	"net/http"

	"github.com/AsaHero/e-wallet/internal/delivery/api/apierr"
	"github.com/AsaHero/e-wallet/internal/delivery/api/middleware"
	"github.com/AsaHero/e-wallet/internal/delivery/api/models"
	"github.com/AsaHero/e-wallet/internal/usecase/transactions/command"
	"github.com/AsaHero/e-wallet/internal/usecase/transactions/query"
	"github.com/gin-gonic/gin"
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
		UserID:       userID,
		AccountID:    req.AccountID,
		CategoryID:   req.CategoryID,
		Type:         req.Type,
		Amount:       req.Amount,
		CurrencyCode: req.CurrencyCode,
		Note:         req.Note,
		PerformedAt:  req.PerformedAt,
	})
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	transaction := models.Transaction{
		ID:           trn.ID.String(),
		UserID:       trn.UserID.String(),
		AccountID:    trn.AccountID.String(),
		CategoryID:   &trn.Category.ID,
		Type:         trn.Type.String(),
		Status:       trn.Status.String(),
		Amount:       trn.AmountMajor(),
		CurrencyCode: trn.CurrencyCode.String(),
		Note:         trn.RowText,
		PerformedAt:  pointer.TimeOrNil(trn.PerformedAt),
		RejectedAt:   pointer.TimeOrNil(trn.RejectedAt),
		CreatedAt:    trn.CreatedAt,
	}

	c.JSON(http.StatusCreated, transaction)
}

// GetTransactions godoc
// @Summary      Lists transactions with pagination
// @Tags         Transactions
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query    int false "limit"
// @Param        offset query    int false "offset"
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

	var page models.PaginationRequest
	if err := c.ShouldBindQuery(&page); err != nil {
		apierr.BadRequest(c, "invalid pagination params", err.Error())
		return
	}

	if page.Limit == 0 {
		page.Limit = 20
	}

	transactions, total, err := h.TransactionsUsecase.Query.GetByFilter(ctx, &query.GetByFilterQuery{
		UserID: userID,
		Limit:  int(page.Limit),
		Offset: int(page.Offset),
	})
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	resp := models.TransactionsResponse{
		Items: make([]models.Transaction, 0, len(transactions)),
		Pagination: models.PaginationResponse{
			Limit:  page.Limit,
			Offset: page.Offset,
			Total:  int64(total),
		},
	}

	for _, trn := range transactions {
		resp.Items = append(resp.Items, models.Transaction{
			ID:           trn.ID.String(),
			UserID:       trn.UserID.String(),
			AccountID:    trn.AccountID.String(),
			CategoryID:   &trn.Category.ID,
			Type:         trn.Type.String(),
			Status:       trn.Status.String(),
			Amount:       trn.AmountMajor(),
			CurrencyCode: trn.CurrencyCode.String(),
			Note:         trn.RowText,
			PerformedAt:  pointer.TimeOrNil(trn.PerformedAt),
			RejectedAt:   pointer.TimeOrNil(trn.RejectedAt),
			CreatedAt:    trn.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, resp)
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
		ID:           trn.ID.String(),
		UserID:       trn.UserID.String(),
		AccountID:    trn.AccountID.String(),
		CategoryID:   &trn.Category.ID,
		Type:         trn.Type.String(),
		Status:       trn.Status.String(),
		Amount:       trn.AmountMajor(),
		CurrencyCode: trn.CurrencyCode.String(),
		Note:         trn.RowText,
		PerformedAt:  pointer.TimeOrNil(trn.PerformedAt),
		RejectedAt:   pointer.TimeOrNil(trn.RejectedAt),
		CreatedAt:    trn.CreatedAt,
	}

	c.JSON(http.StatusOK, transaction)
}
