// Package queue allows servies to manipulate queue (an abstraction above underlying redis)
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type stream interface {
	Add(ctx context.Context, stream string, values map[string]any) error
	ReadGroup(ctx context.Context, stream, group, consumer string) (*redis.XMessage, error)
	Ack(ctx context.Context, stream, group, id string) error
	AutoClaim(ctx context.Context, stream, group, consumer string, minIdle time.Duration, start string, count int64) ([]redis.XMessage, string, error)
}

type Item struct {
	JobID      int64
	ChunkOrder int
	Addr       string
}

type Message struct {
	ID   string
	Item Item
}

type Queue struct {
	stream     stream
	streamName string
	groupName  string
}

const (
	claimMinIdle   = 2 * time.Minute
	claimBatchSize = 1
	payloadField   = "chunk"
)

func New(s stream, streamName, groupName string) *Queue {
	return &Queue{stream: s, streamName: streamName, groupName: groupName}
}

func (q *Queue) Enqueue(ctx context.Context, item *Item) error {
	b, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("marshal item: %w", err)
	}

	if err := q.stream.Add(ctx, q.streamName, map[string]any{payloadField: string(b)}); err != nil {
		return fmt.Errorf("stream add: %w", err)
	}

	return nil
}

func (q *Queue) Dequeue(ctx context.Context, consumerName string) (*Message, error) {
	msg, err := q.stream.ReadGroup(ctx, q.streamName, q.groupName, consumerName)
	if err != nil {
		return nil, fmt.Errorf("dequeue: %w", err)
	}

	raw, ok := msg.Values[payloadField].(string)
	if !ok {
		return nil, fmt.Errorf("invalid payload: expected string")
	}

	var item Item
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		return nil, fmt.Errorf("unmarshal item: %w", err)
	}

	return &Message{
		ID:   msg.ID,
		Item: item,
	}, nil
}

func (q *Queue) Confirm(ctx context.Context, messageID string) error {
	return q.stream.Ack(ctx, q.streamName, q.groupName, messageID)
}

func (q *Queue) ClaimStale(
	ctx context.Context,
	consumerName string,
) (*Message, error) {
	start := "0-0"

	for {
		messages, next, err := q.stream.AutoClaim(
			ctx,
			q.streamName,
			q.groupName,
			consumerName,
			claimMinIdle,
			start,
			claimBatchSize,
		)
		if err != nil {
			return nil, fmt.Errorf("claim stale: %w", err)
		}

		if len(messages) > 0 {
			msg := messages[0]

			raw, ok := msg.Values[payloadField].(string)
			if !ok {
				return nil, fmt.Errorf("invalid payload: expected string")
			}

			var item Item
			if err := json.Unmarshal([]byte(raw), &item); err != nil {
				return nil, fmt.Errorf("unmarshal item: %w", err)
			}

			return &Message{
				ID:   msg.ID,
				Item: item,
			}, nil
		}

		if next == "0-0" {
			return nil, nil
		}

		start = next
	}
}
