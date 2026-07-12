package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/apperror"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/httpserver/middleware"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/models"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/service"
)

type BotHandler struct {
	svc *service.BotService
}

func NewBotHandler(svc *service.BotService) *BotHandler {
	return &BotHandler{svc: svc}
}

func (h *BotHandler) parseID(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, apperror.InvalidRequest("invalid bot id", err)
	}
	return id, nil
}

// List godoc
// @Summary List bots
// @Tags bots
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param search query string false "Search term"
// @Param teamspeak_id query string false "TeamSpeak ID filter"
// @Param enabled query bool false "Enabled filter"
// @Param sort_by query string false "Sort field"
// @Param sort_order query string false "Sort order"
// @Success 200 {object} models.BotListResponse
// @Router /bots [get]
func (h *BotHandler) List(w http.ResponseWriter, r *http.Request) {
	filter, err := h.parseListFilter(r)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	resp, err := h.svc.List(r.Context(), filter)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *BotHandler) parseListFilter(r *http.Request) (models.BotListFilter, error) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	enabled, err := parseBoolPtr(q.Get("enabled"))
	if err != nil {
		return models.BotListFilter{}, apperror.InvalidRequest("invalid enabled filter", err)
	}

	var teamspeakID uuid.UUID
	if t := q.Get("teamspeak_id"); t != "" {
		var err error
		teamspeakID, err = uuid.Parse(t)
		if err != nil {
			return models.BotListFilter{}, apperror.InvalidRequest("invalid teamspeak_id filter", err)
		}
	}

	return models.BotListFilter{
		Search:      strings.TrimSpace(q.Get("search")),
		TeamSpeakID: teamspeakID,
		Enabled:     enabled,
		SortBy:      q.Get("sort_by"),
		SortOrder:   q.Get("sort_order"),
		Limit:       limit,
		Offset:      offset,
	}, nil
}

// Create godoc
// @Summary Create a bot
// @Tags bots
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.CreateBotRequest true "Create request"
// @Success 201 {object} models.BotResponse
// @Router /bots [post]
func (h *BotHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateBotRequest
	if err := readJSON(r, &req); err != nil {
		apperror.WriteJSON(w, validationErr(err))
		return
	}
	resp, err := h.svc.Create(r.Context(), req)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

// GetByID godoc
// @Summary Get a bot by ID
// @Tags bots
// @Security BearerAuth
// @Produce json
// @Param id path string true "Bot ID"
// @Success 200 {object} models.BotResponse
// @Router /bots/{id} [get]
func (h *BotHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	resp, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// Update godoc
// @Summary Update a bot
// @Tags bots
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Bot ID"
// @Param request body models.UpdateBotRequest true "Update request"
// @Success 200 {object} models.BotResponse
// @Router /bots/{id} [patch]
func (h *BotHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	var req models.UpdateBotRequest
	if err := readJSON(r, &req); err != nil {
		apperror.WriteJSON(w, validationErr(err))
		return
	}
	resp, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// Delete godoc
// @Summary Delete a bot
// @Tags bots
// @Security BearerAuth
// @Produce json
// @Param id path string true "Bot ID"
// @Success 200 {object} models.MessageResponse
// @Router /bots/{id} [delete]
func (h *BotHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, models.MessageResponse{Message: "bot deleted"})
}

// Start godoc
// @Summary Start a bot container
// @Tags bots
// @Security BearerAuth
// @Produce json
// @Param id path string true "Bot ID"
// @Success 200 {object} models.BotContainerStatus
// @Router /bots/{id}/start [post]
func (h *BotHandler) Start(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	resp, err := h.svc.Start(r.Context(), id)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// Stop godoc
// @Summary Stop a bot container
// @Tags bots
// @Security BearerAuth
// @Produce json
// @Param id path string true "Bot ID"
// @Success 200 {object} models.MessageResponse
// @Router /bots/{id}/stop [post]
func (h *BotHandler) Stop(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	if err := h.svc.Stop(r.Context(), id); err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, models.MessageResponse{Message: "bot stopped"})
}

// Config godoc
// @Summary Get a bot runtime configuration
// @Tags bots-service
// @Produce json
// @Param id path string true "Bot ID"
// @Success 200 {object} models.BotConfig
// @Router /bots/{id}/config [get]
func (h *BotHandler) Config(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	resp, err := h.svc.GetConfig(r.Context(), id)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// UpdateStatus godoc
// @Summary Update a bot runtime status
// @Tags bots-service
// @Accept json
// @Produce json
// @Param id path string true "Bot ID"
// @Param request body models.UpdateStatusRequest true "Status request"
// @Success 200 {object} models.MessageResponse
// @Router /bots/{id}/status [post]
func (h *BotHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	var req models.UpdateStatusRequest
	if err := readJSON(r, &req); err != nil {
		apperror.WriteJSON(w, validationErr(err))
		return
	}
	if err := h.svc.UpdateStatus(r.Context(), id, req.Status); err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, models.MessageResponse{Message: "status updated"})
}

var _ = middleware.UserID // ensure middleware is referenced
