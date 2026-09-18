// Package redis is responsible for queuing
package redis

import "github.com/redis/go-redis/v9"

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
