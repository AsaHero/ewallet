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

	var response *query.GetStatsView
	response, err := h.TransactionsUsecase.Query.GetStats(ctx, userID)
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}
