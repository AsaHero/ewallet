package handlers

import (
	"net/http"

	"github.com/AsaHero/e-wallet/internal/delivery/api/apierr"
	"github.com/AsaHero/e-wallet/internal/delivery/api/middleware"
	"github.com/AsaHero/e-wallet/internal/delivery/api/models"
	"github.com/AsaHero/e-wallet/internal/usecase/parser"
	"github.com/gin-gonic/gin"
)

// ParseText godoc
// @Summary Parse text
// @Description Parse text
// @Tags Parse
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body models.ParseTransactionRequest true "Parse transaction request"
// @Success      200 {object} parser.ParseTextView
// @Failure      400 {object} apierr.Response
// @Failure      401 {object} apierr.Response
// @Router       /parse/text [post]
func (h *Handlers) ParseText(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	var req models.ParseTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "invalid request payload", err.Error())
		return
	}

	var response *parser.ParseTextView
	response, err := h.ParserUsecase.Command.ParseText(c, req.Content)
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handlers) ParseVoice(c *gin.Context) {
	c.Status(http.StatusOK)
}

func (h *Handlers) ParseImage(c *gin.Context) {
	c.Status(http.StatusOK)
}
