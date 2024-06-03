package integration

import (
	"github.com/mguley/web-scraper-v1/internal/cache"
	"strconv"
	"sync"
	"testing"
	"time"
)

const (
	expiration      = 5 * time.Minute
	cleanupInterval = 1 * time.Minute
)

// TestCacheInitialization verifies that the InMemoryCache initializes correctly and is not nil.
// It also checks that the Stop method can be called without errors to terminate the cleanup goroutine.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestCacheInitialization(t *testing.T) {
	inMemoryCache := cache.NewInMemoryCache(expiration, cleanupInterval)
	defer inMemoryCache.Stop()

	if inMemoryCache == nil {
		t.Error("Failed to initialize in memory cache.")
	}
}

// TestSetAndGet checks the ability of the cache to store and retrieve items correctly.
// It verifies that the values are retrieved as expected and handles non-existent key appropriately.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestSetAndGet(t *testing.T) {
	inMemoryCache := cache.NewInMemoryCache(expiration, cleanupInterval)
	defer inMemoryCache.Stop()

	key := "testKey"
	value := "testValue"
	inMemoryCache.Set(key, value)

	retrievedValue, found := inMemoryCache.Get(key)
	if !found || retrievedValue != value {
		t.Errorf("Failed to retrieve the correct value: got %v want %v", retrievedValue, value)
	}

	// Test retrieving non-existent key
	_, found = inMemoryCache.Get("nonExistentKey")
	if found {
		t.Error("Non-existent key should not be found in the cache.")
	}
}

// TestExpiration verifies that items in the cache expire correctly based on the expiration duration set during initialization.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestExpiration(t *testing.T) {
	inMemoryCache := cache.NewInMemoryCache(1*time.Second, 10*time.Minute)
	defer inMemoryCache.Stop()

	key := "testKey"
	value := "testValue"
	inMemoryCache.Set(key, value)

	time.Sleep(2 * time.Second) // Wait for the item to expire

	_, found := inMemoryCache.Get(key)
	if found {
		t.Error("Expired item was retrieved from the cache")
	}
}

// TestCleanup verifies that the periodic cleanup process correctly removes expired items from the cache.
// It also includes a stress test to ensure cleanup works under load.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestCleanup(t *testing.T) {
	inMemoryCache := cache.NewInMemoryCache(1*time.Second, 500*time.Millisecond)
	defer inMemoryCache.Stop()

	key := "testKey"
	value := "testValue"
	inMemoryCache.Set(key, value)

	time.Sleep(1500 * time.Millisecond) // Wait longer than the cleanup interval

	_, exists := inMemoryCache.Get(key)
	if exists {
		t.Error("Cleanup did not remove the expired item")
	}

	// Test cleanup under load
	for i := 0; i < 1000; i++ {
		inMemoryCache.Set("key"+strconv.Itoa(i), "value"+strconv.Itoa(i))
	}
	time.Sleep(2 * time.Second)
	for i := 0; i < 1000; i++ {
		_, exists := inMemoryCache.Get("key" + strconv.Itoa(i))
		if exists {
			t.Errorf("Cleanup did not remove the expired item: key%d", i)
		}
	}
}

// TestConcurrency verifies that the cache handles concurrent access correctly by setting and getting multiple items in parallel.
// It ensures that all items are retrievable after concurrent operations.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestConcurrency(t *testing.T) {
	inMemoryCache := cache.NewInMemoryCache(expiration, cleanupInterval)
	defer inMemoryCache.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "key" + strconv.Itoa(i)
			value := "value" + strconv.Itoa(i)
			inMemoryCache.Set(key, value)
			_, found := inMemoryCache.Get(key)
			if !found {
				t.Errorf("Item %v was not found in cache", key)
			}
		}(i)
	}
	wg.Wait()

	// Verify all keys
	for i := 0; i < 100; i++ {
		key := "key" + strconv.Itoa(i)
		_, found := inMemoryCache.Get(key)
		if !found {
			t.Errorf("Item %v was not found in cache after concurrency test", key)
		}
	}
}
