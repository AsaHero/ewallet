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
// @Param        request body models.ParseTextRequest true "Parse transaction request"
// @Success      200 {object} parser.ParseTextView
// @Failure      400 {object} apierr.Response
// @Failure      401 {object} apierr.Response
// @Router       /parse/text [post]
func (h *Handlers) ParseText(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	var req models.ParseTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "invalid request payload", err.Error())
		return
	}

	var response *parser.ParseTextView
	response, err := h.ParserUsecase.Command.ParseText(ctx, userID, req.Content)
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ParseVoice godoc
// @Summary Parse voice
// @Description Parse voice
// @Tags Parse
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body models.ParseAudioRequest true "Parse transaction request"
// @Success      200 {object} parser.ParseAudioView
// @Failure      400 {object} apierr.Response
// @Failure      401 {object} apierr.Response
// @Router       /parse/voice [post]
func (h *Handlers) ParseVoice(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	var req models.ParseAudioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "invalid request payload", err.Error())
		return
	}

	var response *parser.ParseAudioView
	response, err := h.ParserUsecase.Command.ParseAudio(ctx, userID, req.FileURL)
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ParseImage godoc
// @Summary Parse image
// @Description Parse image using OCR
// @Tags Parse
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body models.ParseImageRequest true "Parse transaction request"
// @Success      200 {object} parser.ParseImageView
// @Failure      400 {object} apierr.Response
// @Failure      401 {object} apierr.Response
// @Router       /parse/image [post]
func (h *Handlers) ParseImage(c *gin.Context) {
	ctx := c.Request.Context()

	userID := middleware.GetUserID(c)
	if userID == "" {
		apierr.Unauthorized(c, "user context is missing")
		return
	}

	var req models.ParseImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "invalid request payload", err.Error())
		return
	}

	var response *parser.ParseImageView
	response, err := h.ParserUsecase.Command.ParseImage(ctx, userID, req.ImageURL)
	if err != nil {
		apierr.Handle(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}
