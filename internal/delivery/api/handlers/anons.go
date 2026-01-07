package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/AsaHero/e-wallet/internal/delivery/api/apierr"
	"github.com/AsaHero/e-wallet/internal/delivery/api/models"
	"github.com/AsaHero/e-wallet/internal/usecase/anons/command"
	"github.com/AsaHero/e-wallet/internal/usecase/anons/query"
	"github.com/gin-gonic/gin"
)

// CreateAnonBroadcast godoc
// @Summary      Creates an anonymous broadcast
// @Description  Creates a broadcast task to send anonymous messages to Telegram users with optional filters
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BasicAuth
// @Param        request body models.CreateAnonBroadcastRequest true "request"
// @Success      200 {object} models.AnonBroadcastResponse
// @Failure      400 {object} apierr.Response
// @Failure      401 {object} apierr.Response
// @Router       /admin/anons [post]
func (h *Handlers) CreateAnons(c *gin.Context) {
	ctx := c.Request.Context()

	var req models.CreateAnonBroadcastRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		apierr.BadRequest(c, "invalid request payload", err.Error())
		return
	}

	if err := h.Validator.Validate(&req); err != nil {
		apierr.BadRequest(c, "invalid request payload", err.Error())
		return
	}

	cmd := &command.CreateAnonCommand{
		VideoFileID:  req.VideoFileID,
		PhotoFileID:  req.PhotoFileID,
		Message:      req.Message,
		ReplyMarkup:  req.ReplyMarkup,
		LanguageCode: req.Language,
	}

	anon, err := h.AnonsUsecase.Command.CreateAnon(ctx, cmd)
	if err != nil {
		apierr.InternalError(c, "failed to create anon broadcast", err.Error())
		return
	}

	var replyMarkup map[string]any
	if len(anon.ReplyMarkup) > 0 {
		if err := json.Unmarshal(anon.ReplyMarkup, &replyMarkup); err != nil {
			h.Logger.ErrorContext(ctx, "failed to unmarshal reply markup", err)
		}
	}

	videoFileID := ""
	if anon.VideoFileID != "" {
		videoFileID = anon.VideoFileID
	}

	photoFileID := ""
	if anon.PhotoFileID != "" {
		photoFileID = anon.PhotoFileID
	}

	c.JSON(http.StatusOK, models.AnonBroadcastResponse{
		ID:          anon.ID.String(),
		VideoFileID: videoFileID,
		PhotoFileID: photoFileID,
		Message:     anon.Message,
		ReplyMarkup: replyMarkup,
		CreatedAt:   anon.CreatedAt,
	})
}

// BroadcastAnons godoc
// @Summary      Triggers an anonymous broadcast
// @Description  Triggers sending of an existing broadcast to users based on filters
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BasicAuth
// @Param        id path string true "Anon ID"
// @Param        request body models.AnonBroadcastFilters true "filters"
// @Success      200 {object} models.SuccessResponse
// @Failure      400 {object} apierr.Response
// @Failure      404 {object} apierr.Response
// @Failure      500 {object} apierr.Response
// @Router       /admin/anons/{id}/broadcast [post]
func (h *Handlers) BroadcastAnons(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req models.AnonBroadcastFilters
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		apierr.BadRequest(c, "invalid request payload", err.Error())
		return
	}

	cmd := &command.TriggerAnonBroadcastCommand{
		AnonID:        id,
		UserIDs:       req.UserIDs,
		LanguageCodes: req.LanguageCodes,
	}

	err := h.AnonsUsecase.Command.TriggerAnonBroadcast(ctx, cmd)
	if err != nil {
		apierr.InternalError(c, "failed to trigger broadcast", err.Error())
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
	})
}

// GetAnons godoc
// @Summary      List anonymous broadcasts
// @Description  Returns a list of all anonymous broadcasts
// @Tags         Admin
// @Produce      json
// @Security     BasicAuth
// @Success      200 {array} models.AnonBroadcastResponse
// @Failure      500 {object} apierr.Response
// @Router       /admin/anons [get]
func (h *Handlers) GetAnons(c *gin.Context) {
	ctx := c.Request.Context()

	q := &query.GetAnonsQuery{}
	anons, err := h.AnonsUsecase.Query.GetAnons(ctx, q)
	if err != nil {
		apierr.InternalError(c, "failed to list anons", err.Error())
		return
	}

	response := make([]models.AnonBroadcastResponse, len(anons))
	for i, anon := range anons {
		var replyMarkup map[string]any
		if len(anon.ReplyMarkup) > 0 {
			if err := json.Unmarshal(anon.ReplyMarkup, &replyMarkup); err != nil {
				h.Logger.ErrorContext(ctx, "failed to unmarshal reply markup", err)
			}
		}

		response[i] = models.AnonBroadcastResponse{
			ID:          anon.ID.String(),
			VideoFileID: anon.VideoFileID,
			PhotoFileID: anon.PhotoFileID,
			Message:     anon.Message,
			ReplyMarkup: replyMarkup,
			CreatedAt:   anon.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, response)
}
