package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ChumbaJ/go-transcriber/internal/api"
	"github.com/ChumbaJ/go-transcriber/internal/config"
	"github.com/ChumbaJ/go-transcriber/internal/infra/postgres"
	"github.com/ChumbaJ/go-transcriber/internal/infra/redis"
	"github.com/ChumbaJ/go-transcriber/internal/infra/s3"
	"github.com/ChumbaJ/go-transcriber/internal/infra/transcription"
	"github.com/ChumbaJ/go-transcriber/internal/job"
	"github.com/ChumbaJ/go-transcriber/internal/queue"
	"github.com/ChumbaJ/go-transcriber/internal/worker"
)

type app struct {
	server *http.Server
	db     *postgres.Postgres
	logger *slog.Logger
}

const (
	workersNum = 3
	streamName = "chunks"
	groupName  = "transcribers"
)

func newApp(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*app, error) {
	pg, err := postgres.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("init postgres: %w", err)
	}
	storage, err := s3.New(ctx, cfg.Bucket)
	if err != nil {
		pg.Close()
		return nil, fmt.Errorf("init storage: %w", err)
	}

	rdb, err := redis.New(ctx, cfg.RedisURL)
	if err != nil {
		pg.Close()
		return nil, fmt.Errorf("init redis: %w", err)
	}

	if err := rdb.EnsureGroup(ctx, streamName, groupName); err != nil {
		pg.Close()
		rdb.Client.Close()
		return nil, fmt.Errorf("init redis: %w", err)
	}
	queue := queue.New(rdb, streamName, groupName)

	jobSrv := job.NewService(
		pg.Jobs,
		pg.Chunks,
		storage,
		queue,
	)

	tc := transcription.New(cfg)

	wp := worker.NewPool(
		workersNum,
		queue,
		storage,
		pg.Jobs,
		pg.Chunks,
		pg.Transcripts,
		tc,
		logger,
	)
	go wp.Run(ctx)

	h := api.NewHandler(jobSrv, logger)
	r := api.NewRouter(h)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	return &app{
		server: srv,
		db:     pg,
		logger: logger,
	}, nil
}

func (a *app) run() error {
	if err := a.server.ListenAndServe(); err != nil {
		return err
	}
	return nil
}
