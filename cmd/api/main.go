// Package main

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/ChumbaJ/go-transcriber/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	defer func() {
		if value := recover(); value != nil {
			logger.Error(
				"unexpected panic",
				"panic", value,
				"stack", string(debug.Stack()),
			)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	if err := godotenv.Load(); err != nil {
		logger.WarnContext(ctx, "could not load .env, using environment variables", "error", err)
	}
	cfg := config.Load()

	app, err := newApp(ctx, cfg, logger)
	if err != nil {
		logger.ErrorContext(ctx, "application initialization failed", "error", err)
		return
	}

	app.run()
}
