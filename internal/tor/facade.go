package tor

import (
	"fmt"
	"github.com/mguley/web-scraper-v1/config"
	"net/http"
	"sync"
)

// Facade provides a simplified interface for interacting with the Tor network.
// It encapsulates various functionalities such as establishing a Tor connection, testing the connection,
// and fetching the current exit IP address.
type Facade struct {
	proxyConfig *config.TorProxyConfig // proxyConfig holds the Tor proxy configuration.
	once        sync.Once              // once ensures the HTTP client is initialized only once.
	torPool     *Pool                  // torPool manages a pool of Tor connections.
}

// NewTorFacade creates a new instance of the Facade with the provided proxy configuration and pool size.
//
// Parameters:
// - proxyConfig: The *config.TorProxyConfig struct containing the configuration for the Tor proxy.
// - poolSize: An integer representing the number of connections to maintain in the pool.
//
// Returns:
// - *Facade: A pointer to the newly created Facade instance.
// - error: An error if the pool could not be initialized.
func NewTorFacade(proxyConfig *config.TorProxyConfig, poolSize int) (*Facade, error) {
	torPool, err := NewTorPool(proxyConfig, poolSize)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Tor pool: %w", err)
	}

	return &Facade{
		proxyConfig: proxyConfig,
		torPool:     torPool,
	}, nil
}

// borrowConn borrows an HTTP client from the Tor pool.
//
// Returns:
// - *pool.TorConnection: The borrowed TorConnection to use the Tor network.
// - error: An error if no connection is available.
func (torFacade *Facade) borrowConn() (*Connection, error) {
	conn, err := torFacade.torPool.Borrow()
	if err != nil {
		return nil, fmt.Errorf("failed to borrow Tor connection: %w", err)
	}
	return conn, nil
}

// returnConn returns an HTTP client to the Tor pool
//
// Parameters:
// - conn: The TorConnection to be returned to the pool.
func (torFacade *Facade) returnConn(conn *Connection) {
	torFacade.torPool.Return(conn)
}

// EstablishConnection establishes a Tor-enabled HTTP client connection using the provided timeout duration.
//
// Returns:
// - *http.Client: The HTTP client configured to use the Tor network.
// - error: An error if the connection could not be established.
func (torFacade *Facade) EstablishConnection() (*http.Client, error) {
	conn, err := torFacade.borrowConn()
	if err != nil {
		return nil, fmt.Errorf("failed to establish Tor connection: %w", err)
	}
	return conn.HttpClient, nil
}

// TestConnection tests if the Tor-enabled HTTP client can successfully make a request to a given URL.
//
// Parameters:
// - url: A string representing the URL to test the client against.
//
// Returns:
// - string: The response body from the test URL.
// - error: An error if the request failed or the response could not be read.
func (torFacade *Facade) TestConnection(url string) (string, error) {
	conn, err := torFacade.borrowConn()
	if err != nil {
		return "", fmt.Errorf("failed to establish Tor connection: %w", err)
	}
	defer torFacade.returnConn(conn)

	client := NewTorClient()
	return client.TestConnection(conn.HttpClient, url)
}

// FetchExitIP fetches the current public IP address using the Tor-enabled HTTP client.
//
// Parameters:
// - url: A string representing the URL to fetch the IP address from.
//
// Returns:
// - string: The current public IP address.
// - error: An error if the IP address could not be fetched.
func (torFacade *Facade) FetchExitIP(url string) (string, error) {
	return torFacade.TestConnection(url)
}

// ChangeIdentity changes the Tor identity.
//
// Returns:
// - error: An error if the identity could not be changed.
func (torFacade *Facade) ChangeIdentity() error {
	return ChangeIdentity(torFacade.torPool, torFacade.proxyConfig)
}
