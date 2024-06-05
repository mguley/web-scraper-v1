package tor

import (
	"github.com/mguley/web-scraper-v1/internal/tor"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestChangeIdentitySuccess tests the ChangeIdentity function for a successful identity change.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestChangeIdentitySuccess(t *testing.T) {
	// Attempt to change identity
	err := tor.ChangeIdentity(&cfg.TorProxy, cfg.TorProxy.ControlPassword)
	assert.NoError(t, err)
}

// TestChangeIdentityFailAuth tests the ChangeIdentity function for a failed identity change due to incorrect password.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestChangeIdentityFailAuth(t *testing.T) {
	incorrectPassword := "incorrect_password"
	err := tor.ChangeIdentity(&cfg.TorProxy, incorrectPassword)
	assert.Error(t, err, "Expected an error due to incorrect password")
}
