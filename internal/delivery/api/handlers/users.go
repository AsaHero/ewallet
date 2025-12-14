package handlers

import (
	"net/http"

	"github.com/AsaHero/e-wallet/internal/delivery/api/apierr"
	"github.com/AsaHero/e-wallet/internal/delivery/api/middleware"
	"github.com/AsaHero/e-wallet/internal/delivery/api/models"
	"github.com/AsaHero/e-wallet/internal/usecase/users/command"
	"github.com/gin-gonic/gin"
	"github.com/shogo82148/pointer"
)

// GetMe godoc
// @Summary      Returns profile for the authenticated user
// @Description	 Provides profile of the user extracted from JWT claims
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.User
// @Failure      401 {object} apierr.Response
// @Router       /users/me [get]
func (h *Handlers) GetMe(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	user, err := h.UsersUsecase.Query.GetByID(ctx, userID)
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	response := models.User{
		ID:           user.ID.String(),
		TgUserID:     user.TGUserID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Username:     user.Username,
		LanguageCode: user.LanguageCode.String(),
		CurrencyCode: user.CurrencyCode.String(),
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    pointer.TimeOrNil(user.UpdatedAt),
	}

	c.JSON(http.StatusOK, response)
}

// UpdateMe godoc
// @Summary      Updates profile for the authenticated user
// @Description	 Updates language/currency preferences for the user
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.UpdateUserRequest true "request"
// @Success      200 {object} models.User
// @Failure      400 {object} apierr.Response
// @Failure      401 {object} apierr.Response
// @Router       /users/me [patch]
func (h *Handlers) UpdateMe(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "invalid request payload", err.Error())
		return
	}

	user, err := h.UsersUsecase.Command.Update(ctx, &command.UpdateCommand{
		UserID:       userID,
		LanguageCode: req.LanguageCode,
		CurrencyCode: req.CurrencyCode,
		Timezone:     req.Timezone,
	})
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	response := models.User{
		ID:           user.ID.String(),
		TgUserID:     user.TGUserID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Username:     user.Username,
		LanguageCode: user.LanguageCode.String(),
		CurrencyCode: user.CurrencyCode.String(),
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    pointer.TimeOrNil(user.UpdatedAt),
	}

	c.JSON(http.StatusOK, response)
}
