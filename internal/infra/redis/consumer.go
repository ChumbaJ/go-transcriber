// Package redis
package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func (r *Redis) ConsumeJob(ctx context.Context) error {
	_, err := r.Client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    "transcribers",
		Consumer: "transcriber-consumer",
		Streams:  []string{"jobs", ">"},
	}).Result()
	if err != nil {
		return fmt.Errorf("xreadgroup: %w", err)
	}

	fmt.Println("DO JOB")

	return nil
}
