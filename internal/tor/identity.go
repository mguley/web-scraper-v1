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

// RetryStrategy defines the interface for a retry strategy.
type RetryStrategy interface {
	// WaitDuration returns the duration to wait before the next retry attempt.
	//
	// Parameters:
	// - attempt: The current retry attempt number.
	//
	// Returns:
	// - time.Duration: The duration to wait before the next retry.
	WaitDuration(attempt int) time.Duration
}

// ExponentialBackoff implements the RetryStrategy with exponential backoff
type ExponentialBackoff struct {
	initialWaitTime time.Duration // initialWaitTime is the base wait time for the first retry attempt.
}

// WaitDuration calculates the wait duration for the given retry attempt using exponential backoff.
func (exponentialBackoff *ExponentialBackoff) WaitDuration(attempt int) time.Duration {
	return time.Duration(math.Pow(2, float64(attempt))) * exponentialBackoff.initialWaitTime
}

// IdentityChanger is responsible for changing the Tor identity.
type IdentityChanger struct {
	torPool       *Pool                  // torPool manages a pool of Tor connections.
	proxyConfig   *config.TorProxyConfig // proxyConfig holds the Tor proxy configuration.
	retryStrategy RetryStrategy          // retryStrategy defines the strategy for retrying failed identity changes.
	maxAttempts   int                    // maxAttempts is the maximum number of retry attempts.
	signal        string                 // signal is the command signal to send for changing identity.
}

// NewIdentityChanger creates a new instance of IdentityChanger.
//
// Parameters:
// - torPool: A pointer to the TorPool instance.
// - proxyConfig: A config.TorProxyConfig struct containing the control port configuration.
// - retryStrategy: A RetryStrategy to define the wait duration between retries.
// - maxAttempts: The maximum number of retry attempts.
//
// Returns:
// - *IdentityChanger: A pointer to the newly created IdentityChanger instance.
func NewIdentityChanger(torPool *Pool, proxyConfig *config.TorProxyConfig, retryStrategy RetryStrategy,
	maxAttempts int) *IdentityChanger {
	return &IdentityChanger{
		torPool:       torPool,
		proxyConfig:   proxyConfig,
		retryStrategy: retryStrategy,
		maxAttempts:   maxAttempts,
		signal:        "NEWNYM",
	}
}

// ChangeIdentity changes the Tor identity by connecting to the Tor control port and sending the 'NEWNYM' signal.
// It uses the configured retry strategy to wait between attempts and retries up to the maximum number of attempts.
//
// Returns:
// - error: An error if the identity could not be changed after the maximum number of attempts.
func (identityChanger *IdentityChanger) ChangeIdentity() error {
	// Borrow a connection from the pool
	conn, borrowErr := identityChanger.torPool.Borrow()
	if borrowErr != nil {
		return fmt.Errorf("failed to borrow Tor connection: %w", borrowErr)
	}
	defer identityChanger.torPool.Return(conn)

	// Fetch the initial exit IP address
	initialIP, fetchErr := fetchExitIP(conn.HttpClient, identityChanger.proxyConfig.VerifyUrl)
	if fetchErr != nil {
		return fmt.Errorf("failed to fetch initial exit IP address: %w", fetchErr)
	}

	// Set up the control connection
	builder := builders.NewControlConnectionBuilder(identityChanger.proxyConfig, identityChanger.proxyConfig.ControlPassword)
	if authErr := authenticate(builder); authErr != nil {
		return fmt.Errorf("failed to authenticate Tor connection: %w", authErr)
	}
	defer closeControlConnection(builder)

	sendSignalCmd := &commands.SendSignalCommand{Builder: builder, Signal: identityChanger.signal}

	// Retry changing identity up to the maximum number of attempts
	for attempt := 0; attempt < identityChanger.maxAttempts; attempt++ {
		if signalCmdErr := sendSignalCmd.Execute(); signalCmdErr != nil {
			return fmt.Errorf("failed to send signal command: %w", signalCmdErr)
		}

		// Wait using the retry strategy
		time.Sleep(identityChanger.retryStrategy.WaitDuration(attempt))

		// Fetch the new exit IP address
		newIP, err := fetchExitIP(conn.HttpClient, identityChanger.proxyConfig.VerifyUrl)
		if err != nil {
			return fmt.Errorf("failed to fetch new exit IP address: %w", err)
		}

		if newIP != initialIP {
			fmt.Printf("Successfully changed IP from %s to %s\n", initialIP, newIP)
			return nil // Successfully changed the identity
		}

		fmt.Printf("Attempt %d: IP address did not change. Retrying...\n", attempt+1)
	}

	return fmt.Errorf("could not obtain a different exit IP address after %d attempts", identityChanger.maxAttempts)
}

// authenticate sends an authentication command to the Tor control port.
//
// Parameters:
// - builder: The ControlConnectionBuilder instance to use for the authentication.
//
// Returns:
// - error: An error if the authentication failed.
func authenticate(builder *builders.ControlConnectionBuilder) error {
	authenticateCmd := &commands.AuthenticateCommand{Builder: builder}
	if err := authenticateCmd.Execute(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	return nil
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
	return NewTorClient().TestConnection(httpClient, url)
}

// closeControlConnection handles the closing of the control connection.
//
// Parameters:
// - builder: The ControlConnectionBuilder instance to close.
func closeControlConnection(builder *builders.ControlConnectionBuilder) {
	if err := builder.Close(); err != nil {
		fmt.Printf("Failed to close control connection: %s", err)
	}
}
