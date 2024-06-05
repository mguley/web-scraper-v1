package builders

import (
	"bufio"
	"fmt"
	"github.com/mguley/web-scraper-v1/config"
	"net"
	"time"
)

// ControlConnectionBuilder is a builder for creating and configuring a connection to the Tor control port.
type ControlConnectionBuilder struct {
	Address         string        // Address is the address of the Tor control port.
	Timeout         time.Duration // Timeout is the duration to wait before timing out the connection.
	ControlPassword string        // ControlPassword is the password for authenticating with the Tor control port.
	Conn            net.Conn      // Conn is the net.Conn instance for the connection to the control port.
	Reader          *bufio.Reader // Reader is a buffered reader for reading responses from the control port.
}

// NewControlConnectionBuilder creates a new instance of ControlConnectionBuilder.
//
// Parameters:
// - proxyConfig: A config.TorProxyConfig struct containing the control port configuration.
// - controlPassword: A string representing the password for the Tor control port.
//
// Returns:
// - *ControlConnectionBuilder: A pointer to the newly created ControlConnectionBuilder instance.
func NewControlConnectionBuilder(proxyConfig *config.TorProxyConfig, controlPassword string) *ControlConnectionBuilder {
	address := fmt.Sprintf("%s:%s", proxyConfig.Host, proxyConfig.ControlPort)
	return &ControlConnectionBuilder{
		Address:         address,
		Timeout:         time.Second * 10,
		ControlPassword: controlPassword,
	}
}

// Connect establishes the connection to the Tor control port.
//
// Returns:
// - error: An error if the connection could not be established.
func (builder *ControlConnectionBuilder) Connect() error {
	if builder.Conn != nil {
		fmt.Println("Already connected to control port: ", builder.Address)
		return nil
	}

	conn, err := net.DialTimeout("tcp", builder.Address, builder.Timeout)
	if err != nil {
		return fmt.Errorf("error connecting to control port %s: %s", builder.Address, err)
	}
	builder.Conn = conn
	builder.Reader = bufio.NewReader(conn)
	fmt.Println("Successfully connected to control port: ", builder.Address)
	return nil
}

// Close closes the connection to the Tor control port.
//
// Returns:
// - error: An error if the connection could not be closed.
func (builder *ControlConnectionBuilder) Close() error {
	if builder.Conn != nil {
		return builder.Conn.Close()
	}
	return nil
}
