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
