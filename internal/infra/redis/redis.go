// Package redis is responsible for queuing
package redis

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
}

func New(ctx context.Context, redisURL string) *Redis {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(err)
	}
	rdb := redis.NewClient(opt)

	return &Redis{
		Client: rdb,
	}
}

func (r *Redis) Init(ctx context.Context) error {
	err := r.Client.XGroupCreateMkStream(ctx, "jobs", "transcribers", "0").Err()
	if err != nil && !strings.HasPrefix(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("create consumer group: %w", err)
	}

	return nil
}

func (r *Redis) Run(ctx context.Context) {
	for {
		// Consume job

		// create a worker for a job
	}
}
