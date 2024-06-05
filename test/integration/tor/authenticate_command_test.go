package tor

import (
	"github.com/mguley/web-scraper-v1/internal/tor/builders"
	"github.com/mguley/web-scraper-v1/internal/tor/commands"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestAuthenticateCommandSuccess tests the AuthenticateCommand for a successful authentication.
// It sets up a ControlConnectionBuilder with the correct control password and asserts that no error is returned,
// indicating successful authentication.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestAuthenticateCommandSuccess(t *testing.T) {
	builder := builders.NewControlConnectionBuilder(&cfg.TorProxy, cfg.TorProxy.ControlPassword)
	cmd := &commands.AuthenticateCommand{Builder: builder}

	execErr := cmd.Execute()
	assert.NoError(t, execErr)
}

// TestAuthenticateCommandFailureEmptyPass tests the AuthenticateCommand for a failure due to an empty password.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestAuthenticateCommandFailureEmptyPass(t *testing.T) {
	builder := builders.NewControlConnectionBuilder(&cfg.TorProxy, "")
	cmd := &commands.AuthenticateCommand{Builder: builder}

	execErr := cmd.Execute()
	assert.Error(t, execErr)
}

// TestAuthenticateCommandFailureWrongPass tests the AuthenticateCommand for a failure due to an incorrect password.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestAuthenticateCommandFailureWrongPass(t *testing.T) {
	builder := builders.NewControlConnectionBuilder(&cfg.TorProxy, "wrong")
	cmd := &commands.AuthenticateCommand{Builder: builder}

	execErr := cmd.Execute()
	assert.Error(t, execErr)
}
