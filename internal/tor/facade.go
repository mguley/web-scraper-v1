package tor

import (
	"fmt"
	"github.com/mguley/web-scraper-v1/config"
	"net/http"
	"sync"
	"time"
)

// Facade provides a simplified interface for interacting with the Tor network.
// It encapsulates various functionalities such as establishing a Tor connection, testing the connection,
// and fetching the current exit IP address.
type Facade struct {
	proxyConfig *config.TorProxyConfig // proxyConfig holds the Tor proxy configuration.
	httpClient  *http.Client           // httpClient is the HTTP client configured to use the Tor network.
	once        sync.Once              // once ensures the HTTP client is initialized only once.
}

// NewTorFacade creates a new instance of the Facade with the provided proxy configuration.
//
// Parameters:
// - proxyConfig: The TorProxyConfig struct containing the configuration for the Tor proxy.
//
// Returns:
// - *Facade: A pointer to the newly created Facade instance.
func NewTorFacade(proxyConfig *config.TorProxyConfig) *Facade {
	return &Facade{
		proxyConfig: proxyConfig,
	}
}

// initializeClient initializes the HTTP client if it hasn't been initialized yet.
//
// Parameters:
// - timeout: A time.Duration representing the timeout for HTTP requests.
//
// Returns:
// - error: An error if the client could not be initialized.
func (torFacade *Facade) initializeClient(timeout time.Duration) error {
	var initErr error
	torFacade.once.Do(func() {
		client := NewTorClient()
		httpClient, err := client.CreateHTTPClient(torFacade.proxyConfig.Host, torFacade.proxyConfig.Port, timeout)
		if err != nil {
			initErr = fmt.Errorf("could not establish connection to Tor: %w", err)
			return
		}
		torFacade.httpClient = httpClient
	})
	return initErr
}

// EstablishConnection establishes a Tor-enabled HTTP client connection using the provided timeout duration.
//
// Parameters:
// - timeout: A time.Duration representing the timeout for HTTP requests.
//
// Returns:
// - *http.Client: The HTTP client configured to use the Tor network.
// - error: An error if the connection could not be established.
func (torFacade *Facade) EstablishConnection(timeout time.Duration) (*http.Client, error) {
	if err := torFacade.initializeClient(timeout); err != nil {
		return nil, err
	}
	return torFacade.httpClient, nil
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
	if torFacade.httpClient == nil {
		return "", fmt.Errorf("HTTP client not initialized")
	}
	client := NewTorClient()
	return client.TestConnection(torFacade.httpClient, url)
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
	return ChangeIdentity(torFacade.proxyConfig, torFacade.proxyConfig.ControlPassword)
}
