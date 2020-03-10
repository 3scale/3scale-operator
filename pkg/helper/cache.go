package helper

import (
	"errors"
)

// Errors
var (
	ErrNonExistentKey = errors.New("non-existent key")
)

// MemoryCache implements a simple memory cache
// No expiration
// No eviction policies
// Not threadsafe
type MemoryCache struct {
	store map[string]interface{}
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		store: make(map[string]interface{}),
	}
}

func (c *MemoryCache) Exists(key string) bool {
	_, ok := c.store[key]
	return ok
}

func (c *MemoryCache) Get(key string) (interface{}, error) {
	if !c.Exists(key) {
		return nil, ErrNonExistentKey
	}

	return c.store[key], nil
}

func (c *MemoryCache) Put(key string, data interface{}) {
	c.store[key] = data
}
