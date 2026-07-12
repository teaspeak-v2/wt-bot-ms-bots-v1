package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/models"
)

// SpawnResponse is the runner response for spawning a container.
type SpawnResponse struct {
	ContainerID string `json:"container_id"`
	Status      string `json:"status"`
}

// Client is a lightweight HTTP client for the bot runner service.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// New creates a new runner client.
func New(baseURL, apiKey string, timeout time.Duration) *Client {
	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: timeout},
	}
}

// Spawn asks the runner to create and start a bot container.
func (c *Client) Spawn(ctx context.Context, botID uuid.UUID) (*SpawnResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/api/v1/bots/%s/spawn", c.baseURL, botID.String()), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("runner returned status %d", resp.StatusCode)
	}

	var result SpawnResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Stop asks the runner to stop a bot container.
func (c *Client) Stop(ctx context.Context, botID uuid.UUID) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/api/v1/bots/%s/stop", c.baseURL, botID.String()), nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("runner returned status %d", resp.StatusCode)
	}

	return nil
}

// Status asks the runner for the current status of a bot container.
func (c *Client) Status(ctx context.Context, botID uuid.UUID) (*models.BotContainerStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/bots/%s/status", c.baseURL, botID.String()), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Service-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("runner returned status %d", resp.StatusCode)
	}

	var result models.BotContainerStatus
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}
