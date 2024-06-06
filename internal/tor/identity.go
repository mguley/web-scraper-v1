package tor

import (
	"fmt"
	"github.com/mguley/web-scraper-v1/config"
	"github.com/mguley/web-scraper-v1/internal/tor/builders"
	"github.com/mguley/web-scraper-v1/internal/tor/commands"
	"net/http"
	"time"
)

const (
	attempts     = 5
	sleepTimeout = 10 * time.Second
	signal       = "NEWNYM"
)

// ChangeIdentity changes the Tor identity by connecting to the Tor control port and sending the 'NEWNYM' signal.
// It attempts to ensure that the exit IP address is changed.
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

	for attempt := 0; attempt < attempts; attempt++ {
		if signalCmdErr := sendSignalCmd.Execute(); signalCmdErr != nil {
			return fmt.Errorf("failed to send NEWNYM signal: %w", signalCmdErr)
		}
		time.Sleep(sleepTimeout) // Wait for the new identity to take effect

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
