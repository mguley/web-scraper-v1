package tor

import (
	"github.com/mguley/web-scraper-v1/internal/tor"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestEstablishConnection tests the ability of the Facade to establish a Tor connection.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestEstablishConnection(t *testing.T) {
	torFacade, createErr := tor.NewTorFacade(&cfg.TorProxy, poolSize)
	assert.NoError(t, createErr, "No error should have been returned")

	// Establish connection
	httpClient, err := torFacade.EstablishConnection()
	assert.NoError(t, err)
	assert.NotNil(t, httpClient)
}

// TestCheckConnection verifies the ability of the Facade to test a Tor connection by making a request to URL.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestCheckConnection(t *testing.T) {
	torFacade, createErr := tor.NewTorFacade(&cfg.TorProxy, poolSize)
	assert.NoError(t, createErr, "No error should have been returned")

	// Establish connection
	httpClient, err := torFacade.EstablishConnection()
	assert.NoError(t, err)
	assert.NotNil(t, httpClient)

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
	torFacade, createErr := tor.NewTorFacade(&cfg.TorProxy, poolSize)
	assert.NoError(t, createErr, "No error should have been returned")

	httpClient, err := torFacade.EstablishConnection()
	assert.NoError(t, err)
	assert.NotNil(t, httpClient)

	// Fetch the exit IP address
	ipAddress, fetchErr := torFacade.FetchExitIP(cfg.TorProxy.VerifyUrl)
	assert.NoError(t, fetchErr)
	assert.NotEmpty(t, ipAddress)
}

// TestChangeIdentitySuccess tests the successful change of Tor identity.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
//func TestChangeIdentitySuccess(t *testing.T) {
//	torFacade, createErr := tor.NewTorFacade(&cfg.TorProxy, poolSize)
//	assert.NoError(t, createErr, "No error should have been returned")

//	httpClient, err := torFacade.EstablishConnection()
//	assert.NoError(t, err)
//	assert.NotNil(t, httpClient)

// Change identity
//	changeErr := torFacade.ChangeIdentity()
//	assert.NoError(t, changeErr)
//}

// TestChangeIdentityFailure tests the failure scenario of changing Tor's identity due to an invalid password.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestChangeIdentityFailure(t *testing.T) {
	invalidConfig := cfg.TorProxy
	invalidConfig.ControlPassword = "invalid_password"
	torFacade, createErr := tor.NewTorFacade(&invalidConfig, poolSize)
	assert.NoError(t, createErr, "No error should have been returned")

	// Establish connection
	httpClient, err := torFacade.EstablishConnection()
	assert.NoError(t, err)
	assert.NotNil(t, httpClient)

	// Change identity should fail
	identityErr := torFacade.ChangeIdentity()
	assert.Error(t, identityErr, "Identity change failed due to invalid credentials")
}

// TestChangeIdentityExitIPChange tests that the exit IP address changes after changing Tor identity.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
//func TestChangeIdentityExitIPChange(t *testing.T) {
//	torFacade, createErr := tor.NewTorFacade(&cfg.TorProxy, poolSize)
//	assert.NoError(t, createErr, "No error should have been returned")

// Establish connection
//	httpClient, err := torFacade.EstablishConnection()
//	assert.NoError(t, err)
//	assert.NotNil(t, httpClient)

// Fetch initial exit IP address
//	initialIP, fetchErr := torFacade.FetchExitIP(cfg.TorProxy.VerifyUrl)
//	assert.NoError(t, fetchErr)
//	assert.NotEmpty(t, initialIP)

// Change identity
//	changeErr := torFacade.ChangeIdentity()
//	assert.NoError(t, changeErr)

// Fetch new exit IP address
//	newIP, fetchErr := torFacade.FetchExitIP(cfg.TorProxy.VerifyUrl)
//	assert.NoError(t, fetchErr)
//	assert.NotEmpty(t, newIP)

//	fmt.Println("IP: " + initialIP + ", newIP: " + newIP)
// Ensure the IP address has changed
//	assert.NotEqual(t, initialIP, newIP)
//}
