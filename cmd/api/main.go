// Package main

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ChumbaJ/go-transcriber/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	signalCtx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	if err := godotenv.Load(); err != nil {
		logger.WarnContext(signalCtx, "could not load .env, using environment variables", "error", err)
	}
	cfg := config.Load()

	app, err := newApp(signalCtx, cfg, logger)
	if err != nil {
		logger.ErrorContext(signalCtx, "application initialization failed", "error", err)
		return
	}

	go app.run()
	defer app.db.Close()

	fmt.Println("server is listening on port ", cfg.Port)

	<-signalCtx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := app.server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown error", "error", err)
	}
}
