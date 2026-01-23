package cache

import (
	"context"
	"time"
)

// Cache defines an interface for caching API responses.
type Cache interface {
	// Get retrieves a cached value by key.
	Get(ctx context.Context, key string) ([]byte, error)
	// Set stores a value in the cache with an optional TTL.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	// Delete removes a cached value.
	Delete(ctx context.Context, key string) error
}

// MemoryCache is an in-memory cache implementation.
type MemoryCache struct {
	// implementation omitted for design phase
}

func (c *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, nil
}

func (c *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return nil
}

func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	return nil
}

// FileCache is a file-based cache implementation.
type FileCache struct {
	// implementation omitted for design phase
}

func (c *FileCache) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, nil
}

func (c *FileCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return nil
}

func (c *FileCache) Delete(ctx context.Context, key string) error {
	return nil
}
