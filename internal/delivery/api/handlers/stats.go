package handlers

import (
	"net/http"

	"github.com/AsaHero/e-wallet/internal/delivery/api/apierr"
	"github.com/AsaHero/e-wallet/internal/delivery/api/middleware"
	"github.com/AsaHero/e-wallet/internal/usecase/transactions/query"
	"github.com/gin-gonic/gin"
)

// GetStats godoc
// @Summary      Returns aggregated statistics
// @Tags         Stats
// @Produce      json
// @Security     BearerAuth
// @Param        from query string false "From Date"
// @Param        to query string false "To Date"
// @Param        account_id query string false "Account ID"
// @Success      200 {object} query.GetStatsView
// @Failure      401 {object} apierr.Response
// @Router       /stats/summary [get]
func (h *Handlers) GetStats(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	from := c.Query("from")
	to := c.Query("to")
	accountID := c.Query("account_id")

	var response *query.GetStatsView
	response, err := h.TransactionsUsecase.Query.GetStats(ctx, userID, accountID, from, to)
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}
