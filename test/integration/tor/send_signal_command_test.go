package tor

import (
	"github.com/mguley/web-scraper-v1/internal/tor/builders"
	"github.com/mguley/web-scraper-v1/internal/tor/commands"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestSendSignalCommandFailAuth tests the SendSignalCommand for failure due to missed authentication.
// It sets up a ControlConnectionBuilder with the correct control password and executes the SendSignalCommand without authenticating first.
// It asserts that an error is returned, indicating that the command failed due to missed authentication.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestSendSignalCommandFailAuth(t *testing.T) {
	builder := builders.NewControlConnectionBuilder(&cfg.TorProxy, cfg.TorProxy.ControlPassword)
	cmd := &commands.SendSignalCommand{Builder: builder, Signal: "NEWNYM"}

	execErr := cmd.Execute()
	assert.Error(t, execErr, "Expected an error due to missed authentication")
}

// TestSendSignalCommandSuccess tests the SendSignalCommand for successful signal sending after authentication.
// It sets up a ControlConnectionBuilder with the correct control password, authenticates with the Tor control port,
// and then executes the SendSignalCommand with a valid signal.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestSendSignalCommandSuccess(t *testing.T) {
	builder := builders.NewControlConnectionBuilder(&cfg.TorProxy, cfg.TorProxy.ControlPassword)
	authCmd := &commands.AuthenticateCommand{Builder: builder}

	// Authenticate first
	authExecErr := authCmd.Execute()
	assert.NoError(t, authExecErr, "Expected no error")

	// Send signal command
	signalCmd := &commands.SendSignalCommand{Builder: builder, Signal: "NEWNYM"}
	signalExecErr := signalCmd.Execute()
	assert.NoError(t, signalExecErr, "Expected no error")
}
