package handlers

import (
	"net/http"

	"github.com/AsaHero/e-wallet/internal/delivery/api/apierr"
	"github.com/AsaHero/e-wallet/internal/delivery/api/models"
	"github.com/AsaHero/e-wallet/internal/usecase/users/command"
	"github.com/AsaHero/e-wallet/pkg/security"
	"github.com/gin-gonic/gin"
	"github.com/shogo82148/pointer"
)

// AuthTelegram godoc
// @Summary      Authenticates user through telegram
// @Description	 Creates a new user session using Telegram payload and returns JWT token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body models.AuthRequest true "request"
// @Success      200 {object} models.AuthResponse
// @Failure      400 {object} apierr.Response
// @Failure      500 {object} apierr.Response
// @Router       /auth/telegram [post]
func (h *Handlers) AuthTelegram(c *gin.Context) {
	var req models.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "invalid request payload", err.Error())
		return
	}

	req.CurrencyCode = ""

	user, err := h.UsersUsecase.Command.AuthTelegram(c, &command.AuthTelegramCommand{
		TelegramUserID: req.TgUserID,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Username:       req.Username,
		CurrencyCode:   req.CurrencyCode,
	})
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	token, err := security.GenerateToken(user.ID.String(), req.TgUserID)
	if err != nil {
		apierr.InternalError(c, "failed to issue auth token")
		return
	}

	resp := models.AuthResponse{
		Token: token,
		User: models.User{
			ID:           user.ID.String(),
			TgUserID:     user.TGUserID,
			FirstName:    user.FirstName,
			LastName:     user.LastName,
			Username:     user.Username,
			LanguageCode: user.LanguageCode.String(),
			CurrencyCode: user.CurrencyCode.String(),
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    pointer.TimeOrNil(user.UpdatedAt),
		},
	}

	c.JSON(http.StatusOK, resp)
}
