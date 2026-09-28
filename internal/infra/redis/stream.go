package redis

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const readBlockIntervalSec = 5

func (r *Redis) EnsureGroup(ctx context.Context, stream, group string) error {
	err := r.Client.XGroupCreateMkStream(ctx, stream, group, "0").Err()
	if err != nil && !strings.HasPrefix(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("create consumer group: %w", err)
	}

	return nil
}

func (r *Redis) Add(ctx context.Context, stream string, values map[string]any) error {
	if err := r.Client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: values,
	}).Err(); err != nil {
		return fmt.Errorf("xadd: %w", err)
	}
	return nil
}

func (r *Redis) ReadGroup(ctx context.Context, stream, group, consumer string) (*redis.XMessage, error) {
	streams, err := r.Client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  []string{stream, ">"},
		Count:    1,
		Block:    time.Second * readBlockIntervalSec,
	}).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("xreadgroup: %w", err)
	}

	if len(streams) == 0 || len(streams[0].Messages) == 0 {
		return nil, nil
	}

	msg := streams[0].Messages[0]
	return &msg, nil
}

func (r *Redis) Ack(ctx context.Context, stream, group, id string) error {
	if err := r.Client.XAck(ctx, stream, group, id).Err(); err != nil {
		return fmt.Errorf("error xack: %w", err)
	}
	return nil
}

func (r *Redis) AutoClaim(ctx context.Context, stream, group, consumer string, minIdle time.Duration, start string, count int64) ([]redis.XMessage, string, error) {
	messages, next, err := r.Client.XAutoClaim(ctx, &redis.XAutoClaimArgs{
		Stream:   stream,
		Group:    group,
		Consumer: consumer,
		MinIdle:  minIdle,
		Start:    start,
		Count:    count,
	}).Result()
	if err != nil {
		return nil, "", fmt.Errorf("xautoclaim: %w", err)
	}
	return messages, next, nil
}
