// Package config
package config

import (
	"errors"
	"os"
)

type Config struct {
	Port         string
	Database     string
	Bucket       string
	RedisURL     string
	OpenAiAPIKey string
}

func Load() (*Config, error) {
	c := &Config{
		Port:         os.Getenv("PORT"),
		Database:     os.Getenv("DATABASE_URL"),
		Bucket:       os.Getenv("S3_BUCKET"),
		RedisURL:     os.Getenv("REDIS_URL"),
		OpenAiAPIKey: os.Getenv("OPENAI_API_KEY"),
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Config) Validate() error {
	if c.Database == "" {
		return errors.New("DATABASE_URL is required")
	}
	if c.RedisURL == "" {
		return errors.New("REDIS_URL is required")
	}
	if c.Bucket == "" {
		return errors.New("BUCKET is required")
	}
	if c.OpenAiAPIKey == "" {
		return errors.New("OPENAI_API_KEY  required")
	}
	return nil
}
