package tor

import (
	"fmt"
	"github.com/mguley/web-scraper-v1/config"
	"github.com/mguley/web-scraper-v1/internal/tor/builders"
	"github.com/mguley/web-scraper-v1/internal/tor/commands"
)

// ChangeIdentity changes the Tor identity by connecting to the Tor control port and sending the NEWNYM signal.
//
// Parameters:
// - proxyConfig: A config.TorProxyConfig struct containing the control port configuration
// - controlPassword: A string representing the password for the Tor control port.
//
// Returns:
// - error: An error if the identity could not be changed.
func ChangeIdentity(proxyConfig *config.TorProxyConfig, controlPassword string) error {
	builder := builders.NewControlConnectionBuilder(proxyConfig, controlPassword)
	authenticateCmd := &commands.AuthenticateCommand{Builder: builder}
	if authErr := authenticateCmd.Execute(); authErr != nil {
		return authErr
	}
	defer closeConnection(builder)

	sendSignalCmd := &commands.SendSignalCommand{Builder: builder, Signal: "NEWNYM"}
	return sendSignalCmd.Execute()
}

// closeConnection handles the closing of the control connection.
func closeConnection(builder *builders.ControlConnectionBuilder) {
	if err := builder.Close(); err != nil {
		fmt.Printf("Failed to close control connection: %s", err)
	}
}
