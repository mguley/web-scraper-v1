package tor

import (
	"fmt"
	"github.com/mguley/web-scraper-v1/config"
	"net/http"
	"sync"
	"time"
)

// Connection represents a single Tor connection with its associated HTTP client.
type Connection struct {
	proxyConfig *config.TorProxyConfig // proxyConfig holds the Tor proxy configuration.
	HttpClient  *http.Client           // HttpClient is the HTTP client configured to use the Tor network.
}

// Pool manages a pool of Tor connections to provide efficient usage and recycling of connections.
type Pool struct {
	Connections    map[*Connection]struct{}
	mutex          sync.Mutex     // mutex is used to synchronize access to the pool.
	cond           *sync.Cond     // cond is a condition variable for managing connection borrowing and returning.
	shutdown       chan struct{}  // shutdown is a channel to signal pool shutdown.
	isShutdown     bool           // isShutdown indicates whether the pool has been shut down.
	waitGroup      sync.WaitGroup // waitGroup is used to wait for all connections to close during shutdown.
	timeout        time.Duration  // timeout is the duration for which an HTTP client waits for a response.
	recycleTimeout time.Duration  // recycleTimeout is the duration after which connections are recycled.
}

// NewTorPool creates a new pool of Tor connections.
//
// Parameters:
// - proxyConfig: The *config.TorProxyConfig struct containing the configuration for the Tor proxy.
// - poolSize: An integer representing the number of connections to maintain in the pool.
// - timeout: A duration representing the HTTP client timeout.
// - recycleTimeout: A duration representing the interval for recycling connections.
//
// Returns:
// - *Pool: A pointer to the newly created TorPool instance.
// - error: An error if the pool could not be initialized.
func NewTorPool(proxyConfig *config.TorProxyConfig, poolSize int, timeout time.Duration,
	recycleTimeout time.Duration) (*Pool, error) {
	pool := &Pool{
		Connections:    make(map[*Connection]struct{}, poolSize),
		shutdown:       make(chan struct{}),
		isShutdown:     false,
		timeout:        timeout,
		recycleTimeout: recycleTimeout,
	}
	pool.cond = sync.NewCond(&pool.mutex)

	for i := 0; i < poolSize; i++ {
		conn, connErr := createTorConnection(proxyConfig, timeout)
		if connErr != nil {
			return nil, fmt.Errorf("failed to create Tor connection: %w", connErr)
		}
		pool.Connections[conn] = struct{}{}
	}

	pool.waitGroup.Add(1)
	go pool.recycleConnections()

	return pool, nil
}

// createTorConnection creates a new Tor connection.
//
// Parameters:
// - proxyConfig: The *config.TorProxyConfig struct containing the configuration for the Tor proxy.
// - timeout: A duration representing the HTTP client timeout.
//
// Returns:
// - *TorConnection: A pointer to the newly created TorConnection instance.
// - error: An error if the connection could not be established.
func createTorConnection(proxyConfig *config.TorProxyConfig, timeout time.Duration) (*Connection, error) {
	client, clientErr := NewTorClient().CreateHTTPClient(proxyConfig.Host, proxyConfig.Port, timeout)
	if clientErr != nil {
		return nil, fmt.Errorf("could not establish connection to Tor: %w", clientErr)
	}

	return &Connection{
		proxyConfig: proxyConfig,
		HttpClient:  client,
	}, nil
}

// Borrow retrieves a connection from the pool.
//
// Returns:
// - *TorConnection: A pointer to the borrowed TorConnection instance.
// - error: An error if no connection is available.
func (pool *Pool) Borrow() (*Connection, error) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if pool.isShutdown {
		return nil, fmt.Errorf("pool is shut down")
	}

	if len(pool.Connections) == 0 {
		pool.cond.Wait()
	}

	for conn := range pool.Connections {
		delete(pool.Connections, conn)
		return conn, nil
	}

	return nil, fmt.Errorf("no connections available")
}

// Return adds a connection back to the pool.
//
// Parameters:
// - conn: A pointer to the TorConnection instance to be returned.
func (pool *Pool) Return(conn *Connection) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if pool.isShutdown {
		conn.HttpClient.CloseIdleConnections()
		return
	}

	pool.Connections[conn] = struct{}{}
	pool.cond.Signal()
}

// RefreshConnections refreshes all connections in the pool.
func (pool *Pool) RefreshConnections() {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	for conn := range pool.Connections {
		conn.HttpClient.CloseIdleConnections()
		delete(pool.Connections, conn)
		newConn, createErr := createTorConnection(conn.proxyConfig, pool.timeout)
		if createErr != nil {
			fmt.Printf("Failed to refresh Tor connection: %s\n", createErr)
			continue
		}
		pool.Connections[newConn] = struct{}{}
	}
}

// recycleConnections periodically refreshes the Tor connections in the pool.
// It runs in a separate goroutine and periodically creates new Tor connections to replace the existing ones,
// ensuring that the connections are fresh and functional.
func (pool *Pool) recycleConnections() {
	defer pool.waitGroup.Done()

	for {
		select {
		case <-pool.shutdown:
			return
		case <-time.After(pool.recycleTimeout):
			pool.mutex.Lock()
			for conn := range pool.Connections {
				newConn, createErr := createTorConnection(conn.proxyConfig, pool.timeout)
				if createErr != nil {
					fmt.Printf("Failed to recycle Tor connection: %s\n", createErr)
					continue
				}
				delete(pool.Connections, conn)
				pool.Connections[newConn] = struct{}{}
			}
			pool.mutex.Unlock()
		}
	}
}

// LogPoolMetrics logs the current state of the pool for monitoring purposes.
// It prints the number of connections currently in the pool and details about each connection,
// helping in monitoring and debugging the state of the connection pool.
func (pool *Pool) LogPoolMetrics() {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	fmt.Printf("Current pool size: %d\n", len(pool.Connections))
	for conn := range pool.Connections {
		fmt.Printf("Connection: %v\n", conn.HttpClient)
	}
}

// Shutdown gracefully shuts down the pool by closing all existing connections.
// It signals the recycling goroutine to stop, waits for it to finish, and then closes all HTTP connections in the pool
// to clean up resources.
func (pool *Pool) Shutdown() {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	pool.isShutdown = true
	close(pool.shutdown)
	pool.waitGroup.Wait()

	for conn := range pool.Connections {
		conn.HttpClient.CloseIdleConnections()
	}
	pool.Connections = nil
}
