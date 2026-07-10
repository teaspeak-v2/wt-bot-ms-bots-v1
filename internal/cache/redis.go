package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/models"
)

type Client struct {
	client *redis.Client
}

func New(opts *redis.Options) *Client {
	if opts == nil || opts.Addr == "" {
		return nil
	}
	return &Client{client: redis.NewClient(opts)}
}

func (c *Client) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}

func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Ping(ctx).Err()
}

func (c *Client) SetBot(ctx context.Context, bot *models.Bot, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	b, err := json.Marshal(bot)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, botKey(bot.ID), b, ttl).Err()
}

func (c *Client) GetBot(ctx context.Context, id uuid.UUID) (*models.Bot, error) {
	if c == nil || c.client == nil {
		return nil, redis.Nil
	}
	val, err := c.client.Get(ctx, botKey(id)).Bytes()
	if err != nil {
		return nil, err
	}
	var bot models.Bot
	if err := json.Unmarshal(val, &bot); err != nil {
		return nil, err
	}
	return &bot, nil
}

func (c *Client) DeleteBot(ctx context.Context, id uuid.UUID) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Del(ctx, botKey(id)).Err()
}

func botKey(id uuid.UUID) string { return fmt.Sprintf("bots:%s", id.String()) }
