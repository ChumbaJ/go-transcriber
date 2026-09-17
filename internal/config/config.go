// Package config
package config

import "os"

type Config struct {
	Port     string
	Database string
	Bucket   string
	AwsAK    string
	AwsSK    string
}

func Load() *Config {
	return &Config{
		Port:     os.Getenv("PORT"),
		Database: os.Getenv("DATABASE_URL"),
		Bucket:   os.Getenv("S3_BUCKET"),
		AwsAK:    os.Getenv("AWS_ACCESS_KEY"),
		AwsSK:    os.Getenv("AWS_SECRET_KEY"),
	}
}
