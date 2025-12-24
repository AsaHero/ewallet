package handlers

import (
	"fmt"
	"net/http"

	"github.com/AsaHero/e-wallet/internal/delivery/api/apierr"
	"github.com/AsaHero/e-wallet/internal/delivery/api/middleware"
	"github.com/AsaHero/e-wallet/internal/delivery/api/models"
	"github.com/AsaHero/e-wallet/internal/usecase/categories/command"
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

// CreateCategory godoc
// @Summary      Create a new category
// @Tags         Categories
// @Produce      json
// @Security     BearerAuth
// @Param 		 request body models.CreateCategoryRequest true "request"
// @Success      201 {object} models.Category
// @Failure      401 {object} apierr.Response
// @Router       /categories [post]
func (h *Handlers) CreateCategory(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	var req models.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "invalid request payload", err.Error())
		return
	}

	var response *models.Category
	response, err := h.CategoriesUsecase.Command.CreateCategory(ctx, &command.CreateCategoryCommand{
		UserID: userID,
		Name:   req.Name,
		Emoji:  req.Emoji,
	})
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// CreateSubcategory godoc
// @Summary      Create a new subcategory
// @Tags         Categories
// @Produce      json
// @Security     BearerAuth
// @Param 		 request body models.CreateSubcategoryRequest true "request"
// @Success      201 {object} models.Subcategory
// @Failure      401 {object} apierr.Response
// @Router       /subcategories [post]
func (h *Handlers) CreateSubcategory(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	var req models.CreateSubcategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "invalid request payload", err.Error())
		return
	}

	var response *models.Subcategory
	response, err := h.CategoriesUsecase.Command.CreateSubcategory(ctx, &command.CreateSubcategoryCommand{
		UserID:     userID,
		CategoryID: req.CategoryID,
		Name:       req.Name,
		Emoji:      req.Emoji,
	})
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// DeleteCategory godoc
// @Summary      Delete a category
// @Tags         Categories
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Category ID"
// @Success      204
// @Failure      401 {object} apierr.Response
// @Failure      404 {object} apierr.Response
// @Router       /categories/{id} [delete]
func (h *Handlers) DeleteCategory(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	categoryID := c.Param("id")
	if categoryID == "" {
		apierr.BadRequest(c, "category id is required", "")
		return
	}

	var categoryIDInt int
	if _, err := fmt.Sscanf(categoryID, "%d", &categoryIDInt); err != nil {
		apierr.BadRequest(c, "invalid category id", err.Error())
		return
	}

	err := h.CategoriesUsecase.Command.DeleteCategory(ctx, &command.DeleteCategoryCommand{
		UserID:     userID,
		CategoryID: categoryIDInt,
	})
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// DeleteSubcategory godoc
// @Summary      Delete a subcategory
// @Tags         Categories
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Subcategory ID"
// @Success      204
// @Failure      401 {object} apierr.Response
// @Failure      404 {object} apierr.Response
// @Router       /subcategories/{id} [delete]
func (h *Handlers) DeleteSubcategory(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	subcategoryID := c.Param("id")
	if subcategoryID == "" {
		apierr.BadRequest(c, "subcategory id is required", "")
		return
	}

	var subcategoryIDInt int
	if _, err := fmt.Sscanf(subcategoryID, "%d", &subcategoryIDInt); err != nil {
		apierr.BadRequest(c, "invalid subcategory id", err.Error())
		return
	}

	err := h.CategoriesUsecase.Command.DeleteSubcategory(ctx, &command.DeleteSubcategoryCommand{
		UserID:        userID,
		SubcategoryID: subcategoryIDInt,
	})
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
