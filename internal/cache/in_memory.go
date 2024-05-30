package cache

import (
	"sync"
	"time"
)

// InMemoryCache represents an in-memory cache with expiration and periodic cleanup capabilities.
type InMemoryCache struct {
	mu              sync.RWMutex
	data            map[string]InMemoryCacheItem
	expiration      time.Duration
	cleanupInterval time.Duration
	stopChan        chan struct{}
}

// InMemoryCacheItem represents an item stored in the in-memory cache.
type InMemoryCacheItem struct {
	value      interface{}
	expiryTime time.Time
}

// NewInMemoryCache initializes a new cache with a given expiration duration and cleanup interval.
//
// Parameters:
// - expiration: The duration for which an item should remain in the cache before expiring.
// - cleanupInterval: The interval at which the cache is cleaned up to remove expired items.
//
// Returns:
// - *Cache: A pointer to the initialized Cache instance.
func NewInMemoryCache(expiration time.Duration, cleanupInterval time.Duration) *InMemoryCache {
	cache := &InMemoryCache{
		data:            make(map[string]InMemoryCacheItem),
		expiration:      expiration,
		cleanupInterval: cleanupInterval,
		stopChan:        make(chan struct{}),
	}

	go cache.cleanup()
	return cache
}

// Get retrieves an item from the cache.
//
// Parameters:
// - key: The key for the item to retrieve.
//
// Returns:
// - interface{}: The value associated with the key, or nil if the key does not exist or the item has expired.
// - bool: A boolean indicating whether the key was found and the item is valid.
func (cache *InMemoryCache) Get(key string) (interface{}, bool) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	item, exists := cache.data[key]
	if !exists || time.Now().After(item.expiryTime) {
		return nil, false
	}

	return item.value, true
}

// Set stores an item in the cache.
//
// Parameters:
// - key: The key for the item to store.
// - value: The value to store in the cache.
func (cache *InMemoryCache) Set(key string, value interface{}) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.data[key] = InMemoryCacheItem{
		value:      value,
		expiryTime: time.Now().Add(cache.expiration),
	}
}

// Delete removes an item from the cache.
//
// Parameters:
// - key: The key for the item to remove.
func (cache *InMemoryCache) Delete(key string) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	delete(cache.data, key)
}

// cleanup periodically removes expired items from the cache.
// This method runs in a separate goroutine and checks for expired items at the interval specified by cleanupInterval.
func (cache *InMemoryCache) cleanup() {
	ticker := time.NewTicker(cache.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cache.mu.Lock()
			for key, item := range cache.data {
				if time.Now().After(item.expiryTime) {
					delete(cache.data, key)
				}
			}
			cache.mu.Unlock()
		case <-cache.stopChan:
			return
		}
	}
}

// Stop stops the cleanup goroutine.
// This method closes the stopChan to signal the cleanup goroutine to exit.
func (cache *InMemoryCache) Stop() {
	close(cache.stopChan)
}
