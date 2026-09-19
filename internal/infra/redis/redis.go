// Package redis is responsible for queuing
package redis

import (
	"context"
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
	_ = r.Client.XAdd(ctx, &redis.XAddArgs{
		Stream: "jobs",
		Values: item,
	})
	return nil
}

func (r *Redis) Log(ctx context.Context) {
	streams, err := r.Client.XRead(ctx, &redis.XReadArgs{
		Streams: []string{"jobs"},
	}).Result()
	if err != nil {
		fmt.Println("error logging redis", err.Error())
		return
	}

	fmt.Println("streams: ", streams)
}
