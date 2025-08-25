package cache

import (
	"context"
	"sync"

	"github.com/golang/groupcache/lru"
)

// LRUCache is a simple thread-safe LRU cache.
type LRUCache struct {
	cache *lru.Cache
	mu    sync.Mutex
}

// NewLRUCache creates a new LRU cache with the given max size.
func NewLRUCache(maxEntries int) *LRUCache {
	return &LRUCache{
		cache: lru.New(maxEntries),
	}
}

// Get gets a value from the cache.
func (c *LRUCache) Get(ctx context.Context, key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	val, ok := c.cache.Get(key)
	if !ok {
		return nil, false
	}

	byteVal, ok := val.([]byte)
	if !ok {
		// This should not happen if we only add byte slices
		return nil, false
	}

	return byteVal, true
}

// Set sets a value in the cache.
func (c *LRUCache) Set(ctx context.Context, key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache.Add(key, value)
}