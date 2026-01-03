package handlers

import (
	"net/http"

	"github.com/AsaHero/e-wallet/internal/delivery/api/apierr"
	"github.com/AsaHero/e-wallet/internal/delivery/api/middleware"
	accountsquery "github.com/AsaHero/e-wallet/internal/usecase/accounts/query"
	"github.com/AsaHero/e-wallet/internal/usecase/transactions/query"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/form/v4"
)

// GetTimeseriesStats godoc
// @Summary      Returns time-series aggregated statistics
// @Tags         Stats
// @Produce      json
// @Security     BearerAuth
// @Param        from query string true "From Date (YYYY-MM-DD)"
// @Param        to query string true "To Date (YYYY-MM-DD)"
// @Param        account_ids query []string false "Account IDs"
// @Param        category_ids query []int false "Category IDs"
// @Param        subcategory_ids query []int false "Subcategory IDs"
// @Param        type query string false "Transaction Type (deposit/withdrawal/transfer/adjustment)"
// @Param        group_by query string false "Group by (day/week/month)" default(day)
// @Success      200 {object} query.TimeseriesStatsView
// @Failure      400 {object} apierr.Response
// @Failure      401 {object} apierr.Response
// @Router       /stats/timeseries [get]
func (h *Handlers) GetTimeseriesStats(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	var query query.GetTimeseriesStatsQuery
	if err := form.NewDecoder().Decode(&query, c.Request.URL.Query()); err != nil {
		apierr.BadRequest(c, "invalid query parameters", err)
		return
	}

	if err := h.Validator.Validate(&query); err != nil {
		apierr.BadRequest(c, "invalid query parameters", err)
		return
	}

	response, err := h.TransactionsUsecase.Query.GetTimeseriesStats(ctx, userID, &query)
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetStatsByCategory godoc
// @Summary      Returns statistics grouped by category
// @Tags         Stats
// @Produce      json
// @Security     BearerAuth
// @Param        from query string true "From Date (YYYY-MM-DD)"
// @Param        to query string true "To Date (YYYY-MM-DD)"
// @Param        account_ids query []string false "Account IDs"
// @Param        type query string false "Transaction Type (deposit/withdrawal)"
// @Success      200 {object} query.CategoryStatsView
// @Failure      400 {object} apierr.Response
// @Failure      401 {object} apierr.Response
// @Router       /stats/by-category [get]
func (h *Handlers) GetStatsByCategory(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	var query query.GetStatsByCategoryQuery
	if err := form.NewDecoder().Decode(&query, c.Request.URL.Query()); err != nil {
		apierr.BadRequest(c, "invalid query parameters", err)
		return
	}

	if err := h.Validator.Validate(&query); err != nil {
		apierr.BadRequest(c, "invalid query parameters", err)
		return
	}

	response, err := h.TransactionsUsecase.Query.GetStatsByCategory(ctx, userID, &query)
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetStatsBySubcategory godoc
// @Summary      Returns statistics grouped by subcategory
// @Tags         Stats
// @Produce      json
// @Security     BearerAuth
// @Param        from query string true "From Date (YYYY-MM-DD)"
// @Param        to query string true "To Date (YYYY-MM-DD)"
// @Param        account_ids query []string false "Account IDs"
// @Param        category_ids query []int false "Category IDs"
// @Param        type query string false "Transaction Type (deposit/withdrawal)"
// @Success      200 {object} query.SubcategoryStatsView
// @Failure      400 {object} apierr.Response
// @Failure      401 {object} apierr.Response
// @Router       /stats/by-subcategory [get]
func (h *Handlers) GetStatsBySubcategory(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	var query query.GetStatsBySubcategoryQuery
	if err := form.NewDecoder().Decode(&query, c.Request.URL.Query()); err != nil {
		apierr.BadRequest(c, "invalid query parameters", err)
		return
	}

	if err := h.Validator.Validate(&query); err != nil {
		apierr.BadRequest(c, "invalid query parameters", err)
		return
	}

	response, err := h.TransactionsUsecase.Query.GetStatsBySubcategory(ctx, userID, &query)
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetStatsByAccount godoc
// @Summary      Returns statistics grouped by account
// @Tags         Stats
// @Produce      json
// @Security     BearerAuth
// @Param        from query string true "From Date (YYYY-MM-DD)"
// @Param        to query string true "To Date (YYYY-MM-DD)"
// @Param        type query string false "Transaction Type (deposit/withdrawal)"
// @Success      200 {object} query.AccountStatsView
// @Failure      400 {object} apierr.Response
// @Failure      401 {object} apierr.Response
// @Router       /stats/by-account [get]
func (h *Handlers) GetStatsByAccount(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	from := c.Query("from")
	to := c.Query("to")
	trnType := c.Query("type")

	response, err := h.TransactionsUsecase.Query.GetStatsByAccount(ctx, userID, from, to, trnType)
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetStatsCompare godoc
// @Summary      Returns period-over-period comparison statistics
// @Tags         Stats
// @Produce      json
// @Security     BearerAuth
// @Param        period query string false "Preset period: this_month_vs_last_month, last_7_days_vs_previous_7_days, this_year_vs_last_year"
// @Param        base_from query string false "Base period start date (YYYY-MM-DD)"
// @Param        base_to query string false "Base period end date (YYYY-MM-DD)"
// @Param        compare_from query string false "Compare period start date (YYYY-MM-DD)"
// @Param        compare_to query string false "Compare period end date (YYYY-MM-DD)"
// @Param        account_ids query []string false "Account IDs"
// @Param        type query string false "Transaction Type (deposit/withdrawal)"
// @Param        top_limit query int false "Number of top changes to return" default(5)
// @Success      200 {object} query.StatsCompareView
// @Failure      400 {object} apierr.Response
// @Failure      401 {object} apierr.Response
// @Router       /stats/compare [get]
func (h *Handlers) GetStatsCompare(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	var query query.GetStatsCompareQuery
	if err := form.NewDecoder().Decode(&query, c.Request.URL.Query()); err != nil {
		apierr.BadRequest(c, "invalid query parameters", err)
		return
	}

	if err := h.Validator.Validate(&query); err != nil {
		apierr.BadRequest(c, "invalid query parameters", err)
		return
	}

	response, err := h.TransactionsUsecase.Query.GetStatsCompare(ctx, userID, &query)
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetBalanceTimeseries godoc
// @Summary      Returns balance over time for accounts
// @Tags         Stats
// @Produce      json
// @Security     BearerAuth
// @Param        from query string true "From Date (YYYY-MM-DD)"
// @Param        to query string true "To Date (YYYY-MM-DD)"
// @Param        group_by query string false "Group by (day/week/month)" default(day)
// @Param        account_ids query []string false "Account IDs"
// @Param        mode query string false "Mode (aggregate/per_account)" default(aggregate)
// @Success      200 {object} accountsquery.BalanceTimeseriesView
// @Failure      400 {object} apierr.Response
// @Failure      401 {object} apierr.Response
// @Router       /stats/timeseries/balance [get]
func (h *Handlers) GetBalanceTimeseries(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	var query accountsquery.GetBalanceTimeseriesQuery
	if err := form.NewDecoder().Decode(&query, c.Request.URL.Query()); err != nil {
		apierr.BadRequest(c, "invalid query parameters", err)
		return
	}

	if err := h.Validator.Validate(&query); err != nil {
		apierr.BadRequest(c, "invalid query parameters", err)
		return
	}

	var response *accountsquery.BalanceTimeseriesView
	response, err := h.AccountsUsecase.Query.GetBalanceTimeseries(ctx, userID, &query)
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}
