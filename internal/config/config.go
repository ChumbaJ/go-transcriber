// Package config
package config

import "os"

type Config struct {
	Port     string
	Database string
}

func Load() *Config {
	return &Config{
		Port:     os.Getenv("PORT"),
		Database: os.Getenv("DATABASE"),
	}
}
