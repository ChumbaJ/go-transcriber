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
	"github.com/ChumbaJ/go-transcriber/internal/job"
	"github.com/ChumbaJ/go-transcriber/internal/queue"
	"github.com/ChumbaJ/go-transcriber/internal/worker"
)

type app struct {
	server *http.Server
	db     *postgres.Postgres
}

func newApp(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*app, error) {
	// TODO: add ping
	pg, err := postgres.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("init postgres: %w", err)
	}
	storage := s3.New(ctx, cfg.Bucket)

	const (
		workersNum = 3
		streamName = "chunks"
		groupName  = "transcribers"
	)

	rdb := redis.New(cfg.RedisURL)
	if err := rdb.Init(ctx, streamName, groupName); err != nil {
		pg.Close()
		return nil, fmt.Errorf("init redis: %w", err)
	}
	queue := queue.New(rdb, streamName, groupName)

	jobSrv := job.NewService(
		pg.Jobs,
		pg.Chunks,
		storage,
		queue,
	)

	wp := worker.NewPool(
		workersNum,
		queue,
		storage,
		pg.Jobs,
		pg.Chunks,
		pg.Transcripts,
	)
	// TODO: handle exceptions here
	go wp.Run(ctx)

	h := api.NewHandler(jobSrv)
	r := api.NewRouter(h)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	return &app{
		server: srv,
		db:     pg,
	}, nil
}

func (a *app) run() {
	defer a.db.Close()
	if err := a.server.ListenAndServe(); err != nil {

		return
	}
}
