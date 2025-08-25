package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Config holds the configuration for the application.
// It is populated from environment variables.
type Config struct {
	PostgresDSN string `envconfig:"POSTGRES_DSN" required:"true"`
	RedisDSN    string `envconfig:"REDIS_DSN" required:"true"`
	Port        string `envconfig:"PORT" default:"8080"`
}

// Load returns a new Config struct populated from environment variables.
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}