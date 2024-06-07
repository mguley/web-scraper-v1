package tor

import (
	"fmt"
	"github.com/mguley/web-scraper-v1/config"
	"github.com/mguley/web-scraper-v1/internal/tor/builders"
	"github.com/mguley/web-scraper-v1/internal/tor/commands"
	"math"
	"net/http"
	"time"
)

const (
	maxAttempts     = 5
	initialWaitTime = 5 * time.Second
	signal          = "NEWNYM"
)

// ChangeIdentity changes the Tor identity by connecting to the Tor control port and sending the 'NEWNYM' signal.
// It attempts to ensure that the exit IP address is changed by using an exponential backoff strategy.
//
// This function performs the following steps:
// 1. Borrow a Tor connection from the pool.
// 2. Fetch the initial exit IP address to compare later.
// 3. Set up a control connection to the Tor control port and authenticate.
// 4. Send the 'NEWNYM' signal to request a new identity.
// 5. Wait for a dynamically increasing period before checking the new IP address, using exponential backoff.
// 6. Fetch the new exit IP address and compare it with the initial IP.
// 7. Retry up to a maximum number of attempts if the IP address has not changed, doubling the wait time after each attempt.
//
// Parameters:
// - torPool: A pointer to the TorPool instance.
// - proxyConfig: A config.TorProxyConfig struct containing the control port configuration.
//
// Returns:
// - error: An error if the identity could not be changed.
func ChangeIdentity(torPool *Pool, proxyConfig *config.TorProxyConfig) error {
	// Borrow a connection from the pool
	conn, borrowErr := torPool.Borrow()
	if borrowErr != nil {
		return fmt.Errorf("failed to borrow Tor connection: %w", borrowErr)
	}
	defer torPool.Return(conn)

	// Fetch the initial exit IP address
	initialIp, fetchErr := fetchExitIP(conn.HttpClient, proxyConfig.VerifyUrl)
	if fetchErr != nil {
		return fmt.Errorf("failed to fetch initial exit IP address: %w", fetchErr)
	}

	// Set up the control connection
	builder := builders.NewControlConnectionBuilder(proxyConfig, proxyConfig.ControlPassword)
	authenticateCmd := &commands.AuthenticateCommand{Builder: builder}
	if authErr := authenticateCmd.Execute(); authErr != nil {
		return authErr
	}
	defer closeConnection(builder)

	sendSignalCmd := &commands.SendSignalCommand{Builder: builder, Signal: signal}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if signalCmdErr := sendSignalCmd.Execute(); signalCmdErr != nil {
			return fmt.Errorf("failed to send NEWNYM signal: %w", signalCmdErr)
		}
		waitTime := time.Duration(math.Pow(2, float64(attempt))) * initialWaitTime
		time.Sleep(waitTime) // Exponential backoff

		// Fetch the new exit IP address
		newIP, err := fetchExitIP(conn.HttpClient, proxyConfig.VerifyUrl)
		if err != nil {
			return fmt.Errorf("failed to fetch new exit IP address: %w", err)
		}

		// Check if the new IP is different
		if newIP != initialIp {
			fmt.Printf("Successfully changed IP from %s to %s\n", initialIp, newIP)
			return nil // Successfully changed the identity
		}
		fmt.Printf("Attempt %d: IP address did not change. Retrying...\n", attempt+1)
	}

	return fmt.Errorf("could not obtain a different exit IP address after 5 attempts")
}

// fetchExitIP fetches the current public IP address using the provided HTTP client.
// It sends a request to the specified URL using the provided HTTP client and returns the IP address found in the response.
//
// Parameters:
// - httpClient: The *http.Client to use for making the request.
// - url: A string representing the URL to fetch the IP address from.
//
// Returns:
// - string: The current public IP address.
// - error: An error if the IP address could not be fetched.
func fetchExitIP(httpClient *http.Client, url string) (string, error) {
	client := NewTorClient()
	return client.TestConnection(httpClient, url)
}

// closeConnection handles the closing of the control connection.
//
// Parameters:
// - builder: The ControlConnectionBuilder instance to close.
func closeConnection(builder *builders.ControlConnectionBuilder) {
	if err := builder.Close(); err != nil {
		fmt.Printf("Failed to close control connection: %s", err)
	}
}
