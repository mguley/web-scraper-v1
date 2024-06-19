package tor

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/net/proxy"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

// TorsClient is a concrete implementation of the Client interface.
type TorsClient struct{}

// NewTorClient creates a new instance of TorsClient.
func NewTorClient() *TorsClient {
	return &TorsClient{}
}

// validateProxyDetails validates the host and port of the Tor proxy server.
//
// Parameters:
// - host: A string representing the hostname or IP address of the Tor proxy server.
// - port: A string representing the port number of the Tor proxy server.
//
// Returns:
// - error: An error if the host or port are invalid.
func validateProxyDetails(host string, port string) error {
	if host == "" || port == "" {
		return errors.New("proxy host or port is empty")
	}

	// Validate IP address
	//if net.ParseIP(host) == nil {
	//	return errors.New(fmt.Sprintf("invalid proxy host: %s", host))
	//}

	// Validate port number
	if _, err := strconv.Atoi(port); err != nil {
		return errors.New("invalid port number")
	}

	return nil
}

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
func (client *TorsClient) CreateHTTPClient(host string, port string, timeout time.Duration) (*http.Client, error) {
	if validationErr := validateProxyDetails(host, port); validationErr != nil {
		return nil, validationErr
	}

	proxyURI := fmt.Sprintf("%s:%s", host, port)
	dialer, err := proxy.SOCKS5("tcp", proxyURI, nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy: %s", err)
	}

	// Define a context-aware dialer function
	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.Dial(network, addr)
	}
	httpTransport := &http.Transport{
		DialContext: dialContext,
	}
	httpClient := &http.Client{
		Transport: httpTransport,
		Timeout:   timeout,
	}

	return httpClient, nil
}

// TestConnection tests if the Tor-enabled HTTP client can successfully make a request to the given URL.
//
// Parameters:
// - httpClient: The *http.Client to use for making the request.
// - url: A string representing the URL to test the client against.
//
// Returns:
// - string: The response body from the test URL.
// - error: An error if there was an issue making the request or reading the response.
func (client *TorsClient) TestConnection(httpClient *http.Client, url string) (string, error) {
	response, getErr := httpClient.Get(url)
	if getErr != nil {
		return "", fmt.Errorf("failed to connect to proxy: %s", getErr)
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			return
		}
	}()
	body, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return "", fmt.Errorf("failed to read response body: %s", readErr)
	}

	return string(body), nil
}
