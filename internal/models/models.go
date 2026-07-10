package models

import (
	"time"

	"github.com/google/uuid"
)

// UserRole mirrors the role enum from the users microservice.
type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// Bot represents a managed TeamSpeak bot.
type Bot struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	TeamSpeakID uuid.UUID `json:"teamspeak_id"`
	OwnerID     uuid.UUID `json:"owner_id"`
	Nickname    string    `json:"nickname"`
	Greeting    string    `json:"greeting"`
	HelpMessage string    `json:"help_message"`
	Enabled     bool      `json:"enabled"`
	Status      string    `json:"status"`
	APIKey      string    `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BotResponse is the public view of a bot.
type BotResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	TeamSpeakID uuid.UUID `json:"teamspeak_id"`
	OwnerID     uuid.UUID `json:"owner_id"`
	Nickname    string    `json:"nickname"`
	Greeting    string    `json:"greeting"`
	HelpMessage string    `json:"help_message"`
	Enabled     bool      `json:"enabled"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BotConfig is the runtime configuration delivered to a bot instance.
type BotConfig struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	TeamSpeakID uuid.UUID `json:"teamspeak_id"`
	Nickname    string    `json:"nickname"`
	Greeting    string    `json:"greeting"`
	HelpMessage string    `json:"help_message"`
	Enabled     bool      `json:"enabled"`
	Status      string    `json:"status"`
}

// BotListResponse is a paginated list of bots.
type BotListResponse struct {
	Bots   []BotResponse `json:"bots"`
	Total  int64         `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

// BotListFilter is used for listing bots.
type BotListFilter struct {
	Search      string
	TeamSpeakID uuid.UUID
	Enabled     *bool
	OwnerID     uuid.UUID
	SortBy      string
	SortOrder   string
	Limit       int
	Offset      int
}

// CreateBotRequest is the payload for creating a new bot.
type CreateBotRequest struct {
	Name        string    `json:"name" validate:"required,min=2,max=120"`
	TeamSpeakID uuid.UUID `json:"teamspeak_id" validate:"required"`
	Nickname    string    `json:"nickname" validate:"required,min=1,max=64"`
	Greeting    string    `json:"greeting" validate:"max=1024"`
	HelpMessage string    `json:"help_message" validate:"max=1024"`
	Enabled     bool      `json:"enabled"`
}

// UpdateBotRequest is the payload for updating a bot.
type UpdateBotRequest struct {
	Name        *string    `json:"name,omitempty" validate:"omitempty,min=2,max=120"`
	TeamSpeakID *uuid.UUID `json:"teamspeak_id,omitempty"`
	Nickname    *string    `json:"nickname,omitempty" validate:"omitempty,min=1,max=64"`
	Greeting    *string    `json:"greeting,omitempty" validate:"omitempty,max=1024"`
	HelpMessage *string    `json:"help_message,omitempty" validate:"omitempty,max=1024"`
	Enabled     *bool      `json:"enabled,omitempty"`
}

// UpdateStatusRequest is the payload for updating a bot runtime status.
type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

// MessageResponse is a generic message response.
type MessageResponse struct {
	Message string `json:"message"`
}

// ErrorResponse is a generic error response.
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// ClaimsContextKey is the context key for JWT claims.
type ClaimsContextKey struct{}
