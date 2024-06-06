package tor

import (
	"fmt"
	"github.com/mguley/web-scraper-v1/internal/tor"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// TestEstablishConnection tests the ability of the Facade to establish a Tor connection.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestEstablishConnection(t *testing.T) {
	torFacade := tor.NewTorFacade(&cfg.TorProxy)

	// Establish connection
	httpClient, err := torFacade.EstablishConnection(timeout)
	assert.NoError(t, err)
	assert.NotNil(t, httpClient)
}

// TestCheckConnection verifies the ability of the Facade to test a Tor connection by making a request to URL.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestCheckConnection(t *testing.T) {
	torFacade := tor.NewTorFacade(&cfg.TorProxy)

	// Establish connection
	_, err := torFacade.EstablishConnection(timeout)
	assert.NoError(t, err)

	// Test the connection
	response, responseErr := torFacade.TestConnection(cfg.TorProxy.PingUrl)
	assert.NoError(t, responseErr)
	assert.Contains(t, response, "Congratulations. This browser is configured to use Tor.")
}

// TestFetchExitIP verifies the ability of the Facade to fetch the current public IP address using the Tor network.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestFetchExitIP(t *testing.T) {
	torFacade := tor.NewTorFacade(&cfg.TorProxy)

	_, err := torFacade.EstablishConnection(timeout)
	assert.NoError(t, err)

	// Fetch the exit IP address
	ipAddress, fetchErr := torFacade.FetchExitIP(cfg.TorProxy.VerifyUrl)
	assert.NoError(t, fetchErr)
	assert.NotEmpty(t, ipAddress)
}

// TestChangeIdentityWrapperSuccess tests the successful change of Tor identity.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestChangeIdentityWrapperSuccess(t *testing.T) {
	torFacade := tor.NewTorFacade(&cfg.TorProxy)

	_, err := torFacade.EstablishConnection(timeout)
	assert.NoError(t, err)

	// Change identity
	changeErr := torFacade.ChangeIdentity()
	// Wait a bit for the identity to change, Tor network latency
	time.Sleep(10 * time.Second)

	assert.NoError(t, changeErr)
}

// TestChangeIdentityWrapperFailure tests the failure scenario of changing Tor's identity due to an invalid password.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestChangeIdentityWrapperFailure(t *testing.T) {
	invalidConfig := cfg.TorProxy
	invalidConfig.ControlPassword = "invalid_password"
	torFacade := tor.NewTorFacade(&invalidConfig)

	// Establish connection
	_, err := torFacade.EstablishConnection(timeout)
	assert.NoError(t, err)

	// Change identity should fail
	identityErr := torFacade.ChangeIdentity()
	// Wait a bit for the identity to change, Tor network latency
	time.Sleep(10 * time.Second)

	assert.Error(t, identityErr, "Identity change failed due to invalid credentials")
}

// TestChangeIdentityWrapperExitIPChange tests that the exit IP address changes after changing Tor identity.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestChangeIdentityWrapperExitIPChange(t *testing.T) {
	torFacade := tor.NewTorFacade(&cfg.TorProxy)

	// Establish connection
	_, err := torFacade.EstablishConnection(timeout)
	assert.NoError(t, err)

	// Fetch initial exit IP address
	initialIP, fetchErr := torFacade.FetchExitIP(cfg.TorProxy.VerifyUrl)
	assert.NoError(t, fetchErr)
	assert.NotEmpty(t, initialIP)

	// Change identity
	changeErr := torFacade.ChangeIdentity()
	// Wait a bit for the identity to change, Tor network latency
	time.Sleep(10 * time.Second)

	assert.NoError(t, changeErr)

	// Fetch new exit IP address
	newIP, fetchErr := torFacade.FetchExitIP(cfg.TorProxy.VerifyUrl)
	assert.NoError(t, fetchErr)
	assert.NotEmpty(t, newIP)

	fmt.Println("IP: " + initialIP + ", newIP: " + newIP)
	// Ensure the IP address has changed
	assert.NotEqual(t, initialIP, newIP)
}
