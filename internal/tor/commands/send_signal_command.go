package commands

import (
	"errors"
	"fmt"
	"github.com/mguley/web-scraper-v1/internal/tor/builders"
	"strings"
)

// SendSignalCommand encapsulates the command for sending a signal to the Tor control port.
type SendSignalCommand struct {
	Builder *builders.ControlConnectionBuilder // Builder is the ControlConnectionBuilder used to manage the connection.
	Signal  string                             // Signal is the string representing the signal command to send (e.g., "NEWNYM").
}

// Execute sends a signal command to the Tor control port.
//
// Returns:
// - error: An error if the signal command could not be sent or if the response indicates failure.
func (cmd *SendSignalCommand) Execute() error {
	if connErr := cmd.Builder.Connect(); connErr != nil {
		return connErr
	}

	_, signalErr := fmt.Fprintf(cmd.Builder.Conn, "SIGNAL %s\r\n", cmd.Signal)
	if signalErr != nil {
		return fmt.Errorf("failed to send signal command: %w", signalErr)
	}

	signalResponse, err := cmd.Builder.Reader.ReadString('\n')
	return validateSignalResponse(signalResponse, err)
}

// validateSignalResponse validates the signal response from the Tor control port.
//
// Parameters:
// - signalResponse: The response string from the signal command.
// - signalErr: The error from reading the signal response.
//
// Returns:
// - error: An error if the response indicates failure or if there was an error reading the response.
func validateSignalResponse(signalResponse string, signalErr error) error {
	if signalErr != nil {
		return fmt.Errorf("failed to read signal response: %w", signalErr)
	}

	fmt.Printf("Signal response: %s", signalResponse)

	switch {
	case strings.HasPrefix(signalResponse, "250"):
		return nil
	case strings.HasPrefix(signalResponse, "514"):
		return errors.New("authentication required")
	default:
		return errors.New("unexpected signal response: " + signalResponse)
	}
}
