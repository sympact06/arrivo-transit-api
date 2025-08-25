package cache

import (
	"github.com/redis/go-redis/v9"
)

// NewRedis creates a new Redis client.
func NewRedis(dsn string) (*redis.Client, error) {
	opts, err := redis.ParseURL(dsn)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)
	return client, nil
}