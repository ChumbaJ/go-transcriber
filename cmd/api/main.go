// Package main

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ChumbaJ/go-transcriber/internal/api"
	"github.com/ChumbaJ/go-transcriber/internal/config"
	"github.com/ChumbaJ/go-transcriber/internal/infra/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	cfg := config.Load()

	pool, err := pgxpool.New(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("error while connecting to database: %w", err)
	}

	jobRepo := postgres.NewJobRepo(pool)

	h := api.NewHandler()
	r := api.NewRouter(h)

	srv := &http.Server{
		Addr:    ":3030",
		Handler: r,
	}

	errCh := make(chan error)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()

	select {

	case <-ctx.Done():
		log.Println("shutdown signal recieved")
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)

	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownContext); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	log.Println("server stopped")

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("fatal %v", err)
	}
}
