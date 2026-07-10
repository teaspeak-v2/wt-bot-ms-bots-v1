package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/app"
	"github.com/teaspeak-v2/wt-bot-ms-bots-v1/internal/config"
)

// @title WT-Bot Bots API
// @version 0.1.0
// @description WT-Bot bots microservice for managing bot configuration and lifecycle.
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	migrateMode := flag.String("migrate", "", "run migrations only: up or down")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if err := app.RunMigrations(cfg.Database.URL, *migrateMode); err != nil {
		logger.Error("migrations failed", "error", err)
		os.Exit(1)
	}
	if *migrateMode != "" {
		return
	}

	application, err := app.New(cfg, logger)
	if err != nil {
		logger.Error("app init failed", "error", err)
		os.Exit(1)
	}
	defer func() { _ = application.Close() }()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := application.Server.Run(ctx); err != nil {
		logger.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}
