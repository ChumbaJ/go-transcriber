// Package redis is responsible for queuing
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
}

func New(ctx context.Context, redisURL string) (*Redis, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(err)
	}
	rdb := redis.NewClient(opt)

	rdbCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	if err := rdb.Ping(rdbCtx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Redis{
		Client: rdb,
	}, nil
}
