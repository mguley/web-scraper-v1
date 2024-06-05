package commands

import (
	"errors"
	"fmt"
	"github.com/mguley/web-scraper-v1/internal/tor/builders"
	"strings"
)

// AuthenticateCommand encapsulates the command for authenticating with the Tor control port.
type AuthenticateCommand struct {
	Builder *builders.ControlConnectionBuilder // Builder is the ControlConnectionBuilder used to manage the connection.
}

// Execute sends the authentication command to the Tor control port.
//
// Returns:
// - error: An error if the authentication command could not be sent or if the authentication failed.
func (cmd *AuthenticateCommand) Execute() error {
	if cmd.Builder.ControlPassword == "" {
		return errors.New("no control password provided")
	}

	if connErr := cmd.Builder.Connect(); connErr != nil {
		return connErr
	}
	_, authErr := fmt.Fprintf(cmd.Builder.Conn, "AUTHENTICATE \"%s\"\r\n", cmd.Builder.ControlPassword)
	if authErr != nil {
		return fmt.Errorf("failed to send authentication command: %w", authErr)
	}

	authResponse, err := cmd.Builder.Reader.ReadString('\n')
	return validateAuthResponse(authResponse, err)
}

// validateAuthResponse validates the authentication response from the Tor control port.
//
// Parameters:
// - authResponse: The response string from the authentication command.
// - authErr: The error from reading the authentication response.
//
// Returns:
// - error: An error if the authentication failed, the response is unexpected, or if there was an error reading the response.
func validateAuthResponse(authResponse string, authErr error) error {
	if authErr != nil {
		return fmt.Errorf("failed to send authentication command: %w", authErr)
	}

	fmt.Printf("Authentication response: %s\n", authResponse)

	switch {
	case strings.HasPrefix(authResponse, "250"):
		return nil
	case strings.HasPrefix(authResponse, "515"):
		return errors.New("authentication failed: incorrect password")
	default:
		return fmt.Errorf("unexpected authentication response: %s", authResponse)
	}
}
