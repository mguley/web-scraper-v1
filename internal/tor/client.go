package tor

import (
	"net/http"
	"time"
)

// Client interface defines the methods for managing HTTP clients that route through Tor.
type Client interface {
	// CreateHTTPClient creates an HTTP client configured to use a Tor SOCKS5 proxy.
	//
	// Parameters:
	// - host: A string representing the hostname or IP address of the Tor proxy server.
	// - port: A string representing the port number of the Tor proxy server.
	// - timeout: A time.Duration representing the timeout for HTTP requests.
	//
	// Returns:
	// - *http.Client: An HTTP client configured to use the Tor SOCKS5 proxy with a context-aware dialer and the specified timeout.
	// - error: An error if there was an issue setting up the SOCKS5 dialer.
	CreateHTTPClient(host string, port string, timeout time.Duration) (*http.Client, error)

	// TestConnection tests if the Tor-enabled HTTP client can successfully make a request to the given URL.
	//
	// Parameters:
	// - client: The *http.Client to use for making the request.
	// - url: A string representing the URL to test the client against.
	//
	// Returns:
	// - string: The response body from the test URL.
	// - error: An error if there was an issue making the request or reading the response.
	TestConnection(client *http.Client, url string) (string, error)
}
