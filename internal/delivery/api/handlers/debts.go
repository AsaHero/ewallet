package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/AsaHero/e-wallet/internal/delivery/api/apierr"
	"github.com/AsaHero/e-wallet/internal/delivery/api/middleware"
	"github.com/AsaHero/e-wallet/internal/delivery/api/models"
	"github.com/AsaHero/e-wallet/internal/usecase/debts/command"
	"github.com/AsaHero/e-wallet/internal/usecase/debts/query"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/form/v4"
	"github.com/shogo82148/pointer"
)

// CreateDebt godoc
// @Summary      Creates a debt
// @Tags         Debts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.CreateDebtRequest true "request"
// @Success      200 {object} models.Debt
// @Failure      400 {object} apierr.Response
// @Failure      401 {object} apierr.Response
// @Router       /debts [post]
func (h *Handlers) CreateDebt(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	var req models.CreateDebtRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		apierr.BadRequest(c, "invalid request payload", err.Error())
		return
	}

	debt, err := h.DebtsUsecase.Command.CreateDebt(ctx, &command.CreateDebtCommand{
		TransactionID: req.TransactionID,
		RemindAt:      req.RemindAt,
	})
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	response := models.Debt{
		ID:            debt.ID.String(),
		UserID:        debt.UserID.String(),
		TransactionID: debt.TransactionID.String(),
		Type:          debt.Type.String(),
		Status:        debt.Status.String(),
		Amount:        debt.AmountMajor(),
		CurrencyCode:  debt.CurrencyCode.String(),
		RemindAt:      pointer.TimeOrNil(debt.RemindAt),
		PaidAt:        pointer.TimeOrNil(debt.PaidAt),
		CreatedAt:     debt.CreatedAt,
		UpdatedAt:     pointer.TimeOrNil(debt.UpdatedAt),
	}

	c.JSON(http.StatusOK, response)
}

// UpdateDebt godoc
// @Summary      Updates a debt
// @Tags         Debts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "debt id"
// @Param        request body models.UpdateDebtRequest true "request"
// @Success      200 {object} models.Debt
// @Failure      400 {object} apierr.Response
// @Failure      401 {object} apierr.Response
// @Router       /debts/{id} [put]
func (h *Handlers) UpdateDebt(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	debtID := c.Param("id")
	if debtID == "" {
		apierr.BadRequest(c, "debt id is missing")
		return
	}

	var req models.UpdateDebtRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "invalid request payload", err.Error())
		return
	}

	debt, err := h.DebtsUsecase.Command.UpdateDebt(ctx, &command.UpdateDebtCommand{
		UserID:       userID,
		DebtID:       debtID,
		Amount:       req.Amount,
		CurrencyCode: req.CurrencyCode,
		RemindAt:     req.RemindAt,
	})
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	response := models.Debt{
		ID:            debt.ID.String(),
		UserID:        debt.UserID.String(),
		TransactionID: debt.TransactionID.String(),
		Type:          debt.Type.String(),
		Status:        debt.Status.String(),
		Amount:        debt.AmountMajor(),
		CurrencyCode:  debt.CurrencyCode.String(),
		RemindAt:      pointer.TimeOrNil(debt.RemindAt),
		PaidAt:        pointer.TimeOrNil(debt.PaidAt),
		CreatedAt:     debt.CreatedAt,
		UpdatedAt:     pointer.TimeOrNil(debt.UpdatedAt),
	}

	c.JSON(http.StatusOK, response)
}

// PayDebt godoc
// @Summary      Marks a debt as paid
// @Tags         Debts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "debt id"
// @Param        request body models.PayDebtRequest true "request"
// @Success      200 {object} models.Debt
// @Failure      400 {object} apierr.Response
// @Failure      401 {object} apierr.Response
// @Router       /debts/{id}/pay [post]
func (h *Handlers) PayDebt(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	debtID := c.Param("id")
	if debtID == "" {
		apierr.BadRequest(c, "debt id is missing")
		return
	}

	var req models.PayDebtRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "invalid request payload", err.Error())
		return
	}

	debt, err := h.DebtsUsecase.Command.PayDebt(ctx, &command.PayDebtCommand{
		UserID: userID,
		DebtID: debtID,
		PaidAt: req.PaidAt,
	})
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	response := models.Debt{
		ID:            debt.ID.String(),
		UserID:        debt.UserID.String(),
		TransactionID: debt.TransactionID.String(),
		Type:          debt.Type.String(),
		Status:        debt.Status.String(),
		Amount:        debt.AmountMajor(),
		CurrencyCode:  debt.CurrencyCode.String(),
		RemindAt:      pointer.TimeOrNil(debt.RemindAt),
		PaidAt:        pointer.TimeOrNil(debt.PaidAt),
		CreatedAt:     debt.CreatedAt,
		UpdatedAt:     pointer.TimeOrNil(debt.UpdatedAt),
	}

	c.JSON(http.StatusOK, response)
}

// CancelDebt godoc
// @Summary      Cancels a debt
// @Tags         Debts
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "debt id"
// @Success      200 {object} models.Debt
// @Failure      400 {object} apierr.Response
// @Failure      401 {object} apierr.Response
// @Router       /debts/{id}/cancel [post]
func (h *Handlers) CancelDebt(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	debtID := c.Param("id")
	if debtID == "" {
		apierr.BadRequest(c, "debt id is missing")
		return
	}

	debt, err := h.DebtsUsecase.Command.CancelDebt(ctx, &command.CancelDebtCommand{
		UserID: userID,
		DebtID: debtID,
	})
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	response := models.Debt{
		ID:            debt.ID.String(),
		UserID:        debt.UserID.String(),
		TransactionID: debt.TransactionID.String(),
		Type:          debt.Type.String(),
		Status:        debt.Status.String(),
		Amount:        debt.AmountMajor(),
		CurrencyCode:  debt.CurrencyCode.String(),
		RemindAt:      pointer.TimeOrNil(debt.RemindAt),
		PaidAt:        pointer.TimeOrNil(debt.PaidAt),
		CreatedAt:     debt.CreatedAt,
		UpdatedAt:     pointer.TimeOrNil(debt.UpdatedAt),
	}

	c.JSON(http.StatusOK, response)
}

// GetDebt godoc
// @Summary      Returns debt by ID
// @Tags         Debts
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "debt id"
// @Success      200 {object} models.Debt
// @Failure      401 {object} apierr.Response
// @Router       /debts/{id} [get]
func (h *Handlers) GetDebt(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	debtID := c.Param("id")
	if debtID == "" {
		apierr.BadRequest(c, "debt id is missing")
		return
	}

	debt, err := h.DebtsUsecase.Query.GetByID(ctx, debtID)
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	response := models.Debt{
		ID:            debt.ID.String(),
		UserID:        debt.UserID.String(),
		TransactionID: debt.TransactionID.String(),
		Type:          debt.Type.String(),
		Status:        debt.Status.String(),
		Amount:        debt.AmountMajor(),
		CurrencyCode:  debt.CurrencyCode.String(),
		RemindAt:      pointer.TimeOrNil(debt.RemindAt),
		PaidAt:        pointer.TimeOrNil(debt.PaidAt),
		CreatedAt:     debt.CreatedAt,
		UpdatedAt:     pointer.TimeOrNil(debt.UpdatedAt),
	}

	c.JSON(http.StatusOK, response)
}

// GetDebts godoc
// @Summary      Lists debts with pagination
// @Tags         Debts
// @Produce      json
// @Security     BearerAuth
// @Param        limit           query    int      false "limit"
// @Param        offset          query    int      false "offset"
// @Param        transaction_ids query    []string false "transaction ids"
// @Param        types           query    []string false "debt types (borrow, lend)"
// @Param        statuses        query    []string false "debt statuses (open, paid, cancelled)"
// @Success      200 {object} models.DebtsResponse
// @Failure      400 {object} apierr.Response
// @Failure      401 {object} apierr.Response
// @Router       /debts [get]
func (h *Handlers) GetDebts(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	var req query.GetDebtsByFilterQuery
	if err := form.NewDecoder().Decode(&req, c.Request.URL.Query()); err != nil {
		apierr.BadRequest(c, "invalid request form", err.Error())
		return
	}

	if err := h.Validator.Validate(&req); err != nil {
		apierr.BadRequest(c, "invalid request form", err.Error())
		return
	}

	debts, count, err := h.DebtsUsecase.Query.GetByFilter(ctx, userID, &req)
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	items := make([]models.Debt, 0, len(debts))
	for _, debt := range debts {
		items = append(items, models.Debt{
			ID:            debt.ID.String(),
			UserID:        debt.UserID.String(),
			TransactionID: debt.TransactionID.String(),
			Type:          debt.Type.String(),
			Status:        debt.Status.String(),
			Amount:        debt.AmountMajor(),
			CurrencyCode:  debt.CurrencyCode.String(),
			RemindAt:      pointer.TimeOrNil(debt.RemindAt),
			PaidAt:        pointer.TimeOrNil(debt.PaidAt),
			CreatedAt:     debt.CreatedAt,
			UpdatedAt:     pointer.TimeOrNil(debt.UpdatedAt),
		})
	}

	response := models.DebtsResponse{
		Items: items,
		Pagination: models.PaginationResponse{
			Limit:  uint64(req.Limit),
			Offset: uint64(req.Offset),
			Total:  int64(count),
		},
	}

	c.JSON(http.StatusOK, response)
}
