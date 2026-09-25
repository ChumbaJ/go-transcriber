package main

import (
	"context"
	"net/http"

	"github.com/ChumbaJ/go-transcriber/internal/api"
	"github.com/ChumbaJ/go-transcriber/internal/config"
	"github.com/ChumbaJ/go-transcriber/internal/infra/redis"
	"github.com/ChumbaJ/go-transcriber/internal/job"
	"github.com/ChumbaJ/go-transcriber/internal/queue"
)

type app struct {
	server *http.Server
}

func newApp(ctx context.Context, cfg *config.Config) *app {

	const (
		streamName = "chunks"
		groupName  = "transcribers"
	)

	rdb := redis.New()
	err := rdb.Init(ctx, streamName, groupName)
	if err != nil {

	}
	queue.New(rdb)

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
