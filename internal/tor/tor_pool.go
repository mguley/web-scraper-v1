package tor

import (
	"fmt"
	"github.com/mguley/web-scraper-v1/config"
	"net/http"
	"sync"
	"time"
)

const (
	timeout        = time.Second * 30
	recycleTimeout = time.Minute * 5
)

// Connection represents a single Tor connection with its associated HTTP client.
type Connection struct {
	proxyConfig *config.TorProxyConfig // proxyConfig holds the Tor proxy configuration.
	HttpClient  *http.Client           // HttpClient is the HTTP client configured to use the Tor network.
}

// Pool manages a pool of Tor connections to provide efficient usage and recycling of connections.
type Pool struct {
	Connections []*Connection  // connections is a slice of TorConnection pointers.
	mutex       sync.Mutex     // mutex is used to synchronize access to the pool.
	cond        *sync.Cond     // cond is a condition variable for managing connection borrowing and returning.
	shutdown    chan struct{}  // shutdown is a channel to signal pool shutdown.
	waitGroup   sync.WaitGroup // waitGroup is used to wait for all connections to close during shutdown.
}

// NewTorPool creates a new pool of Tor connections.
//
// Parameters:
// - proxyConfig: The *config.TorProxyConfig struct containing the configuration for the Tor proxy.
// - poolSize: An integer representing the number of connections to maintain in the pool.
//
// Returns:
// - *TorPool: A pointer to the newly created TorPool instance.
// - error: An error if the pool could not be initialized.
func NewTorPool(proxyConfig *config.TorProxyConfig, poolSize int) (*Pool, error) {
	pool := &Pool{
		Connections: make([]*Connection, poolSize),
		shutdown:    make(chan struct{}),
	}
	pool.cond = sync.NewCond(&pool.mutex)
	for i := 0; i < poolSize; i++ {
		conn, connErr := createTorConnection(proxyConfig)
		if connErr != nil {
			return nil, fmt.Errorf("failed to create Tor connection: %w", connErr)
		}
		pool.Connections[i] = conn
	}

	pool.waitGroup.Add(1)
	go pool.recycleConnections()

	return pool, nil
}

// createTorConnection creates a new Tor connection.
//
// Parameters:
// - proxyConfig: The *config.TorProxyConfig struct containing the configuration for the Tor proxy.
//
// Returns:
// - *TorConnection: A pointer to the newly created TorConnection instance.
// - error: An error if the connection could not be established.
func createTorConnection(proxyConfig *config.TorProxyConfig) (*Connection, error) {
	conn := &Connection{
		proxyConfig: proxyConfig,
	}
	client, clientErr := NewTorClient().CreateHTTPClient(proxyConfig.Host, proxyConfig.Port, timeout)
	if clientErr != nil {
		return nil, fmt.Errorf("could not establish connection to Tor: %w", clientErr)
	}
	conn.HttpClient = client
	return conn, nil
}

// Borrow retrieves a connection from the pool.
//
// Returns:
// - *TorConnection: A pointer to the borrowed TorConnection instance.
// - error: An error if no connection is available.
func (pool *Pool) Borrow() (*Connection, error) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if len(pool.Connections) == 0 {
		pool.cond.Wait()
	}

	conn := pool.Connections[0]
	pool.Connections = pool.Connections[1:]
	return conn, nil
}

// Return adds a connection back to the pool.
//
// Parameters:
// - conn: A pointer to the TorConnection instance to be returned.
func (pool *Pool) Return(conn *Connection) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	pool.Connections = append(pool.Connections, conn)
	pool.cond.Signal()
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
		case <-time.After(recycleTimeout):
			pool.mutex.Lock()
			for i := range pool.Connections {
				oldConn := pool.Connections[i]
				newConn, createErr := createTorConnection(oldConn.proxyConfig)
				if createErr != nil {
					fmt.Printf("Failed to recycle Tor connection: %s\n", createErr)
					continue
				}
				pool.Connections[i] = newConn
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
	for i, conn := range pool.Connections {
		fmt.Printf("Connection %d: %v\n", i, conn.HttpClient)
	}
}

// Shutdown gracefully shuts down the pool by closing all existing connections.
// It signals the recycling goroutine to stop, waits for it to finish, and then closes all HTTP connections in the pool
// to clean up resources.
func (pool *Pool) Shutdown() {
	close(pool.shutdown)
	pool.waitGroup.Wait()

	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	for _, conn := range pool.Connections {
		conn.HttpClient.CloseIdleConnections()
	}
	pool.Connections = nil
}
