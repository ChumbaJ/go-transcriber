// Package redis
package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func (r *Redis) ConsumeJob(ctx context.Context) error {
	streams, err := r.Client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    "transcribers",
		Consumer: "transcriber-consumer",
		Streams:  []string{"jobs", ">"},
	}).Result()
	if err != nil {
		return fmt.Errorf("xreadgroup: %w", err)
	}

	for _, s := range streams {
		for _, m := range s.Messages {
			fmt.Println("stream: ", s.Stream)
			fmt.Println("message:", m)
		}
	}

	return nil
}
