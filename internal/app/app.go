package app

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/cache"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/config"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/handlers"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/httpserver"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/repository"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/service"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/token"
)

type App struct {
	Config config.Config
	Logger *slog.Logger
	Pool   *pgxpool.Pool
	Redis  *cache.Client
	Server *httpserver.Server
}

func New(cfg config.Config, logger *slog.Logger) (*App, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.Database.URL)
	if err != nil {
		return nil, err
	}
	poolCfg.MaxConns = cfg.Database.MaxConns
	poolCfg.MinConns = cfg.Database.MinConns
	poolCfg.MaxConnLifetime = cfg.Database.MaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.Database.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, err
	}

	redisClient := cache.New(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
		PoolSize:     cfg.Redis.PoolSize,
	})

	repo := repository.New(pool, cfg.App.EncryptionKey)
	tokens := token.New(cfg.JWT.Secret, cfg.JWT.Issuer, cfg.JWT.AccessTTL)

	botSvc := service.NewBotService(repo, redisClient, cfg.App.EncryptionKey)

	dbPing := func() error { return pool.Ping(context.Background()) }
	redisPing := func() error {
		if redisClient == nil {
			return nil
		}
		return redisClient.Ping(context.Background())
	}
	health := handlers.NewHealthHandler(dbPing, redisPing)

	deps := httpserver.RouterDeps{
		Bots:           handlers.NewBotHandler(botSvc),
		Health:         health,
		Tokens:         tokens,
		ServiceAPIKey:  cfg.ServiceAPIKey,
		AllowedOrigins: cfg.App.AllowedOrigins,
	}
	router := httpserver.NewRouter(deps)
	server := httpserver.New(cfg.Server.Addr, router, cfg.Server.ReadTimeout, cfg.Server.WriteTimeout, cfg.Server.IdleTimeout, cfg.Server.ShutdownTimeout)

	return &App{Config: cfg, Logger: logger, Pool: pool, Redis: redisClient, Server: server}, nil
}

func (a *App) Close() error {
	if a.Redis != nil {
		_ = a.Redis.Close()
	}
	if a.Pool != nil {
		a.Pool.Close()
	}
	return nil
}
