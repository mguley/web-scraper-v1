package tor

import (
	"github.com/mguley/web-scraper-v1/internal/tor"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// TestNewTorPool verifies the creation of a new Tor connection pool.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestNewTorPool(t *testing.T) {
	pool, err := tor.NewTorPool(&cfg.TorProxy, poolSize, poolTimeout, poolRecycleTimeout)
	assert.NoError(t, err, "Failed to initialize Tor pool")
	assert.Equal(t, poolSize, len(pool.Connections), "Expected pool size")
}

// TestBorrowReturn verifies borrowing and returning a connection from the pool.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestBorrowReturn(t *testing.T) {
	pool, err := tor.NewTorPool(&cfg.TorProxy, poolSize, poolTimeout, poolRecycleTimeout)
	assert.NoError(t, err, "Failed to initialize Tor pool")

	connection, borrowErr := pool.Borrow()
	assert.NoError(t, borrowErr, "Failed to borrow connection")
	assert.NotNil(t, connection, "Borrowed connection is nil")

	pool.Return(connection)

	// Ensure connection is back in the pool
	_, exists := pool.Connections[connection]
	assert.True(t, exists, "Returned connection not found in pool")
}

// TestShutdown verifies the shutdown process of the pool.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestShutdown(t *testing.T) {
	pool, err := tor.NewTorPool(&cfg.TorProxy, poolSize, poolTimeout, poolRecycleTimeout)
	assert.NoError(t, err, "Failed to initialize Tor pool")
	pool.Shutdown()

	// Ensure all connections are closed
	assert.Nil(t, pool.Connections, "Pool connections not set to nil after shutdown")

	connection, borrowErr := pool.Borrow()
	assert.Error(t, borrowErr, "Expected error when borrowing from a shut down pool")
	assert.Nil(t, connection, "Expected nil connection when borrowing from a shut down pool")
}

// TestRecycleConnections verifies the recycling of connections in the pool.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestRecycleConnections(t *testing.T) {
	pool, err := tor.NewTorPool(&cfg.TorProxy, poolSize, poolTimeout, time.Second*3)
	assert.NoError(t, err, "Failed to initialize Tor pool")

	initialConnections := make(map[*tor.Connection]struct{})
	for connection := range pool.Connections {
		initialConnections[connection] = struct{}{}
	}

	// Wait for the recycling process to happen
	time.Sleep(time.Second * 5)

	// Check that at least one connection has been recycled
	recycled := false
	for conn := range pool.Connections {
		if _, exists := initialConnections[conn]; !exists {
			recycled = true
			break
		}
	}
	assert.True(t, recycled, "Expected at least one connection to be recycled")
}
