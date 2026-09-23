// Package redis
package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ChumbaJ/go-transcriber/internal/job"
	"github.com/redis/go-redis/v9"
)

func (r *Redis) ConsumeJobChunk(ctx context.Context, name string) (*job.QueueMessage, error) {
	streams, err := r.Client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    "transcribers",
		Consumer: name,
		Streams:  []string{"jobs", ">"},
		Count:    1,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("xreadgroup: %w", err)
	}

	msg := streams[0].Messages[0]

	raw, ok := msg.Values["job"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid job payload")
	}

	var item job.QueueItem

	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		return nil, fmt.Errorf("unmarshal job: %w", err)
	}

	return &job.QueueMessage{
		ID:   msg.ID,
		Item: item,
	}, nil
}

func (r *Redis) Dequeue(ctx context.Context, name string) (*job.QueueMessage, error) {
	return r.ConsumeJobChunk(ctx, name)
}
