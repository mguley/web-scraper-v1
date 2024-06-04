package tor

import (
	"github.com/mguley/web-scraper-v1/internal/tor"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"testing"
)

// setupTorClient sets up a Tor-enabled HTTP client using configuration from an environment file.
// If the configuration initialization fails or if there are errors during the creation of the HTTP client,
// this function will return an error along with an appropriate error message.
//
// Returns:
// - *http.Client: An HTTP client configured to use the Tor SOCKS5 proxy with the provided host, port, and timeout.
// - error: An error if there was an issue initializing the configuration or creating the HTTP client.
func setupTorClient() (*http.Client, error) {
	httpClient, createErr := tor.NewTorClient().CreateHTTPClient(cfg.TorProxy.Host, cfg.TorProxy.Port, timeout)
	if createErr != nil {
		log.Printf("Failed to create tor http client: %v", createErr)
		return nil, createErr
	}

	return httpClient, nil
}

// TestCreateHTTPClientIntegration verifies the integration of creating an HTTP client with a Tor proxy server.
// It initializes a Tor client using configuration from an environment file and attempts to create an HTTP client
// configured to use the Tor SOCKS5 proxy. It verifies that the HTTP client is created successfully without any errors,
// ensuring that the Tor proxy settings are correctly configured.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestCreateHTTPClientIntegration(t *testing.T) {
	httpClient, err := setupTorClient()
	assert.NoError(t, err)
	assert.NotNil(t, httpClient)

	// Test Connection
	torClient := tor.NewTorClient()
	responseBody, responseErr := torClient.TestConnection(httpClient, cfg.TorProxy.PingUrl)
	assert.NoError(t, responseErr)
	assert.Contains(t, responseBody, "Congratulations. This browser is configured to use Tor.")
}

// TestInvalidProxy verifies the behavior when an invalid Tor proxy host or port is provided.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestInvalidProxy(t *testing.T) {
	httpClient, err := tor.NewTorClient().CreateHTTPClient("invalid_host", "invalid_port", timeout)
	assert.Error(t, err)
	assert.Nil(t, httpClient)
}

// TestNonTorURL verifies the connection to a non-Tor URL through the Tor network.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestNonTorURL(t *testing.T) {
	httpClient, err := setupTorClient()
	assert.NoError(t, err)
	assert.NotNil(t, httpClient)

	// Test Connection to a non-Tor URL
	torClient := tor.NewTorClient()
	responseBody, responseErr := torClient.TestConnection(httpClient, "http://example.com")
	assert.NoError(t, responseErr)
	assert.Contains(t, responseBody, "Example Domain")
}
