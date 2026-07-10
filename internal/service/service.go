package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/apperror"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/cache"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/httpserver/middleware"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/models"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/repository"
)

// BotService handles business logic for bots.
type BotService struct {
	repo   repository.BotRepository
	cache  *cache.Client
	encKey string
}

// NewBotService creates a new service.
func NewBotService(repo repository.BotRepository, cache *cache.Client, encKey string) *BotService {
	return &BotService{repo: repo, cache: cache, encKey: encKey}
}

func (s *BotService) toResponse(bot *models.Bot) *models.BotResponse {
	return &models.BotResponse{
		ID:          bot.ID,
		Name:        bot.Name,
		TeamSpeakID: bot.TeamSpeakID,
		OwnerID:     bot.OwnerID,
		Nickname:    bot.Nickname,
		Greeting:    bot.Greeting,
		HelpMessage: bot.HelpMessage,
		Enabled:     bot.Enabled,
		Status:      bot.Status,
		CreatedAt:   bot.CreatedAt,
		UpdatedAt:   bot.UpdatedAt,
	}
}

func (s *BotService) isAdmin(ctx context.Context) bool {
	return middleware.Role(ctx) == models.RoleAdmin
}

func (s *BotService) userID(ctx context.Context) uuid.UUID {
	id, _ := uuid.Parse(middleware.UserID(ctx))
	return id
}

func (s *BotService) checkOwnership(ctx context.Context, bot *models.Bot) error {
	if s.isAdmin(ctx) {
		return nil
	}
	if uid := s.userID(ctx); uid != uuid.Nil && bot.OwnerID == uid {
		return nil
	}
	return apperror.Forbidden("not owner", nil)
}

// List returns a paginated list of bots.
func (s *BotService) List(ctx context.Context, filter models.BotListFilter) (*models.BotListResponse, error) {
	if !s.isAdmin(ctx) {
		filter.OwnerID = s.userID(ctx)
	}
	list, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	resp := &models.BotListResponse{
		Bots:   make([]models.BotResponse, 0, len(list)),
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}
	for i := range list {
		resp.Bots = append(resp.Bots, *s.toResponse(&list[i]))
	}
	return resp, nil
}

// Create creates a new bot record.
func (s *BotService) Create(ctx context.Context, req models.CreateBotRequest) (*models.BotResponse, error) {
	ownerID := s.userID(ctx)
	if ownerID == uuid.Nil {
		return nil, apperror.Unauthorized("missing authenticated user", nil)
	}

	bot := &models.Bot{
		ID:          uuid.New(),
		Name:        req.Name,
		TeamSpeakID: req.TeamSpeakID,
		OwnerID:     ownerID,
		Nickname:    req.Nickname,
		Greeting:    defaultGreeting(req.Greeting),
		HelpMessage: defaultHelpMessage(req.HelpMessage),
		Enabled:     req.Enabled,
		Status:      "offline",
		APIKey:      "",
	}

	if err := s.repo.Create(ctx, bot); err != nil {
		return nil, mapRepoErr(err)
	}
	if s.cache != nil {
		_ = s.cache.DeleteBot(ctx, bot.ID)
	}
	return s.toResponse(bot), nil
}

// GetByID returns a bot by ID.
func (s *BotService) GetByID(ctx context.Context, id uuid.UUID) (*models.BotResponse, error) {
	if s.cache != nil {
		if cached, err := s.cache.GetBot(ctx, id); err == nil {
			if err := s.checkOwnership(ctx, cached); err != nil {
				return nil, err
			}
			return s.toResponse(cached), nil
		}
	}
	bot, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := s.checkOwnership(ctx, bot); err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.SetBot(ctx, bot, 5*time.Minute)
	}
	return s.toResponse(bot), nil
}

// GetConfig returns the runtime configuration for a bot instance.
func (s *BotService) GetConfig(ctx context.Context, id uuid.UUID) (*models.BotConfig, error) {
	bot, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return &models.BotConfig{
		ID:          bot.ID,
		Name:        bot.Name,
		TeamSpeakID: bot.TeamSpeakID,
		Nickname:    bot.Nickname,
		Greeting:    bot.Greeting,
		HelpMessage: bot.HelpMessage,
		Enabled:     bot.Enabled,
		Status:      bot.Status,
	}, nil
}

// Update updates a bot record.
func (s *BotService) Update(ctx context.Context, id uuid.UUID, req models.UpdateBotRequest) (*models.BotResponse, error) {
	bot, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := s.checkOwnership(ctx, bot); err != nil {
		return nil, err
	}

	if req.Greeting != nil {
		*req.Greeting = defaultGreeting(*req.Greeting)
	}
	if req.HelpMessage != nil {
		*req.HelpMessage = defaultHelpMessage(*req.HelpMessage)
	}

	updated, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if s.cache != nil {
		_ = s.cache.DeleteBot(ctx, id)
	}
	return s.toResponse(updated), nil
}

// Delete removes a bot record.
func (s *BotService) Delete(ctx context.Context, id uuid.UUID) error {
	bot, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return mapRepoErr(err)
	}
	if err := s.checkOwnership(ctx, bot); err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return mapRepoErr(err)
	}
	if s.cache != nil {
		_ = s.cache.DeleteBot(ctx, id)
	}
	return nil
}

// UpdateStatus updates the runtime status of a bot.
func (s *BotService) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		return mapRepoErr(err)
	}
	if s.cache != nil {
		_ = s.cache.DeleteBot(ctx, id)
	}
	return nil
}

func defaultGreeting(v string) string {
	if v == "" {
		return "Welcome to the server!"
	}
	return v
}

func defaultHelpMessage(v string) string {
	if v == "" {
		return "Available commands: !help"
	}
	return v
}

func mapRepoErr(err error) error {
	if repository.IsNotFound(err) {
		return apperror.NotFound("resource not found", err)
	}
	if repository.IsUniqueViolation(err) {
		return apperror.Conflict("resource already exists", err)
	}
	return apperror.Internal("internal server error", err)
}
