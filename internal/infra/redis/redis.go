// Package redis is responsible for queuing
package redis

import (
	"context"
	"fmt"

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

	if err := rdb.XGroupCreate(ctx, "jobs", "transcribers", "0").Err(); err != nil {
		panic(err)
	}

	return &Redis{
		Client: rdb,
	}
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
