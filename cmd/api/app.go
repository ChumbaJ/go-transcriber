package main

import (
	"context"
	"net/http"

	"github.com/ChumbaJ/go-transcriber/internal/api"
	"github.com/ChumbaJ/go-transcriber/internal/config"
	"github.com/ChumbaJ/go-transcriber/internal/infra/postgres"
	"github.com/ChumbaJ/go-transcriber/internal/infra/redis"
	"github.com/ChumbaJ/go-transcriber/internal/infra/s3"
	"github.com/ChumbaJ/go-transcriber/internal/job"
	"github.com/ChumbaJ/go-transcriber/internal/queue"
)

type app struct {
	server *http.Server
}

func newApp(ctx context.Context, cfg *config.Config) *app {
	// Add postres.NewPool()

	jobsRepo := postgres.NewJobRepo()
	chunksRepo := postgres.NewChunksRepository()
	transcripRepo := postgres.NewTranscripRepo()
	storage := s3.New(ctx, cfg.Bucket)

	const (
		streamName = "chunks"
		groupName  = "transcribers"
	)

	rdb := redis.New(cfg.RedisURL)
	err := rdb.Init(ctx, streamName, groupName)
	if err != nil {
	}
	queue := queue.New(rdb, streamName, groupName)

	jobSrv := job.NewService(
		jobRepo,
		chunksRepo,
		storage,
		queue,
	)

	h := api.NewHandler(jobSrv)
	r := api.NewRouter(h)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	return &app{
		server: srv,
	}
}

func (a *app) run() {
	a.server.ListenAndServe()
}
