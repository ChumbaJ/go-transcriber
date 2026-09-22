// Package redis is responsible for queuing
package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ChumbaJ/go-transcriber/internal/job"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
}

func New(redisURL string) *Redis {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(err)
	}
	rdb := redis.NewClient(opt)

	return &Redis{
		Client: rdb,
	}
}

func (r *Redis) Push(ctx context.Context, item *job.QueueItem) error {
	payload, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("marshal job item: %w", err)
	}

	if err := r.Client.XAdd(ctx, &redis.XAddArgs{
		Stream: "jobs",
		Values: map[string]any{"job": payload},
	}).Err(); err != nil {
		return fmt.Errorf("XAdd: %w", err)
	}
	return nil
}

func (r *Redis) Log(ctx context.Context) {
	res, err := r.Client.XRange(ctx, "jobs", "-", "+").Result()
	if err != nil {
		fmt.Println("error logging redis", err.Error())
		return
	}

	fmt.Println("stream jobs: ", res)
}
