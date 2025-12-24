package handlers

import (
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
// @Param 		 request body command.CreateCategoryCommand true "request"
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

	var req command.CreateCategoryCommand
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
// @Param 		 request body command.CreateSubcategoryCommand true "request"
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

	var req command.CreateSubcategoryCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "invalid request payload", err.Error())
		return
	}

	var response *models.Subcategory
	response, err := h.CategoriesUsecase.Command.CreateSubcategory(ctx, &command.CreateSubcategoryCommand{
		UserID:     userID,
		Name:       req.Name,
		CategoryID: req.CategoryID,
		Emoji:      req.Emoji,
	})
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}
