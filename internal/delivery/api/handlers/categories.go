package handlers

import (
	"net/http"

	"github.com/AsaHero/e-wallet/internal/delivery/api/apierr"
	"github.com/AsaHero/e-wallet/internal/delivery/api/middleware"
	"github.com/AsaHero/e-wallet/internal/usecase/categories/query"
	"github.com/gin-gonic/gin"
)

// GetCategories godoc
// @Summary      Lists available categories
// @Tags         Categories
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} query.Category
// @Failure      401 {object} apierr.Response
// @Router       /categories [get]
func (h *Handlers) GetCategories(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	var categories []query.Category
	categories, err := h.CategoriesUsecase.Query.GetAllCategories(ctx, userID)
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	c.JSON(http.StatusOK, categories)
}

// GetSubcategories godoc
// @Summary      Lists available subcategories
// @Tags         Categories
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} query.Subcategory
// @Failure      401 {object} apierr.Response
// @Router       /subcategories [get]
func (h *Handlers) GetSubcategories(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	var subcategories []query.Subcategory
	subcategories, err := h.CategoriesUsecase.Query.GetAllSubcategories(ctx, userID)
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	c.JSON(http.StatusOK, subcategories)
}
