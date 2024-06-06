package tor

import (
	"github.com/mguley/web-scraper-v1/internal/tor/builders"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestControlConnectionBuilderConnect tests the Connect method for successfully establishing a connection to the Tor control port.
// It sets up a ControlConnectionBuilder with the correct control port address and attempts to establish a connection.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestControlConnectionBuilderConnect(t *testing.T) {
	builder := builders.NewControlConnectionBuilder(&cfg.TorProxy, cfg.TorProxy.ControlPassword)
	connErr := builder.Connect()
	assert.NoError(t, connErr, "Expected no error when connecting to the control port")
}

// TestControlConnectionBuilderReconnect tests the Connect method for handling reconnection attempts.
// It sets up a ControlConnectionBuilder with the correct control port address, establishes a connection,
// and then attempts to reconnect. It asserts that no error is returned and verifies that the connection is handled properly.
func TestControlConnectionBuilderReconnect(t *testing.T) {
	builder := builders.NewControlConnectionBuilder(&cfg.TorProxy, cfg.TorProxy.ControlPassword)
	connErr := builder.Connect()
	assert.NoError(t, connErr, "Expected no error when connecting to the control port")

	// Attempt to reconnect
	reConnErr := builder.Connect()
	assert.NoError(t, reConnErr, "Expected no error when reconnecting to the control port")
}

// TestControlConnectionBuilderClose tests the Close method for successfully closing the connection to the Tor control port.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestControlConnectionBuilderClose(t *testing.T) {
	builder := builders.NewControlConnectionBuilder(&cfg.TorProxy, cfg.TorProxy.ControlPassword)
	connErr := builder.Connect()
	assert.NoError(t, connErr, "Expected no error when connecting to the control port")

	// Close the connection
	closeErr := builder.Close()
	assert.NoError(t, closeErr, "Expected no error when closing the connection to the control port")
}
