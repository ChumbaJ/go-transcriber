// Package redis
package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ChumbaJ/go-transcriber/internal/job"
	"github.com/redis/go-redis/v9"
)

func (r *Redis) ProduceJob(ctx context.Context, item *job.QueueItem) error {
	b, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("produce job, marshaling: %w", err)
	}

	if err := r.Client.XAdd(ctx, &redis.XAddArgs{
		Stream: "jobs",
		Values: map[string]any{"job": b},
	}).Err(); err != nil {
		return fmt.Errorf("xadd: %w", err)
	}
	return nil
}
