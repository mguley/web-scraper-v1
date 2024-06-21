package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// handleRequests processes each HTTP request by logging the User-Agent, IP address, and Forwarded Host,
// and responding with this information. This function acts as an HTTP handler.
//
// Parameters:
// - writer http.ResponseWriter: The writer to send data to the client.
// - request *http.Request: The HTTP request object containing all details of the request.
func handleRequests(writer http.ResponseWriter, request *http.Request) {
	userAgent := request.Header.Get("User-Agent")
	ipAddress := getIPAddress(request)
	forwardedHost := request.Header.Get("X-Forwarded-Host")

	log.Printf("Received request from IP: %s, User-Agent: %s, Forwarded Host: %s",
		ipAddress, userAgent, forwardedHost)

	// Construct and send the response containing the User-Agent, IP address, and Forwarded Host
	responseMessage := "Received User-Agent: " + userAgent + "\nIP Address: " + ipAddress + "\nForwarded Host: " + forwardedHost
	if _, err := writer.Write([]byte(responseMessage)); err != nil {
		log.Printf("Error writing response: %s", err)
		http.Error(writer, "Failed to write response", http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
}

// getIPAddress extracts the client's IP address from the request headers or RemoteAddr.
//
// Parameters:
// - request *http.Request: The HTTP request object containing all details of the request.
//
// Returns:
// - string: The IP address of the client.
func getIPAddress(request *http.Request) string {
	// Check the X-Forwarded-For header first
	forwarded := request.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// The X-Forwarded-For header can contain multiple IP addresses, the first one is the client's IP
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// If X-Forwarded-For is not present, fall back to RemoteAddr
	ipAddress := request.RemoteAddr
	// Remove the port number from the RemoteAddr
	if colonIndex := strings.LastIndex(ipAddress, ":"); colonIndex != -1 {
		ipAddress = ipAddress[:colonIndex]
	}
	return ipAddress
}

// startServer initializes and starts an HTTP server listening on the specified address.
// It registers handleRequests as the handler for the root path.
//
// Parameters:
// - address string: The network address that the server should listen on.
//
// Returns:
// - *http.Server: A pointer to the initialized server.
func startServer(address string) *http.Server {
	http.HandleFunc("/", handleRequests)

	server := &http.Server{Addr: address}

	go func() {
		log.Println("Starting server on", address)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Could not start server: %v", err)
		}
	}()

	return server
}

// shutdownServer gracefully shuts down the provided HTTP server within the specified timeout.
// It listens for system interrupt or termination signals to trigger a shutdown.
//
// Parameters:
// - server *http.Server: The server to shut down.
// - timeout time.Duration: The maximum duration to wait for a graceful shutdown.
func shutdownServer(server *http.Server, timeout time.Duration) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Block until a shutdown signal is received
	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Could not gracefully shutdown the server: %v", err)
	}

	log.Println("Server gracefully stopped")
}

func main() {
	server := startServer(":8888")
	shutdownServer(server, 10*time.Second)
}
