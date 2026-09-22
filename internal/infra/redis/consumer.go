// Package redis
package redis

import (
	"context"
	"fmt"

	"github.com/ChumbaJ/go-transcriber/internal/job"
	"github.com/redis/go-redis/v9"
)

func (r *Redis) ConsumeJobChunk(ctx context.Context, name string) (*job.QueueItem, error) {
	streams, err := r.Client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    "transcribers",
		Consumer: name,
		Streams:  []string{"jobs", ">"},
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("xreadgroup: %w", err)
	}

	chunkStream := streams[0]

	fmt.Println("DO JOB")

	return nil
}
