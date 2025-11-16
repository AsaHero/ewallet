package handlers

import (
	"net/http"

	"github.com/AsaHero/e-wallet/internal/delivery/api/apierr"
	"github.com/AsaHero/e-wallet/internal/delivery/api/middleware"
	"github.com/AsaHero/e-wallet/internal/delivery/api/models"
	"github.com/gin-gonic/gin"
)

// GetCategories godoc
// @Summary      Lists available categories
// @Tags         Categories
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} models.Category
// @Failure      401 {object} apierr.Response
// @Router       /categories [get]
func (h *Handlers) GetCategories(c *gin.Context) {
	if middleware.GetUserID(c) == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	list, err := h.CategoriesUsecase.Query.GetAll(c)
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	var categories []models.Category
	for _, cat := range list {
		categories = append(categories, models.Category{
			ID:       cat.ID,
			Slug:     cat.Slug,
			Position: cat.Position,
			Name:     cat.Name,
		})
	}

	c.JSON(http.StatusOK, categories)
}
